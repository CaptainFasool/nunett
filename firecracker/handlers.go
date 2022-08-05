// Package firecracker deals with anything related to Firecracker virtual machines. This involves creating, deleting,
// It also deals with keeping track of network interfaces, socket files.
package firecracker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
)

func NewClient(sockFile string) *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sockFile)
			},
		},
	}

	return client
}

func MakeRequest(c *gin.Context, client *http.Client, uri string, body []byte, errMsg string) {
	// set the HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(body))

	if err != nil {
		c.JSON(400, gin.H{
			"message":   errMsg,
			"timestamp": time.Now(),
		})
		return
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	_, err = client.Do(req)
	if err != nil {
		// c.JSON(400, gin.H{
		// 	"message":   errMsg,
		// 	"timestamp": time.Now(),
		// })
		// return
		panic(err)
	}

}

// InitVM		godoc
// @Summary		Starts the VM booting process.
// @Description	Starts the firecracker server for the specific VM. Further configuration are required.
// @Tags		vm
// @Produce 	json
// @Success		200
// @Router		/init [post]
func InitVM(c *gin.Context) {
	var vm models.VirtualMachine
	if err := db.DB.Where("id = ?", c.Param("vmID")).First(&vm).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	// Check if socket file already exists
	if _, err := os.Stat(vm.SocketFile); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Socket file already exists"})
		return
	}

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

	// use below code for testing only, waiting will never end the HTTP request
	// if err := cmd.Wait(); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"message":   fmt.Sprintf("Cmd returned error: %v", err),
	// 		"timestamp": time.Now(),
	// 	})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{
		"message":   "VM initiated. Add boot-source, add filesystem, invoke start",
		"timestamp": time.Now(),
	})

}

// BootSource	godoc
// @Summary		Configures kernel for the VM.
// @Description	Configure kernel for the VM.
// @Tags		vm
// @Produce 	json
// @Success		200
// @Router		/boot-source [put]
func BootSource(c *gin.Context) {
	// var jsonBytes = []byte(`{"kernel_image_path":"/home/santosh/firecracker/vmlinux.bin", "boot_args": "console=ttyS0 reboot=k panic=1 pci=off"}`)
	var vm models.VirtualMachine
	if err := db.DB.Where("id = ?", c.Param("vmID")).First(&vm).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	body := models.BootSource{}

	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	jsonBytes, _ := json.Marshal(body)

	client := NewClient(vm.SocketFile)

	MakeRequest(c, client, "http://localhost/boot-source", jsonBytes, ERR_BOOTSOURCE_REQ)
}

// Drives		godoc
// @Summary		Configures filesystem for the VM.
// @Description	Configures filesystem for the VM.
// @Tags		vm
// @Produce 	json
// @Success		200
// @Router		/drives [put]
func Drives(c *gin.Context) {
	// var jsonBytes = []byte(`{"drive_id": "rootfs", "path_on_host":"/home/santosh/firecracker/bionic.rootfs.ext4", "is_root_device": true, "is_read_only": false}`)
	var vm models.VirtualMachine
	if err := db.DB.Where("id = ?", c.Param("vmID")).First(&vm).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	body := models.Drives{}

	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	jsonBytes, _ := json.Marshal(body)

	client := NewClient(vm.SocketFile)

	MakeRequest(c, client, "http://localhost/drives/rootfs", jsonBytes, ERR_DRIVES_REQ)

}

// MachineConfig godoc
// @Summary		Configures system spec for the VM.
// @Description	Configures system spec for the VM like CPU and Memory.
// @Tags		vm
// @Produce 	json
// @Success		200
// @Router		/machine-config [put]
func MachineConfig(c *gin.Context) {
	// var jsonBytes = []byte(`{"vcpu_count": 2,"mem_size_mib": 512}`)
	var vm models.VirtualMachine
	if err := db.DB.Where("id = ?", c.Param("vmID")).First(&vm).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	body := models.MachineConfig{}

	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	jsonBytes, _ := json.Marshal(body)

	client := NewClient(vm.SocketFile)

	MakeRequest(c, client, "http://localhost/machine-config", jsonBytes, ERR_MACHINE_CONFIG_REQ)

}

