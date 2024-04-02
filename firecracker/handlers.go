// Package firecracker deals with anything related to Firecracker virtual machines. This involves creating, deleting,
// It also deals with keeping track of network interfaces, socket files.
package firecracker

import (
	"bytes"
	"context"
	"fmt"
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
	"gitlab.com/nunet/device-management-service/utils"
)

type DefaultVM struct {
	KernelImagePath string `json:"kernel_image_path" binding:"required"`
	FilesystemPath  string `json:"filesystem_path" binding:"required"`
	PublicKey       string `json:"public_key"`
	NodeID          string `json:"node_id"`
}

type CustomVM struct {
	KernelImagePath string `json:"kernel_image_path" binding:"required"`
	FilesystemPath  string `json:"filesystem_path" binding:"required"`
	VCPUCount       int    `json:"vcpu_count" binding:"required"`
	MemSizeMib      int    `json:"mem_size_mib" binding:"required"`
	TapDevice       string `json:"tap_device"`
}

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

func StartCustom(ctx context.Context, custom CustomVM) error {
	tapDevName := networking.NextTapDevice()

	vm := models.VirtualMachine{
		SocketFile: GenerateSocketFile(10),
		BootSource: custom.KernelImagePath,
		Filesystem: custom.FilesystemPath,
		VCPUCount:  custom.VCPUCount,
		MemSizeMib: custom.MemSizeMib,
		TapDevice:  custom.TapDevice,
		State:      "awaiting",
	}

	res := db.DB.WithContext(ctx).Create(&vm)
	if res.Error != nil {
		return fmt.Errorf("could not create database table: %w", res.Error)
	}

	err := initVM(vm)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("could not initialize virtual machine: %w", err)
	}

	// PUT /boot-source
	bootSourceBody := models.BootSource{}
	bootSourceBody.KernelImagePath = custom.KernelImagePath
	bootSourceBody.BootArgs = "console=ttyS0 reboot=k panic=1 pci=off"

	err = bootSource(nil, vm, bootSourceBody)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("failed to configure boot source: %w", err)
	}

	// PUT /drives
	drivesBody := models.Drives{}

	drivesBody.DriveID = "rootfs"
	drivesBody.PathOnHost = custom.FilesystemPath
	drivesBody.IsRootDevice = true
	drivesBody.IsReadOnly = false

	err = drives(nil, vm, drivesBody)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("failed to configure drives: %w", err)
	}

	// PUT /machine-config
	machineConfigBody := models.MachineConfig{}
	// TODO: vCPU and memory has to be estimated based on how much capacity is remaining in nunet quota
	machineConfigBody.MemSizeMib = vm.MemSizeMib
	machineConfigBody.VCPUCount = vm.VCPUCount

	err = machineConfig(nil, vm, machineConfigBody)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("failed to configure machine config: %w", err)
	}

	// PUT /network-interfaces
	networkInterfacesBody := models.NetworkInterfaces{}
	networkInterfacesBody.IfaceID = "eth0"
	networkInterfacesBody.GuestMac = "AA:FC:00:00:00:01"
	networkInterfacesBody.HostDevName = tapDevName

	err = networkInterfaces(nil, vm, networkInterfacesBody)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("failed to configure network-interfaces: %w", err)
	}

	mmdsConfigBody := models.MMDSConfig{}
	mmdsConfigBody.NetworkInterface = append(mmdsConfigBody.NetworkInterface, networkInterfacesBody.IfaceID)

	err = setupMMDS(nil, vm, mmdsConfigBody)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("failed to setup MMDS: %w", err)
	}

	mmdsMsg := models.MMDSMsg{}
	mmdsMetadata := models.MMDSMetadata{}
	//TODO: Currently passing fake data will be replaced with information from Deployment Request
	mmdsMetadata.NodeId = utils.RandomString(15)
	mmdsMetadata.PKey = utils.RandomString(15)
	mmdsMsg.Latest.Metadata.MMDSMetadata = mmdsMetadata

	err = passMMDSMsg(nil, vm, mmdsMsg)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("failed to pass MMDS message: %w", err)
	}

	// POST /start
	err = startVM(nil, vm)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("unable to start virtual machine: %w", err)
	}
	return nil
}

