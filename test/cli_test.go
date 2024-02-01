package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"os"
	"os/exec"

	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	library "gitlab.com/nunet/device-management-service/dms/lib"
	"gitlab.com/nunet/device-management-service/dms/onboarding"
)

func SetupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")

	onboardingRoute := v1.Group("/onboarding")
	{
		onboardingRoute.GET("/metadata", onboarding.GetMetadata)
		onboardingRoute.POST("/onboard", onboarding.Onboard)
		onboardingRoute.GET("/provisioned", onboarding.ProvisionedCapacity)
		onboardingRoute.GET("/address/new", onboarding.CreatePaymentAddress)
	}

	return router
}

type CLI struct {
	suite.Suite
	sync.WaitGroup
	cleanup func(context.Context) error
}

func TestCLI(t *testing.T) {
	s := new(CLI)
	suite.Run(t, s)
}

func (s *CLI) SetupSuite() {
	os.Mkdir("/tmp/nunet.test", 0755)
	config.LoadConfig()
	config.SetConfig("general.metadata_path", "/tmp/nunet.test/")

	mockDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		s.T().Errorf("error trying to initialize mock db: %v", err)
	}
	db.DB = mockDB

	s.cleanup = tracing.InitTracer()

	s.WaitGroup.Add(1)

	router := SetupRouter()
	go router.Run(":9999")
	time.Sleep(4 * time.Second)
}

func (s *CLI) TearDownSuite() {
	s.WaitGroup.Done()
	s.cleanup(context.Background())
	os.RemoveAll("/tmp/nunet.test")
}

func (s *CLI) TestNunetWalletNewEthereumCLI() {
	out, err := exec.Command("../maint-scripts/nunet-dms/usr/bin/nunet",
		"wallet", "new", "-e").Output()

	if err != nil {
		s.T().Logf("Error executing command: %v", err)
		s.FailNow("Failed to execute command")
	}

	s.T().Logf("Output: %s", string(out))

	s.Nil(err)

	type WalletNewEthereumOutput struct {
		Address    string `json:"address"`
		PrivateKey string `json:"private_key"`
	}
	var resp WalletNewEthereumOutput
	err = json.Unmarshal(out, &resp)

	if err != nil {
		s.T().Logf("Error unmarshalling JSON: %v", err)
		s.FailNow("Failed to unmarshal JSON")
	}

	s.Nil(err)
}

func (s *CLI) TestNunetWalletNewCardanoCLI() {
	out, err := exec.Command("../maint-scripts/nunet-dms/usr/bin/nunet",
		"wallet", "new", "--cardano").Output()

	s.Nil(err)

	type WalletNewCardanoOutput struct {
		Address  string `json:"address"`
		Mnemonic string `json:"mnemonic"`
	}
	var resp WalletNewCardanoOutput
	err = json.Unmarshal(out, &resp)

	s.Nil(err)
}

func (s *CLI) TestNuNetAvailableCLI() {
	out, err := exec.Command("../maint-scripts/nunet-dms/usr/bin/nunet",
		"capacity", "-v").Output()

	s.Nil(err)

	outLines := strings.Split(string(out), "\n")

	s.Equal("Total Machine Capacity", outLines[0])

	type TotalMachineCapacityOutput struct {
		CPU        float64 `json:"cpu"`
		Memory     float64 `json:"memory"`
		TotalCores int     `json:"total_cores"`
	}
	var totalResp TotalMachineCapacityOutput
	err = json.Unmarshal([]byte(outLines[1]), &totalResp)
	s.Nil(err)

	s.Equal("Reserved for Nunet", outLines[3])
	type ReservedForNuNetOutput struct {
		CPU    float64 `json:"cpu"`
		Memory float64 `json:"memory"`
	}

	var reservedResp ReservedForNuNetOutput
	err = json.Unmarshal([]byte(outLines[4]), &reservedResp)
	s.Nil(err)

	s.Equal("Available Machine Capacity", outLines[6])
	type AvailableOutput struct {
		CPU    float64 `json:"cpu"`
		Memory float64 `json:"memory"`
	}
	var availableResp AvailableOutput
	err = json.Unmarshal([]byte(outLines[7]), &availableResp)
	s.Nil(err)
}

