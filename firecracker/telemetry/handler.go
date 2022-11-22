package telemetry

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"gitlab.com/nunet/device-management-service/adapter"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
	"gorm.io/gorm"
)

func QueryRunningVMs(DB *gorm.DB) []models.VirtualMachine {
	var vm []models.VirtualMachine
	result := DB.Where("state = ?", "running").Find(&vm)
	if result.Error != nil {
		panic(result.Error)
	}
	return vm

}

func CalcUsedResources(vms []models.VirtualMachine) (int, int) {
	var totalVCPU, totalMemSizeMib, totalCPUMhz int
	for i := 0; i < len(vms); i++ {
		totalVCPU += vms[i].VCPUCount
		totalMemSizeMib += vms[i].MemSizeMib
	}
	cores, _ := cpu.Info()
	totalCPUMhz = totalVCPU * int(cores[0].Mhz)
	return totalCPUMhz, totalMemSizeMib
}

func CalcFreeResources() error {
	vms := QueryRunningVMs(db.DB)

	totalCPUUsed, totalMem := CalcUsedResources(vms)

	var availableRes models.AvailableResources
	if res := db.DB.Find(&availableRes); res.RowsAffected == 0 {
		return res.Error
	}
	cpuProvisioned, memProvisioned, cpuHz := availableRes.TotCpuHz, availableRes.Ram, availableRes.CpuHz

	var freeResource models.FreeResources
	freeResource.ID = 1
	freeResource.TotCpuHz = cpuProvisioned - totalCPUUsed
	freeResource.Ram = memProvisioned - totalMem
	freeResource.Vcpu = int((cpuProvisioned - int(totalCPUUsed)) / int(cpuHz))
	freeResource.PriceCpu = availableRes.PriceCpu
	freeResource.PriceRam = availableRes.PriceRam
	freeResource.PriceDisk = availableRes.PriceDisk
	// TODO: Calculate remaining disk space

	// Check if we have a previous entry in the table
	var freeresource models.FreeResources
	if res := db.DB.Find(&freeresource); res.RowsAffected == 0 {
		result := db.DB.Create(&freeResource)
		if result.Error != nil {
			return result.Error
		}
	} else {
		result := db.DB.Model(&models.FreeResources{}).Where("id = ?", 1).Select("*").Updates(freeResource)
		if result.Error != nil {
			return result.Error
		}
	}

	_, err := adapter.UpdateAvailableResoruces()
	if err != nil {
		return err
	}
	return nil
}

// CalcFreeResources godoc
// @Summary		Returns the amount of free resources available
// @Description	Checks and returns the amount of free resources available
// @Tags		telemetry
// @Produce 	json
// @Success		200
// @Router		/free [get]
func GetFreeResource(c *gin.Context) {
	err := CalcFreeResources()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	var freeResource models.FreeResources
	if res := db.DB.Find(&freeResource); res.RowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error})
		return
	}

	c.JSON(http.StatusOK, freeResource)

}