func StartDefault(ctx context.Context, def DefaultVM) error {
	tapDevName := networking.NextTapDevice()

	vm := models.VirtualMachine{
		SocketFile: GenerateSocketFile(10),
		BootSource: def.KernelImagePath,
		Filesystem: def.FilesystemPath,
		TapDevice:  tapDevName,
		State:      "awaiting",
	}

	res := db.DB.WithContext(ctx).Create(&vm)
	if res.Error != nil {
		zlog.Panic(res.Error.Error())
	}

	err := initVM(vm)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("could not initialize virtual machine: %w", err)
	}

	// PUT /boot-source
	bootSourceBody := models.BootSource{}
	bootSourceBody.KernelImagePath = def.KernelImagePath
	bootSourceBody.BootArgs = "console=ttyS0 reboot=k panic=1 pci=off"

	err = bootSource(nil, vm, bootSourceBody)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("failed to configure boot source: %w", err)
	}

	// PUT /drives
	drivesBody := models.Drives{}

	drivesBody.DriveID = "rootfs"
	drivesBody.PathOnHost = def.FilesystemPath
	drivesBody.IsRootDevice = true
	drivesBody.IsReadOnly = false

	err = drives(nil, vm, drivesBody)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("failed to configure drives: %w", err)
	}

	// PUT /machine-config
	machineConfigBody := models.MachineConfig{}
	// TODO: vCPU and memory has to be estimated based on how much capacity is remaining in nunet quota
	machineConfigBody.MemSizeMib = 1024
	machineConfigBody.VCPUCount = 2
	vm.MemSizeMib = 1024
	vm.VCPUCount = 2
	res = db.DB.WithContext(ctx).Save(&vm)

	if res.Error != nil {
		zlog.Panic(res.Error.Error())
	}

	err = machineConfig(nil, vm, machineConfigBody)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("failed to configure machineConfig: %w", err)
	}

	// PUT /network-interfaces
	networkInterfacesBody := models.NetworkInterfaces{}
	networkInterfacesBody.IfaceID = "eth0"
	networkInterfacesBody.GuestMac = "AA:FC:00:00:00:01"
	networkInterfacesBody.HostDevName = tapDevName

	err = networkInterfaces(nil, vm, networkInterfacesBody)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("failed to configure network-interfaces: %w", err)
	}

	mmdsConfigBody := models.MMDSConfig{}
	mmdsConfigBody.NetworkInterface = append(mmdsConfigBody.NetworkInterface, networkInterfacesBody.IfaceID)

	err = setupMMDS(nil, vm, mmdsConfigBody)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("failed to setup MMDS: %w", err)
	}

	mmdsMsg := models.MMDSMsg{}
	mmdsMetadata := models.MMDSMetadata{}
	mmdsMetadata.NodeId = def.NodeID
	mmdsMetadata.PKey = def.PublicKey
	mmdsMsg.Latest.Metadata.MMDSMetadata = mmdsMetadata

	err = passMMDSMsg(nil, vm, mmdsMsg)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("failed to pass MMDS message: %w", err)
	}

	// POST /start
	err = startVM(nil, vm)
	if err != nil {
		zlog.Sugar().Errorln(err)
		return fmt.Errorf("unable to start virtual machine: %w", err)
	}
	return nil
}

