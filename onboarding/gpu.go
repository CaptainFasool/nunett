package onboarding

import (
	"fmt"
        "strconv"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding/gpudetect"
	"gitlab.com/nunet/device-management-service/onboarding/gpuinfo"
)

func Check_gpu() ([]models.Gpu, error) {
        var gpu_info []models.Gpu
        vendors, err := gpudetect.DetectGPUVendors()
        if err != nil {
                return nil, fmt.Errorf("unable to detect GPU Vendor: %v", err)
        }
        foundNVIDIA, foundAMD := false, false
        for _, vendor := range vendors {
                switch vendor {
                case gpudetect.NVIDIA:
                        if !foundNVIDIA {
                                var gpu models.Gpu
                                info, err := gpuinfo.GetNVIDIAGPUInfo()
                                if err != nil {
                                        return nil, fmt.Errorf("error getting NVIDIA GPU info: %v", err)
                                }
                                for _, i := range info {
                                        gpu.Name = i.GPUName
                                        gpu.FreeVram = i.FreeMemory
                                        gpu.TotVram = i.TotalMemory
                                        gpu_info = append(gpu_info, gpu)
                                }
                                foundNVIDIA = true
                        }
                case gpudetect.AMD:
                        if !foundAMD {
                                var gpu models.Gpu
                                info, err := gpuinfo.GetAMDGPUInfo()
                                if err != nil {
                                        return nil, fmt.Errorf("error getting AMD GPU info: %v", err)
                                }
                                for _, i := range info {
                                        gpu.Name = i.GPUName
                                        gpu.FreeVram = i.FreeMemory
                                        gpu.TotVram = i.TotalMemory
                                        gpu_info = append(gpu_info, gpu)
                                }
                                foundAMD = true
                        }
                case gpudetect.Unknown:
                        fmt.Println("Unknown GPU(s) detected")
                }
        }
        return gpu_info, nil
}
