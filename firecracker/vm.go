package firecracker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/dms/config"
	telemetry "gitlab.com/nunet/device-management-service/dms/resources"
	"gitlab.com/nunet/device-management-service/firecracker/networking"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
)

// RunPreviouslyRunningVMs runs `runFromConfig` for every firecracker VM record found in
// local DB which are marked `running`.
func RunPreviouslyRunningVMs() error {
	var vms []models.VirtualMachine

	if result := db.DB.Where("state = ?", "running").Find(&vms); result.Error != nil {
		zlog.Panic(result.Error.Error())
	}

	c := gin.Context{}

	for _, vm := range vms {
		runFromConfig(&c, vm)
	}
	return nil
}

// GenerateSocketFile generates a path for socket file to be used for communication with firecracker.
func GenerateSocketFile(n int) string {
	prefix := fmt.Sprintf("%s/sockets/", config.GetConfig().General.MetadataPath)
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Seed(time.Now().Unix())

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return prefix + string(s) + ".socket"
}

func initVM(c *gin.Context, vm models.VirtualMachine) error {
	cmd := exec.Command("firecracker", "--api-sock", vm.SocketFile)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}
	// output, _ := cmd.CombinedOutput() // for debugging purpose

	cmd.Stdout = os.Stdout // for debugging purpose
	// cmd.Stderr = os.Stderr // for debugging purpose
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// process started with .Start() lives even after parent's death: https://stackoverflow.com/a/46755495/939986
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Failed to start cmd: %v", err)
	}

	// wait for a flash, vm.SocketFile needs some time to create on disk
	time.Sleep(100 * time.Millisecond)
	return nil
}

func bootSource(c *gin.Context, vm models.VirtualMachine, bs models.BootSource) error {
	jsonBytes, _ := json.Marshal(bs)
	client := NewClient(vm.SocketFile)

	err := utils.MakeRequest(c, client, "http://localhost/boot-source", jsonBytes, ERR_BOOTSOURCE_REQ)
	if err != nil {
		return err
	}
	return nil
}

func drives(c *gin.Context, vm models.VirtualMachine, d models.Drives) error {
	jsonBytes, _ := json.Marshal(d)

	client := NewClient(vm.SocketFile)

	err := utils.MakeRequest(c, client, "http://localhost/drives/rootfs", jsonBytes, ERR_DRIVES_REQ)
	if err != nil {
		return err
	}
	return nil

}

func machineConfig(c *gin.Context, vm models.VirtualMachine, mc models.MachineConfig) error {
	jsonBytes, _ := json.Marshal(mc)

	client := NewClient(vm.SocketFile)

	err := utils.MakeRequest(c, client, "http://localhost/machine-config", jsonBytes, ERR_MACHINE_CONFIG_REQ)
	if err != nil {
		return err
	}
	return nil
}

func networkInterfaces(c *gin.Context, vm models.VirtualMachine, ni models.NetworkInterfaces) error {
	err := networking.ConfigureTapByName(vm.TapDevice)
	if err != nil {
		return errors.New("error configuring network")
	}

	jsonBytes, _ := json.Marshal(ni)

	client := NewClient(vm.SocketFile)

	err = utils.MakeRequest(c, client, "http://localhost/network-interfaces/eth0", jsonBytes, ERR_MACHINE_CONFIG_REQ)
	if err != nil {
		return err
	}
	return nil
}

func setupMMDS(c *gin.Context, vm models.VirtualMachine, mmds models.MMDSConfig) error {

	jsonBytes, _ := json.Marshal(mmds)
	client := NewClient(vm.SocketFile)
	err := utils.MakeRequest(c, client, "http://localhost/mmds/config", jsonBytes, ERR_MMDS_CONFIG)
	if err != nil {
		return err
	}
	return nil
}

func passMMDSMsg(c *gin.Context, vm models.VirtualMachine, mmdsMsg models.MMDSMsg) error {

	jsonBytes, _ := json.Marshal(mmdsMsg)
	client := NewClient(vm.SocketFile)
	err := utils.MakeRequest(c, client, "http://localhost/mmds", jsonBytes, ERR_MMDS_MSG)
	if err != nil {
		return err
	}

	return nil
}

func startVM(c *gin.Context, vm models.VirtualMachine) error {
	var jsonBytes = []byte(`{"action_type": "InstanceStart"}`)

	var freeRes models.FreeResources

	if err := db.DB.WithContext(c.Request.Context()).Where("id = ?", 1).First(&freeRes).Error; err != nil {
		return fmt.Errorf("error retrieving free resources: %v", err)
	}

	// Check if we have enough free resources before running VM
	if (vm.MemSizeMib > freeRes.Ram) ||
		(vm.VCPUCount > freeRes.Vcpu) {
		return errors.New("not enough resources available to deploy vm")
	}

	// initialize http client
	client := NewClient(vm.SocketFile)

	utils.MakeRequest(c, client, "http://localhost/actions", jsonBytes, ERR_ACTIONS_REQ)

	vm.State = "running"

	db.DB.WithContext(c.Request.Context()).Save(&vm)

	err := telemetry.CalcFreeResAndUpdateDB()
	if err != nil {
		return fmt.Errorf("Error calculating and updating FreeResources: %v", err)
	}

	_, err = telemetry.GetFreeResources()
	if err != nil {
		return fmt.Errorf("Error getting freeResources: %v", err)
	}

	libp2p.UpdateKadDHT()
	return nil
}

func stopVM(c *gin.Context, vm models.VirtualMachine) error {
	var jsonBytes = []byte(`{"action_type": "SendCtrlAltDel"}`)

	// initialize http client
	client := NewClient(vm.SocketFile)

	utils.MakeRequest(c, client, "http://localhost/actions", jsonBytes, ERR_ACTIONS_REQ)

	vm.State = "stopped"

	db.DB.WithContext(c.Request.Context()).Save(&vm)

	err := telemetry.CalcFreeResAndUpdateDB()
	if err != nil {
		return fmt.Errorf("Error calculating and updating FreeResources: %v", err)
	}

	_, err = telemetry.GetFreeResources()
	if err != nil {
		return fmt.Errorf("Error getting freeResources: %v", err)
	}
	libp2p.UpdateKadDHT()
	return nil
}
