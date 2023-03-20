package onboarding

import (
	"fmt"

	"github.com/jaypipes/ghw"
	"gitlab.com/nunet/device-management-service/models"
)

func Check_gpu() ([]models.Gpu, error) {
	var gpu_info []models.Gpu
	gpu, err := ghw.GPU()
	if err != nil {
		// zlog.Sugar().Error("Error getting GPU info: %v", err)
		return nil, fmt.Errorf("unable to get GPU info")
	}

	if len(gpu.GraphicsCards) == 0 {
		var gpu models.Gpu
		gpu.Name = ""
		gpu.FreeVram = 0
		gpu.TotVram = 0
		return gpu_info, nil
	}
	for _, v := range gpu.GraphicsCards {
		if v.DeviceInfo.Driver == "nvidia" {
			var gpu models.Gpu
			gpu.Name = v.DeviceInfo.Product.Name
			gpu.FreeVram = 0
			gpu.TotVram = 0
			gpu_info = append(gpu_info, gpu)

		}

	}
	return gpu_info, nil

}
