package onboarding

import (
	"fmt"
	"log"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"gitlab.com/nunet/device-management-service/models"
)

func Check_gpu() ([]models.Gpu, error) {
	var gpu_info []models.Gpu

	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		zlog.Sugar().Fatalf("Unable to Setup NVML")
		return nil, fmt.Errorf("Unable to Detect GPU")
	}
	defer func() {
		ret := nvml.Shutdown()
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to shutdown NVML: %v", nvml.ErrorString(ret))
		}
	}()
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		zlog.Sugar().Fatalf("Unable to get device count: %v", nvml.ErrorString(ret))
		return nil, fmt.Errorf("Unable to Detect GPU")
	}

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get device at index %d: %v", i, nvml.ErrorString(ret))
		}
		name, _ := device.GetName()
		memory, _ := nvml.DeviceGetMemoryInfo(device)
		var gpu models.Gpu
		gpu.Name = name
		gpu.TotVram = int(memory.Total / 1024 / 1024)
		gpu.FreeVram = int(memory.Free / 1024 / 1024)

		gpu_info = append(gpu_info, gpu)

	}
	return gpu_info, nil
}
