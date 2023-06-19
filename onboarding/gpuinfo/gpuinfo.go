package gpuinfo

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type GPUVendor int

const (
	Unknown GPUVendor = iota
	NVIDIA
	AMD
)

type GPUInfo struct {
	GPUName      string
	TotalMemory  uint64
	UsedMemory   uint64
	FreeMemory   uint64
	Vendor       GPUVendor
}

func GetAMDGPUInfo() ([]GPUInfo, error) {
	cmd := exec.Command("rocm-smi", "--showid", "--showproductname", "--showmeminfo", "vis_vram")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %s", err)
	}

	outputStr := string(output)

	gpuName := regexp.MustCompile(`Card series:\s+([^\n]+)`)
	total := regexp.MustCompile(`Total Memory \(B\):\s+(\d+)`)
	used := regexp.MustCompile(`Total Used Memory \(B\):\s+(\d+)`)

	gpuNameMatches := gpuName.FindAllStringSubmatch(outputStr, -1)
	totalMatches := total.FindAllStringSubmatch(outputStr, -1)
	usedMatches := used.FindAllStringSubmatch(outputStr, -1)

	if len(gpuNameMatches) == len(totalMatches) && len(totalMatches) == len(usedMatches) {
		var gpuInfos []GPUInfo
		for i := range gpuNameMatches {
			gpuName := gpuNameMatches[i][1]
			totalMemoryBytes, err := strconv.ParseInt(totalMatches[i][1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse total memory: %s", err)
			}

			usedMemoryBytes, err := strconv.ParseInt(usedMatches[i][1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse used memory: %s", err)
			}

			totalMemoryMiB := totalMemoryBytes / 1024 / 1024
			usedMemoryMiB := usedMemoryBytes / 1024 / 1024
			freeMemoryMiB := totalMemoryMiB - usedMemoryMiB

			gpuInfo := GPUInfo{
				GPUName:     "AMD " + gpuName,
				TotalMemory: uint64(totalMemoryMiB),
				UsedMemory:  uint64(usedMemoryMiB),
				FreeMemory:  uint64(freeMemoryMiB),
				Vendor:      AMD,
			}

			gpuInfos = append(gpuInfos, gpuInfo)
		}

		return gpuInfos, nil
	}

	return nil, fmt.Errorf("failed to find GPU information or memory information in the output")
}

func GetNVIDIAGPUInfo() ([]GPUInfo, error) {
	// Initialize NVML
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("failed to initialize NVML: %s", nvml.ErrorString(ret))
	}
	defer nvml.Shutdown()

	// Get the number of GPU devices
	deviceCount, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("failed to get device count: %s", nvml.ErrorString(ret))
	}

	var gpuInfos []GPUInfo

	// Iterate over each device
	for i := uint32(0); i < uint32(deviceCount); i++ {
		// Get the device handle
		device, ret := nvml.DeviceGetHandleByIndex(int(i))
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("failed to get device handle for device %d: %s", i, nvml.ErrorString(ret))
		}

		// Get the device name
		name, ret := nvml.DeviceGetName(device)
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("failed to get name for device %d: %s", i, nvml.ErrorString(ret))
		}

		// Get the memory info
		memory, ret := nvml.DeviceGetMemoryInfo(device)
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("failed to get memory info for device %d: %s", i, nvml.ErrorString(ret))
		}

		gpuInfo := GPUInfo{
			GPUName:     name,
			TotalMemory: memory.Total / 1024 / 1024,
			UsedMemory:  memory.Used / 1024 / 1024,
			FreeMemory:  memory.Free / 1024 / 1024,
			Vendor:      NVIDIA,
		}

		gpuInfos = append(gpuInfos, gpuInfo)
	}

	return gpuInfos, nil
}

func (v GPUVendor) String() string {
	switch v {
	case NVIDIA:
		return "NVIDIA"
	case AMD:
		return "AMD"
	default:
		return "Unknown"
	}
}
