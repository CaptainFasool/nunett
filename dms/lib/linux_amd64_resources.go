//go:build linux && amd64

package library

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

// fetches the max clock speed of a single core
// Assuming all cores have the same clock speed
func Hz_per_cpu() float64 {
	cores, _ := cpu.Info()

	return cores[0].Mhz
}

// GetTotalProvisioned returns Provisioned struct with provisioned memory and CPU.
func GetTotalProvisioned() *models.Provisioned {
	cores, _ := cpu.Info()

	provisioned := &models.Provisioned{
		CPU:      totalCPUInMHz(),
		Memory:   totalRamInMB(),
		NumCores: uint64(len(cores)),
	}
	return provisioned
}
