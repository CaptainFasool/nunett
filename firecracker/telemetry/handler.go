package telemetry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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

// CalcFreeResources godoc
// @Summary		Calculates free resources
// @Description	Calculates total CPU(Mhz) and Mem available for use
// @Tags		telemetry
// @Produce 	json
// @Success		200
//@Router		/free [get]
func GetFreeResource(c *gin.Context) {
	vms := QueryRunningVMs(db.DB)

	tot_cpu_mhz, tot_mem := CalcUsedResources(vms)

	_, err := os.Stat("/etc/nunet/metadataV2.json")
	if os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "/etc/nunet/metadataV2.json does not exist. Is nunet onboarded successfully?"})
		return
	}
	metadata := GetMetadata()
	cpu_provisioned, mem_provisioned := metadata.Reserved.Cpu, metadata.Reserved.Memory

	var freeResource models.FreeResources
	fmt.Printf("Provisioned cpu %d", cpu_provisioned)
	fmt.Printf("Provisioned mem %d", mem_provisioned)
	fmt.Printf("Total mem %d", tot_mem)
	freeResource.CPU = float64(cpu_provisioned) - float64(tot_cpu_mhz)
	freeResource.Memory = int64(mem_provisioned) - int64(tot_mem)
	fmt.Printf("USED --- %d", tot_cpu_mhz)
	c.JSON(http.StatusOK, freeResource)

}
