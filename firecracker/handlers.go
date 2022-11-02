// Package firecracker deals with anything related to Firecracker virtual machines. This involves creating, deleting,
// It also deals with keeping track of network interfaces, socket files.
package firecracker

import (
	"bytes"
	"context"
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
	"gitlab.com/nunet/device-management-service/firecracker/networking"
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

// StartCustom godoc
// @Summary		Start a VM with custom configuration.
// @Description	This endpoint is an abstraction of all primitive endpoints. When invokend, it calls all primitive endpoints in a sequence.
// @Tags		vm
// @Produce 	json
// @Success		200
// @Router		/vm/start-custom [post]
func StartCustom(c *gin.Context) {
	type StartCustomBody struct {
		KernelImagePath string `json:"kernel_image_path"`
		FilesystemPath  string `json:"filesystem_path"`
		VCPUCount       int    `json:"vcpu_count"`
		MemSizeMib      int    `json:"mem_size_mib"`
		TapDevice       string `json:"tap_device"`
	}

	body := StartCustomBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	tapDevName := networking.NextTapDevice()

	vm := models.VirtualMachine{
		SocketFile: GenerateSocketFile(10),
		BootSource: body.KernelImagePath,
		Filesystem: body.FilesystemPath,
		VCPUCount:  body.VCPUCount,
		MemSizeMib: body.MemSizeMib,
		TapDevice:  body.TapDevice,
		State:      "awaiting",
	}

	result := db.DB.Create(&vm)
	if result.Error != nil {
		panic(result.Error)
	}

	InitVM(c, vm)

	// PUT /boot-source
	bootSourceBody := models.BootSource{}
	bootSourceBody.KernelImagePath = body.KernelImagePath
	bootSourceBody.BootArgs = "console=ttyS0 reboot=k panic=1 pci=off"

	BootSource(c, vm, bootSourceBody)

	// PUT /drives
	drivesBody := models.Drives{}

	drivesBody.DriveID = "rootfs"
	drivesBody.PathOnHost = body.FilesystemPath
	drivesBody.IsRootDevice = true
	drivesBody.IsReadOnly = false

	Drives(c, vm, drivesBody)

	// PUT /machine-config
	machineConfigBody := models.MachineConfig{}
	// TODO: vCPU and memory has to be estimated based on how much capacity is remaining in nunet quota
	machineConfigBody.MemSizeMib = vm.MemSizeMib
	machineConfigBody.VCPUCount = vm.VCPUCount

	MachineConfig(c, vm, machineConfigBody)

	// PUT /network-interfaces
	networkInterfacesBody := models.NetworkInterfaces{}
	networkInterfacesBody.IfaceID = "eth0"
	networkInterfacesBody.GuestMac = "AA:FC:00:00:00:01"
	networkInterfacesBody.HostDevName = tapDevName

	NetworkInterfaces(c, vm, networkInterfacesBody)

	StartVM(c, vm)
}

// StartDefault godoc
// @Summary		Start a VM with default configuration.
// @Description	Everything except kernel files and filesystem file will be set by DMS itself.
// @Tags		vm
// @Produce 	json
// @Success		200
// @Router		/vm/start-default [post]
func StartDefault(c *gin.Context) {
	type StartDefaultBody struct {
		KernelImagePath string `json:"kernel_image_path"`
		FilesystemPath  string `json:"filesystem_path"`
	}

	body := StartDefaultBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	tapDevName := networking.NextTapDevice()

	vm := models.VirtualMachine{
		SocketFile: GenerateSocketFile(10),
		BootSource: body.KernelImagePath,
		Filesystem: body.FilesystemPath,
		TapDevice:  tapDevName,
		State:      "awaiting",
	}

	result := db.DB.Create(&vm)
	if result.Error != nil {
		panic(result.Error)
	}

	InitVM(c, vm)

	// PUT /boot-source
	bootSourceBody := models.BootSource{}
	bootSourceBody.KernelImagePath = body.KernelImagePath
	bootSourceBody.BootArgs = "console=ttyS0 reboot=k panic=1 pci=off"

	BootSource(c, vm, bootSourceBody)

	// PUT /drives
	drivesBody := models.Drives{}

	drivesBody.DriveID = "rootfs"
	drivesBody.PathOnHost = body.FilesystemPath
	drivesBody.IsRootDevice = true
	drivesBody.IsReadOnly = false

	Drives(c, vm, drivesBody)

	// PUT /machine-config
	machineConfigBody := models.MachineConfig{}
	// TODO: vCPU and memory has to be estimated based on how much capacity is remaining in nunet quota
	machineConfigBody.MemSizeMib = 1024
	machineConfigBody.VCPUCount = 2
	vm.MemSizeMib = 1024
	vm.VCPUCount = 2
	result = db.DB.Save(&vm)

	if result.Error != nil {
		panic(result.Error)
	}

	MachineConfig(c, vm, machineConfigBody)

	// PUT /network-interfaces
	networkInterfacesBody := models.NetworkInterfaces{}
	networkInterfacesBody.IfaceID = "eth0"
	networkInterfacesBody.GuestMac = "AA:FC:00:00:00:01"
	networkInterfacesBody.HostDevName = tapDevName

	NetworkInterfaces(c, vm, networkInterfacesBody)

	StartVM(c, vm)
}

func runFromConfig(c *gin.Context, vm models.VirtualMachine) {
	// Check if socket file already exists
	if _, err := os.Stat(vm.SocketFile); err == nil {
		log.Println("socket file exists, removing...")
		os.Remove(vm.SocketFile)
		log.Println(vm.SocketFile, "removed")
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

	// PUT /boot-source
	bootSourceBody := models.BootSource{}
	bootSourceBody.KernelImagePath = vm.BootSource
	bootSourceBody.BootArgs = "console=ttyS0 reboot=k panic=1 pci=off"

	BootSource(c, vm, bootSourceBody)

	// PUT /drives
	drivesBody := models.Drives{}

	drivesBody.DriveID = "rootfs"
	drivesBody.PathOnHost = vm.Filesystem
	drivesBody.IsRootDevice = true
	drivesBody.IsReadOnly = false

	Drives(c, vm, drivesBody)

	// PUT /machine-config
	machineConfigBody := models.MachineConfig{}
	// TODO: vCPU and memory has to be estimated based on how much capacity is remaining in nunet quota
	machineConfigBody.MemSizeMib = 256
	machineConfigBody.VCPUCount = 2

	Drives(c, vm, drivesBody)

	// PUT /network-interfaces
	networkInterfacesBody := models.NetworkInterfaces{}
	networkInterfacesBody.IfaceID = "eth0"
	networkInterfacesBody.GuestMac = "AA:FC:00:00:00:01"
	networkInterfacesBody.HostDevName = vm.TapDevice

	NetworkInterfaces(c, vm, networkInterfacesBody)

	// POST /start

	StartVM(c, vm)
}