// NetworkInterfaces godoc
// @Summary		Configures network interface on the host.
// @Description	Configures network interface on the host.
// @Tags		vm
// @Produce 	json
// @Success		200
// @Router		/network-interface [put]
func NetworkInterfaces(c *gin.Context) {
	// var jsonBytes = []byte(`{ "iface_id": "eth0", "guest_mac": "AA:FC:00:00:00:01", "host_dev_name": "tap1" }`)
	var vm models.VirtualMachine
	if err := db.DB.Where("id = ?", c.Param("vmID")).First(&vm).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	body := models.NetworkInterfaces{}

	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	jsonBytes, _ := json.Marshal(body)

	client := NewClient(vm.SocketFile)

	MakeRequest(c, client, "http://localhost/network-interfaces/eth0", jsonBytes, ERR_MACHINE_CONFIG_REQ)
}

// StartVM godoc
// @Summary		Start the VM.
// @Description	Start the VM.
// @Tags		vm
// @Produce 	json
// @Success		200
// @Router		/start [post]
func StartVM(c *gin.Context) {
	var jsonBytes = []byte(`{"action_type": "InstanceStart"}`)
	var vm models.VirtualMachine

	if err := db.DB.Where("id = ?", c.Param("vmID")).First(&vm).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	// initialize http client
	client := NewClient(vm.SocketFile)

	MakeRequest(c, client, "http://localhost/actions", jsonBytes, ERR_ACTIONS_REQ)

	vm.State = "running"

	db.DB.Save(&vm)
}

// StopVM godoc
// @Summary		Stop the VM.
// @Description	Stop the VM.
// @Tags		vm
// @Produce 	json
// @Success		200
// @Router		/stop [post]
func StopVM(c *gin.Context) {
	var jsonBytes = []byte(`{"action_type": "SendCtrlAltDel"}`)

	var vm models.VirtualMachine
	if err := db.DB.Where("id = ?", c.Param("vmID")).First(&vm).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	// initialize http client
	client := NewClient(vm.SocketFile)

	MakeRequest(c, client, "http://localhost/actions", jsonBytes, ERR_ACTIONS_REQ)

	vm.State = "stopped"

	db.DB.Save(&vm)
}

