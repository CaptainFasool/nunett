package firecracker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/firecracker/networking"
	"gitlab.com/nunet/device-management-service/firecracker/telemetry"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
)

func RunPreviouslyRunningVMs() error {
	var vms []models.VirtualMachine

	if result := db.DB.Where("state = ?", "running").Find(&vms); result.Error != nil {
		panic(result.Error)
	}

	c := gin.Context{}

	for _, vm := range vms {
		runFromConfig(&c, vm)
	}
	return nil
}

func GenerateSocketFile(n int) string {
	prefix := "/etc/nunet/sockets/"
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Seed(time.Now().Unix())

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return prefix + string(s) + ".socket"
}

func initVM(c *gin.Context, vm models.VirtualMachine) {
	cmd := exec.Command("firecracker", "--api-sock", vm.SocketFile)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}
	// output, _ := cmd.CombinedOutput() // for debugging purpose

	cmd.Stdout = os.Stdout // for debugging purpose
	// cmd.Stderr = os.Stderr // for debugging purpose
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// process started with .Start() lives even after parent's death: https://stackoverflow.com/a/46755495/939986
	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   fmt.Sprintf("Failed to start cmd: %v", stderr.String()),
			"timestamp": time.Now(),
		})
		return
	}

	// wait for a flash, vm.SocketFile needs some time to create on disk
	time.Sleep(100 * time.Millisecond)
}

func bootSource(c *gin.Context, vm models.VirtualMachine, bs models.BootSource) {
	jsonBytes, _ := json.Marshal(bs)
	client := NewClient(vm.SocketFile)

	utils.MakeRequest(c, client, "http://localhost/boot-source", jsonBytes, ERR_BOOTSOURCE_REQ)
}

func drives(c *gin.Context, vm models.VirtualMachine, d models.Drives) {
	jsonBytes, _ := json.Marshal(d)

	client := NewClient(vm.SocketFile)

	utils.MakeRequest(c, client, "http://localhost/drives/rootfs", jsonBytes, ERR_DRIVES_REQ)

}

func machineConfig(c *gin.Context, vm models.VirtualMachine, mc models.MachineConfig) {
	jsonBytes, _ := json.Marshal(mc)

	client := NewClient(vm.SocketFile)

	utils.MakeRequest(c, client, "http://localhost/machine-config", jsonBytes, ERR_MACHINE_CONFIG_REQ)
}

func networkInterfaces(c *gin.Context, vm models.VirtualMachine, ni models.NetworkInterfaces) {
	err := networking.ConfigureTapByName(vm.TapDevice)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("error configuring network"))
		return
	}

	jsonBytes, _ := json.Marshal(ni)

	client := NewClient(vm.SocketFile)

	utils.MakeRequest(c, client, "http://localhost/network-interfaces/eth0", jsonBytes, ERR_MACHINE_CONFIG_REQ)
}

func setupMMDS(c *gin.Context, vm models.VirtualMachine, mmds models.MMDSConfig) {

	jsonBytes, _ := json.Marshal(mmds)
	client := NewClient(vm.SocketFile)
	utils.MakeRequest(c, client, "http://localhost/mmds/config", jsonBytes, ERR_MMDS_CONFIG)
}

func passMMDSMsg(c *gin.Context, vm models.VirtualMachine, mmdsMsg models.MMDSMsg) {

	jsonBytes, _ := json.Marshal(mmdsMsg)
	client := NewClient(vm.SocketFile)
	utils.MakeRequest(c, client, "http://localhost/mmds", jsonBytes, ERR_MMDS_MSG)
}

func startVM(c *gin.Context, vm models.VirtualMachine) {
	var jsonBytes = []byte(`{"action_type": "InstanceStart"}`)

	var freeRes models.FreeResources

	if err := db.DB.Where("id = ?", 1).First(&freeRes).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	// Check if we have enough free resources before running VM
	if (vm.MemSizeMib > freeRes.Ram) ||
		(vm.VCPUCount > freeRes.Vcpu) {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "not enough resources available to deploy vm"})
		return
	}

	// initialize http client
	client := NewClient(vm.SocketFile)

	utils.MakeRequest(c, client, "http://localhost/actions", jsonBytes, ERR_ACTIONS_REQ)

	vm.State = "running"

	db.DB.Save(&vm)

	telemetry.CalcFreeResources()
	libp2p.UpdateDHT()

	c.JSON(http.StatusOK, gin.H{
		"message":   "VM started successfully.",
		"timestamp": time.Now(),
	})
}

func stopVM(c *gin.Context, vm models.VirtualMachine) {
	var jsonBytes = []byte(`{"action_type": "SendCtrlAltDel"}`)

	// initialize http client
	client := NewClient(vm.SocketFile)

	utils.MakeRequest(c, client, "http://localhost/actions", jsonBytes, ERR_ACTIONS_REQ)

	vm.State = "stopped"

	db.DB.Save(&vm)

	telemetry.CalcFreeResources()
	libp2p.UpdateDHT()
}
