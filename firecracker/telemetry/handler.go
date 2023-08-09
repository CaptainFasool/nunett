package telemetry

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

func GetFreeResources() (models.FreeResources, error) {
	var freeResource models.FreeResources
	if res := db.DB.Find(&freeResource); res.RowsAffected == 0 {
		return freeResource, res.Error
	}
	return freeResource, nil
}

func QueryRunningVMs(DB *gorm.DB) ([]models.VirtualMachine, error) {
	var vm []models.VirtualMachine
	result := DB.Where("state = ?", "running").Find(&vm)
	if result.Error != nil {
		return nil, fmt.Errorf("unable to query running vms - %v", result.Error)
	}
	return vm, nil

}

func QueryRunningConts(DB *gorm.DB) ([]models.Services, error) {
	var services []models.Services
	result := DB.Where("job_status = ?", "running").Find(&services)
	if result.Error != nil {
		return nil, fmt.Errorf("unable to query running containers - %v", result.Error)
	}

	return services, nil
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

func CalcUsedResourcesConts(services []models.Services) (int, int, error) {
	if len(services) == 0 {
		return 0, 0, nil
	}
	var tot_cpu, tot_mem int
	for i := 0; i < len(services); i++ {
		idx := services[i].ResourceRequirements
		var resourceReq models.ServiceResourceRequirements
		result := db.DB.Where("id = ?", idx).Find(&resourceReq)
		if result.Error != nil {
			return 0, 0, fmt.Errorf("unable to query resource requirements - %v", result.Error)
		}
		tot_cpu += resourceReq.CPU
		tot_mem += resourceReq.RAM
	}

	return tot_cpu, tot_mem, nil
}

// queryRunningPlugins query for running plugins on DB
func queryRunningPlugins(DB *gorm.DB) ([]models.PluginInfo, error) {
	var plugins []models.PluginInfo
	if err := DB.Find(&plugins).Error; err != nil {
		return nil, err
	}
	return plugins, nil
}

// calcUsedResourcesPlugins iterate through Plugins and adds the resources usage for each
func calcUsedResourcesPlugins(plugins []models.PluginInfo) (int, int, float64) {
	usageCPU := 0
	usageRAM := 0
	usageDisk := 0.0
	for _, p := range plugins {
		usageCPU += p.ResourcesUsage.TotCpuHz
		usageRAM += p.ResourcesUsage.Ram
		usageDisk += p.ResourcesUsage.Disk
	}
	return usageCPU, usageRAM, usageDisk
}

func CalcFreeResources() error {
	vms, err := QueryRunningVMs(db.DB)
	if err != nil {
		return err
	}
	conts, err := QueryRunningConts(db.DB)
	if err != nil {
		return err
	}

	plugins, err := queryRunningPlugins(db.DB)
	if err != nil {
		return err
	}

	usagePluginsCPU, usagePluginsRAM, _ := calcUsedResourcesPlugins(plugins)
	tot_cpu_vm, tot_mem_vm := CalcUsedResourcesVMs(vms)
	tot_cpu_cont, tot_mem_cont, err := CalcUsedResourcesConts(conts)
	if err != nil {
		return err
	}

	tot_cpu_used := tot_cpu_cont + tot_cpu_vm + usagePluginsCPU
	tot_mem := tot_mem_cont + tot_mem_vm + usagePluginsRAM

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

func DeleteCalcFreeResources() error {
	var freeResource models.FreeResources
	result := db.DB.Where("id = ?", 1).Delete(&freeResource)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// CalcFreeResources godoc
//
//	@Summary		Returns the amount of free resources available
//	@Description	Checks and returns the amount of free resources available
//	@Tags			telemetry
//	@Produce		json
//	@Success		200
//	@Router			/telemetry/free [get]
func GetFreeResource(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/telemetry/free"))

	var freeResource models.FreeResources
	if res := db.DB.WithContext(c.Request.Context()).Find(&freeResource); res.RowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error})
		return
	}

	c.JSON(http.StatusOK, freeResource)
}
