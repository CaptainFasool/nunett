package onboarding

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"gitlab.com/nunet/device-management-service/models"
)

// totalRamInMB fetches total memory installed on host machine
func totalRamInMB() uint64 {
	v, _ := mem.VirtualMemory()

	ramInMB := v.Total / 1024 / 1024

	return ramInMB
}

// totalCPUInMHz fetches compute capacity of the host machine
func totalCPUInMHz() float64 {
	cores, _ := cpu.Info()

	var totalCompute float64

	for i := 0; i < len(cores); i++ {
		totalCompute += cores[i].Mhz
	}

	return totalCompute
}

// GetTotalProvisioned returns Provisioned struct with provisioned memory and CPU.
func GetTotalProvisioned() *models.Provisioned {
	provisioned := &models.Provisioned{
		CPU:    totalCPUInMHz(),
		Memory: totalRamInMB(),
	}
	return provisioned
}