func runFromConfig(c *gin.Context, vm models.VirtualMachine) {
	// Check if socket file already exists
	if _, err := os.Stat(vm.SocketFile); err == nil {
		zlog.Info("socket file exists, removing...")
		os.Remove(vm.SocketFile)
		zlog.Sugar().Infof(vm.SocketFile, "removed")
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
		zlog.Sugar().Errorf("Failed to start cmd: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     fmt.Sprintf("Failed to start cmd: %v", stderr.String()),
			"timestamp": time.Now().In(time.UTC),
		})
		return
	}

	// PUT /boot-source
	bootSourceBody := models.BootSource{}
	bootSourceBody.KernelImagePath = vm.BootSource
	bootSourceBody.BootArgs = "console=ttyS0 reboot=k panic=1 pci=off"

	if err := bootSource(c, vm, bootSourceBody); err != nil {
		zlog.Sugar().Errorln(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     fmt.Sprintf("Failed to configure boot source: %v", err.Error()),
			"timestamp": time.Now().In(time.UTC),
		})
		return
	}

	// PUT /drives
	drivesBody := models.Drives{}

	drivesBody.DriveID = "rootfs"
	drivesBody.PathOnHost = vm.Filesystem
	drivesBody.IsRootDevice = true
	drivesBody.IsReadOnly = false

	if err := drives(c, vm, drivesBody); err != nil {
		zlog.Sugar().Errorln(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     fmt.Sprintf("Failed to configure drives: %v", err.Error()),
			"timestamp": time.Now().In(time.UTC),
		})
		return
	}

	// PUT /machine-config
	machineConfigBody := models.MachineConfig{}
	// TODO: vCPU and memory has to be estimated based on how much capacity is remaining in nunet quota
	machineConfigBody.MemSizeMib = 256
	machineConfigBody.VCPUCount = 2

	if err := drives(c, vm, drivesBody); err != nil {
		zlog.Sugar().Errorln(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     fmt.Sprintf("Failed to configure drives: %v", err.Error()),
			"timestamp": time.Now().In(time.UTC),
		})
		return
	}

	// PUT /network-interfaces
	networkInterfacesBody := models.NetworkInterfaces{}
	networkInterfacesBody.IfaceID = "eth0"
	networkInterfacesBody.GuestMac = "AA:FC:00:00:00:01"
	networkInterfacesBody.HostDevName = vm.TapDevice

	if err := networkInterfaces(c, vm, networkInterfacesBody); err != nil {
		zlog.Sugar().Errorln(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     fmt.Sprintf("Failed to configure network-interfaces: %v", err.Error()),
			"timestamp": time.Now().In(time.UTC),
		})
		return
	}

	mmdsConfigBody := models.MMDSConfig{}
	mmdsConfigBody.NetworkInterface = append(mmdsConfigBody.NetworkInterface, networkInterfacesBody.IfaceID)

	if err := setupMMDS(c, vm, mmdsConfigBody); err != nil {
		zlog.Sugar().Errorln(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     fmt.Sprintf("Failed to setup MMDS: %v", err.Error()),
			"timestamp": time.Now().In(time.UTC),
		})
		return
	}

	mmdsMsg := models.MMDSMsg{}
	mmdsMetadata := models.MMDSMetadata{}
	//TODO: Currently passing fake data will be replaced with information from Deployment Request
	mmdsMetadata.NodeId = "12343124-3423425234-23423534234"
	mmdsMetadata.PKey = "3usf3/3gf/23r sdf3r2rdfsdfa"
	mmdsMsg.Latest.Metadata.MMDSMetadata = mmdsMetadata

	if err := passMMDSMsg(c, vm, mmdsMsg); err != nil {
		zlog.Sugar().Errorln(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     fmt.Sprintf("Failed to pass MMDS message: %v", err.Error()),
			"timestamp": time.Now().In(time.UTC),
		})
		return
	}

	// POST /start
	if err := startVM(c, vm); err != nil {
		zlog.Sugar().Errorln(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     err.Error(),
			"timestamp": time.Now().In(time.UTC),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "VM started successfully.",
		"timestamp": time.Now().In(time.UTC),
	})
}
