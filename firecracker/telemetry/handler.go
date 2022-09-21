package telemetry

import (
	"encoding/json"
	"fmt"
	"io"
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

// CalcUsedResources godoc
// @Summary		Calculates Resources used by VMs
// @Description	Calculates total CPU(Mhz) and Mem used by running VMs
// @Tags		vm
// @Produce 	json
// @Success		200
// TODO: update route @Router		/machine-config/:vmID [put]  ************************TODO*****************
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

func GetMetadata() models.Metadata {
	resp, err := http.Get("http://localhost:9999/api/v1/onboarding/metadata")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var metadata models.Metadata

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		panic(err)
	}
	fmt.Println(metadata.Reserved)
	return metadata
}

func GetFreeResource(c *gin.Context) {
	vms := QueryRunningVMs(db.DB)

	tot_cpu_mhz, tot_mem := CalcUsedResources(vms)

	metadata := GetMetadata()
	cpu_provisioned, mem_provisioned := metadata.Reserved.Cpu, metadata.Reserved.Memory

	var freeResource models.FreeResources
	fmt.Printf("Provisioned cpu %d", cpu_provisioned)
	freeResource.CPU = float64(cpu_provisioned) - float64(tot_cpu_mhz)
	freeResource.Memory = uint64(mem_provisioned) - uint64(tot_mem)
	fmt.Printf("USED --- %d", tot_cpu_mhz)
	c.JSON(http.StatusOK, freeResource)

}
