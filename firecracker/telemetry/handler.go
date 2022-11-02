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
	var tot_vcpu, tot_mem_size_mib, tot_cpu_mhz int
	for i := 0; i < len(vms); i++ {
		tot_vcpu += vms[i].VCPUCount
		tot_mem_size_mib += vms[i].MemSizeMib
	}
	cores, _ := cpu.Info()
	tot_cpu_mhz = tot_vcpu * int(cores[0].Mhz)
	return tot_cpu_mhz, tot_mem_size_mib
}

func CalcFreeResources() error {
	vms := QueryRunningVMs(db.DB)

	tot_cpu_used, tot_mem := CalcUsedResources(vms)

	var avail_resources models.AvailableResources
	if res := db.DB.Find(&avail_resources); res.RowsAffected == 0 {
		return res.Error
	}
	cpu_provisioned, mem_provisioned, cpu_hz := avail_resources.TotCpuHz, avail_resources.Ram, avail_resources.CpuHz

	var freeResource models.FreeResources
	freeResource.ID = 1
	freeResource.TotCpuHz = cpu_provisioned - tot_cpu_used
	freeResource.Ram = mem_provisioned - tot_mem
	freeResource.Vcpu = int((cpu_provisioned - int(tot_cpu_used)) / int(cpu_hz))
	freeResource.PriceCpu = avail_resources.PriceCpu
	freeResource.PriceRam = avail_resources.PriceRam
	freeResource.PriceDisk = avail_resources.PriceDisk
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