func (s *CLI) TestNunetOnboardNoCPUValueCLI() {
	availableMemory := onboarding.GetTotalProvisioned().Memory
	halfMemory := availableMemory / 2
	out, _ := exec.Command("../maint-scripts/nunet-dms/usr/bin/nunet",
		"onboard",
		"-m", fmt.Sprint(halfMemory),
		"-a", "testaddress",
		"-n", "nunet-team").Output()
	s.Contains(string(out), "Error: -c | --cpu must be specified")
}

func (s *CLI) TestNunetOnboardNoMemoryValueCLI() {
	availableCPU := onboarding.GetTotalProvisioned().CPU
	halfCPU := availableCPU / 2
	out, _ := exec.Command("../maint-scripts/nunet-dms/usr/bin/nunet",
		"onboard",
		"-c", fmt.Sprint(halfCPU),
		"-a", "testaddress",
		"-n", "nunet-team").Output()
	s.Contains(string(out), "Error: -m | --memory must be specified")
}

func (s *CLI) TestNunetOnboardNoChannelCLI() {
	provisioned := onboarding.GetTotalProvisioned()
	availableMemory, availableCPU := provisioned.Memory, provisioned.CPU
	fiftyPercentMemory := availableMemory / 2
	fiftyPercentCPU := availableCPU / 2
	out, _ := exec.Command("../maint-scripts/nunet-dms/usr/bin/nunet",
		"onboard",
		"-c", fmt.Sprint(fiftyPercentCPU),
		"-m", fmt.Sprint(fiftyPercentMemory),
		"-a", "testaddress").Output()
	s.Contains(string(out), "Error: -n | --nunet-channel must be specified")
}

func (s *CLI) TestNunetOnboardNoAddressCLI() {
	provisioned := onboarding.GetTotalProvisioned()
	availableMemory, availableCPU := provisioned.Memory, provisioned.CPU
	fiftyPercentMemory := availableMemory / 2
	fiftyPercentCPU := uint64(availableCPU / 2)
	out, err := exec.Command("../maint-scripts/nunet-dms/usr/bin/nunet",
		"onboard",
		"-c", fmt.Sprint(fiftyPercentCPU),
		"-m", fmt.Sprint(fiftyPercentMemory),
		"-n", "nunet-team").Output()
	s.Contains(string(out), "Error: -a | --address must be specified")
	if err != nil {
		s.T().Errorf("Command execution failed: %v", err)
	}
}

func (s *CLI) TestNunetOnboardLowMemoryCLI() {
	provisioned := onboarding.GetTotalProvisioned()
	availableMemory, availableCPU := provisioned.Memory, provisioned.CPU
	fivePercentMemory := availableMemory / 20
	fiftyPercentCPU := uint64(availableCPU / 2)
	out, _ := exec.Command("../maint-scripts/nunet-dms/usr/bin/nunet",
		"onboard",
		"-c", fmt.Sprint(fiftyPercentCPU),
		"-m", fmt.Sprint(fivePercentMemory),
		"-a", "testaddress",
		"-n", "nunet-team").Output()
	s.Contains(string(out), "Memory should be between 10% and 90% of the available memory")
}

func (s *CLI) TestNunetOnboardLowCPUCLI() {
	provisioned := onboarding.GetTotalProvisioned()
	availableMemory, availableCPU := provisioned.Memory, provisioned.CPU
	halfMemory := availableMemory / 2
	fivePercentCPU := uint64(availableCPU / 20)
	out, _ := exec.Command("../maint-scripts/nunet-dms/usr/bin/nunet",
		"onboard",
		"-c", fmt.Sprint(fivePercentCPU),
		"-m", fmt.Sprint(halfMemory),
		"-a", "testaddress",
		"-n", "nunet-team").Output()

	s.Contains(string(out), "CPU should be between 10% and 90% of the available CPU")
}

