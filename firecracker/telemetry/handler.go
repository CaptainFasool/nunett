package telemetry

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
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

func QueryRunningConts(DB *gorm.DB) []models.Services {
	var services []models.Services
	result := DB.Find(&services)
	if result.Error != nil {
		panic(result.Error)
	}

	return services
}

func CalcUsedResourcesVMs(vms []models.VirtualMachine) (int, int) {
	if len(vms) == 0 {
		return 0, 0
	}
	var tot_VCPU, totalMemSizeMib, totalCPUMhz int
	for i := 0; i < len(vms); i++ {
		tot_VCPU += vms[i].VCPUCount
		totalMemSizeMib += vms[i].MemSizeMib
	}
	cores, _ := cpu.Info()
	totalCPUMhz = tot_VCPU * int(cores[0].Mhz)
	return totalCPUMhz, totalMemSizeMib
}

func CalcUsedResourcesConts(services []models.Services) (int, int) {
	if len(services) == 0 {
		return 0, 0
	}
	var tot_cpu, tot_mem int
	for i := 0; i < len(services); i++ {
		idx := services[i].ResourceRequirements
		var resourceReq models.ServiceResourceRequirements
		result := db.DB.Where("id = ?", idx).Find(&resourceReq)
		if result.Error != nil {
			panic(result.Error)
		}
		tot_cpu += resourceReq.CPU
		tot_mem += resourceReq.RAM
	}

	return tot_cpu, tot_mem
}

func CalcFreeResources() error {
	vms := QueryRunningVMs(db.DB)
	conts := QueryRunningConts(db.DB)

	tot_cpu_vm, tot_mem_vm := CalcUsedResourcesVMs(vms)
	tot_cpu_cont, tot_mem_cont := CalcUsedResourcesConts(conts)

	tot_cpu_used := tot_cpu_cont + tot_cpu_vm
	tot_mem := tot_mem_cont + tot_mem_vm

	var availableRes models.AvailableResources
	if res := db.DB.Find(&availableRes); res.RowsAffected == 0 {
		return res.Error
	}
	cpuProvisioned, memProvisioned, cpuHz := availableRes.TotCpuHz, availableRes.Ram, availableRes.CpuHz

	var freeResource models.FreeResources
	freeResource.ID = 1
	freeResource.TotCpuHz = cpuProvisioned - tot_cpu_used
	freeResource.Ram = memProvisioned - tot_mem
	freeResource.Vcpu = int((cpuProvisioned - int(tot_cpu_used)) / int(cpuHz))
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