// MakeInternalRequest is a helper method to make call to DMS's own API
func MakeInternalRequest(c *gin.Context, methodType, internalEndpoint string, body []byte) {
	req, err := http.NewRequest(methodType, DMS_BASE_URL+internalEndpoint, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	client := http.Client{}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	_, err = client.Do(req)
	if err != nil {
		panic(err)
		// c.JSON(400, gin.H{
		// 	"message":   fmt.Sprintf("Error making %s request to %s", methodType, internalEndpoint),
		// 	"timestamp": time.Now(),
		// })
		// return
	}
}

// StartDefault godoc
// @Summary		Start a VM with default configuration.
// @Description	This endpoint is an abstraction of all other endpoints. When invokend, it calls all other endpoints in a sequence.
// @Tags		vm
// @Produce 	json
// @Success		200
// @Router		/start-default [post]
func StartDefault(c *gin.Context) {
	// Everything except kernel files and filesystem file will be set by DMS itself.

	type StartDefaultBody struct {
		KernelImagePath string `json:"kernel_image_path"`
		FilesystemPath  string `json:"filesystem_path"`
	}

	body := StartDefaultBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	vm := models.VirtualMachine{
		SocketFile:       GenerateSocketFile(10),
		BootSource:       "/home/santosh/firecracker/vmlinux.bin",
		Filesystem:       "/home/santosh/firecracker/bionic.rootfs.ext4",
		NetworkInterface: "tap1",
		State:            "awaiting",
	}

	result := db.DB.Create(&vm)
	if result.Error != nil {
		panic(result.Error)
	}

	// POST /init
	MakeInternalRequest(c, "POST", fmt.Sprintf("/vm/init/%d", vm.ID), nil)

	// PUT /boot-source
	bootSourceBody := models.BootSource{}
	bootSourceBody.KernelImagePath = body.KernelImagePath
	bootSourceBody.BootArgs = "console=ttyS0 reboot=k panic=1 pci=off"

	jsonBytes, _ := json.Marshal(bootSourceBody)
	MakeInternalRequest(c, "PUT", fmt.Sprintf("/vm/boot-source/%d", vm.ID), jsonBytes)

	// PUT /drives
	drivesBody := models.Drives{}

	drivesBody.DriveID = "rootfs"
	drivesBody.PathOnHost = body.FilesystemPath
	drivesBody.IsRootDevice = true
	drivesBody.IsReadOnly = false

	jsonBytes, _ = json.Marshal(drivesBody)
	MakeInternalRequest(c, "PUT", fmt.Sprintf("/vm/drives/%d", vm.ID), jsonBytes)

	// PUT /machine-config
	machineConfigBody := models.MachineConfig{}
	// TODO: vCPU and memory has to be estimated based on how much capacity is remaining in nunet quota
	machineConfigBody.MemSizeMib = 256
	machineConfigBody.VCPUCount = 2

	jsonBytes, _ = json.Marshal(machineConfigBody)
	MakeInternalRequest(c, "PUT", fmt.Sprintf("/vm/machine-config/%d", vm.ID), jsonBytes)

	// PUT /network-interfaces
	// MakeInternalRequest(c, "PUT", "/vm/network-interfaces", jsonBytes)

	MakeInternalRequest(c, "PUT", fmt.Sprintf("/vm/start/%d", vm.ID), jsonBytes)
}

func RunFromConfig(c *gin.Context) {
	body := models.VirtualMachine{}
	if err := c.BindJSON(&body); err != nil {
		// panic(err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Check if socket file already exists
	if _, err := os.Stat(body.SocketFile); err == nil {
		log.Println(body.SocketFile)
		panic(err)
		// c.JSON(http.StatusBadRequest, gin.H{"error": "Socket file already exists"})
		// return
	}

	cmd := exec.Command("firecracker", "--api-sock", body.SocketFile)
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

	// PUT /boot-source
	bootSourceBody := models.BootSource{}
	bootSourceBody.KernelImagePath = body.BootSource
	bootSourceBody.BootArgs = "console=ttyS0 reboot=k panic=1 pci=off"

	jsonBytes, _ := json.Marshal(bootSourceBody)
	MakeInternalRequest(c, "PUT", fmt.Sprintf("/vm/boot-source/%d", body.ID), jsonBytes)

	// PUT /drives
	drivesBody := models.Drives{}

	drivesBody.DriveID = "rootfs"
	drivesBody.PathOnHost = body.Filesystem
	drivesBody.IsRootDevice = true
	drivesBody.IsReadOnly = false

	jsonBytes, _ = json.Marshal(drivesBody)
	MakeInternalRequest(c, "PUT", fmt.Sprintf("/vm/drives/%d", body.ID), jsonBytes)

	// PUT /machine-config
	machineConfigBody := models.MachineConfig{}
	// TODO: vCPU and memory has to be estimated based on how much capacity is remaining in nunet quota
	machineConfigBody.MemSizeMib = 256
	machineConfigBody.VCPUCount = 2

	jsonBytes, _ = json.Marshal(machineConfigBody)
	MakeInternalRequest(c, "PUT", fmt.Sprintf("/vm/machine-config/%d", body.ID), jsonBytes)

	// PUT /network-interfaces
	// MakeInternalRequest(c, "PUT", fmt.Sprintf("/vm/network-interfaces/%d", body.ID), jsonBytes)

	// POST /start

	MakeInternalRequest(c, "PUT", fmt.Sprintf("/vm/start/%d", body.ID), nil)
}