func (s *CLI) TestNunetInfoNoMetadataCLI() {
	testMetadataPath := "test/etc/nunet/metadataV2.json"
	backupMetadataPath := "test/etc/nunet/metadata_backup.json"

	os.MkdirAll("test/etc/nunet", 0755)
	defer os.RemoveAll("test/etc/nunet") // Clean up after the test

	// Create a mock metadata file for the test
	err := os.WriteFile(testMetadataPath, []byte("mock data"), 0644)
	s.Require().NoError(err, "Error creating mock metadata file")

	// Remove or rename the metadata file
	err = os.Rename(testMetadataPath, backupMetadataPath)
	s.Require().NoError(err, "Error renaming mock metadata file")
	defer os.Rename(backupMetadataPath, testMetadataPath)

	out, cmdErr := exec.Command("../maint-scripts/nunet-dms/usr/bin/nunet", "info").CombinedOutput()
	s.Require().NoError(cmdErr, "Error executing nunet info command")
	s.Contains(string(out), "unable to read metadata.json", "Expected error message not found in command output")
}

func (s *CLI) TestNunetInfoCLICPU() {
	gpu, _ := library.Check_gpu()
	if len(gpu) == 0 {
		expectedMetadata := models.MetadataV2{
			Name:            "",
			UpdateTimestamp: 1696269936,
			Resource: struct {
				MemoryMax int64 `json:"memory_max,omitempty"`
				TotalCore int64 `json:"total_core,omitempty"`
				CPUMax    int64 `json:"cpu_max,omitempty"`
			}{
				MemoryMax: 64305,
				TotalCore: 12,
				CPUMax:    43200,
			},
			Available: struct {
				CPU    int64 `json:"cpu,omitempty"`
				Memory int64 `json:"memory,omitempty"`
			}{
				CPU:    38200,
				Memory: 57305,
			},
			Reserved: struct {
				CPU    int64 `json:"cpu,omitempty"`
				Memory int64 `json:"memory,omitempty"`
			}{
				CPU:    5000,
				Memory: 7000,
			},
			Network:   "nunet-team",
			PublicKey: "addr_test1qr6jk9ty2xhcxvqcy8w7mr22m03vd0de7hxp7nk0xw0x6slmar57e34n47qyy4zzdlzulrd9e6udfsw4r05qshm2gyesew4fxk",
		}

		// Convert expectedMetadata to JSON string
		expectedJsonBytes, err := json.Marshal(expectedMetadata)
		s.Nil(err)
		expectedJsonString := string(expectedJsonBytes)

		onboarding.AFS.WriteFile(
			fmt.Sprintf("%s/metadataV2.json", config.GetConfig().MetadataPath),
			[]byte(expectedJsonString), 0644)

		out, _ := exec.Command("../maint-scripts/nunet-dms/usr/bin/nunet", "info").Output()

		// Unmarshal the output into receivedMetadata
		var receivedMetadata models.MetadataV2
		err = json.Unmarshal(out, &receivedMetadata)
		s.Nil(err)

		// Compare the expected and received metadata
		s.Equal(expectedMetadata, receivedMetadata)

		// Clean up
		onboarding.AFS.Remove(fmt.Sprintf("%s/metadataV2.json", config.GetConfig().MetadataPath))
	} else {
		s.T().Skip("Skipping tests because GPUs are available.")
	}
}

func (s *CLI) TestNunetInfoCLIGPU() {
	gpu, _ := library.Check_gpu()
	if len(gpu) > 0 {
		expectedJsonString := `{
		"name": "TestMachineData",
		"update_timestamp": 11111111111,
		"resource": {
		 "memory_max": 1,
		 "total_core": 1,
		 "cpu_max": 1
		},
		"available": {
		 "cpu": 1,
		 "memory": 1
		},
		"reserved": {
		 "cpu": 1,
		 "memory": 1
		},
		"network": "test-data",
		"gpu_info":[{
			"name":"",
			"tot_vram": 0,
			"free_vram": 0
		}]
		}`
		onboarding.AFS.WriteFile("/etc/nunet/metadataV2.json", []byte(expectedJsonString), 0644)
		out, _ := exec.Command("../maint-scripts/nunet-dms/usr/bin/nunet", "info").Output()
		var expectedMetadata, receivedMetadata models.MetadataV2

		output := string(out)
		index := strings.LastIndex(output, "}")
		js := output[:index+1]
		fmt.Println(js)
		bArray := []byte(js)
		err := json.Unmarshal(bArray, &receivedMetadata)
		s.Nil(err)
		gpu_log := output[index+1:]
		s.Contains(gpu_log, gpu[0].Name)
		err = json.Unmarshal([]byte(expectedJsonString), &expectedMetadata)
		s.Nil(err)
		s.Equal(expectedMetadata, receivedMetadata)
	}

}
