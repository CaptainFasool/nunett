package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding"
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

type MyTestSuite struct {
	suite.Suite
	sync.WaitGroup
}

func TestMyTestSuite(t *testing.T) {
	s := new(MyTestSuite)
	suite.Run(t, s)
}

func (s *MyTestSuite) SetupSuite() {
	s.WaitGroup.Add(1)

	router := SetupRouter()
	go router.Run(":9999")
}

func (s *MyTestSuite) TearDownSuite() {
	s.WaitGroup.Done()
}

func (s *MyTestSuite) TestNunetWalletNewEthereumCLI() {
	out, err := exec.Command("./maint-scripts/nunet-dms/usr/bin/nunet",
		"wallet", "new", "-e").Output()

	s.Nil(err)

	type WalletNewEthereumOutput struct {
		Address    string `json:"address"`
		PrivateKey string `json:"private_key"`
	}
	var resp WalletNewEthereumOutput
	err = json.Unmarshal(out, &resp)

	s.Nil(err)
}

func (s *MyTestSuite) TestNunetWalletNewCardanoCLI() {
	out, err := exec.Command("./maint-scripts/nunet-dms/usr/bin/nunet",
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

func (s *MyTestSuite) TestNuNetAvailableCLI() {
	out, err := exec.Command("./maint-scripts/nunet-dms/usr/bin/nunet",
		"available").Output()

	s.Nil(err)

	type AvailableOutput struct {
		CPU      float64 `json:"cpu"`
		Memory   uint64  `json:"memory"`
		NumCores uint64  `json:"total_cores"`
	}
	var resp AvailableOutput
	err = json.Unmarshal(out, &resp)

	s.Nil(err)
}

func (s *MyTestSuite) TestNunetOnboardNoCPUValueCLI() {
	availableMemory := onboarding.GetTotalProvisioned().Memory
	halfMemory := availableMemory / 2
	out, _ := exec.Command("./maint-scripts/nunet-dms/usr/bin/nunet",
		"onboard",
		"-m", fmt.Sprint(halfMemory),
		"-a", "testaddress",
		"-n", "nunet-team").Output()
	s.Contains(string(out), "Error: -c | --cpu must be specified")
}

func (s *MyTestSuite) TestNunetOnboardNoMemoryValueCLI() {
	availableCPU := onboarding.GetTotalProvisioned().CPU
	halfCPU := availableCPU / 2
	out, _ := exec.Command("./maint-scripts/nunet-dms/usr/bin/nunet",
		"onboard",
		"-c", fmt.Sprint(halfCPU),
		"-a", "testaddress",
		"-n", "nunet-team").Output()
	s.Contains(string(out), "Error: -m | --memory must be specified")
}

func (s *MyTestSuite) TestNunetOnboardNoChannelCLI() {
	provisioned := onboarding.GetTotalProvisioned()
	availableMemory, availableCPU := provisioned.Memory, provisioned.CPU
	fiftyPercentMemory := availableMemory / 2
	fiftyPercentCPU := availableCPU / 2
	out, _ := exec.Command("./maint-scripts/nunet-dms/usr/bin/nunet",
		"onboard",
		"-c", fmt.Sprint(fiftyPercentCPU),
		"-m", fmt.Sprint(fiftyPercentMemory),
		"-a", "testaddress").Output()
	s.Contains(string(out), "Error: -n | --nunet-channel must be specified")
}

func (s *MyTestSuite) TestNunetOnboardNoAddressCLI() {
	provisioned := onboarding.GetTotalProvisioned()
	availableMemory, availableCPU := provisioned.Memory, provisioned.CPU
	fiftyPercentMemory := availableMemory / 2
	fiftyPercentCPU := uint64(availableCPU / 2)
	out, _ := exec.Command("./maint-scripts/nunet-dms/usr/bin/nunet",
		"onboard",
		"-c", fmt.Sprint(fiftyPercentCPU),
		"-m", fmt.Sprint(fiftyPercentMemory),
		"-n", "nunet-team").Output()
	s.Contains(string(out), "Error: -a | --address must be specified")
}

func (s *MyTestSuite) TestNunetOnboardLowMemoryCLI() {
	provisioned := onboarding.GetTotalProvisioned()
	availableMemory, availableCPU := provisioned.Memory, provisioned.CPU
	fivePercentMemory := availableMemory / 20
	fiftyPercentCPU := uint64(availableCPU / 2)
	out, _ := exec.Command("./maint-scripts/nunet-dms/usr/bin/nunet",
		"onboard",
		"-c", fmt.Sprint(fiftyPercentCPU),
		"-m", fmt.Sprint(fivePercentMemory),
		"-a", "testaddress",
		"-n", "nunet-team").Output()
	s.Contains(string(out), "Memory should be between 10% and 90% of the available memory")
}

func (s *MyTestSuite) TestNunetOnboardLowCPUCLI() {
	provisioned := onboarding.GetTotalProvisioned()
	availableMemory, availableCPU := provisioned.Memory, provisioned.CPU
	halfMemory := availableMemory / 2
	fivePercentCPU := uint64(availableCPU / 20)
	out, _ := exec.Command("./maint-scripts/nunet-dms/usr/bin/nunet",
		"onboard",
		"-c", fmt.Sprint(fivePercentCPU),
		"-m", fmt.Sprint(halfMemory),
		"-a", "testaddress",
		"-n", "nunet-team").Output()

	s.Contains(string(out), "CPU should be between 10% and 90% of the available CPU")
}

func (s *MyTestSuite) TestNunetOnboardLowCardanoValuesCLI() {
	onboarding.FS = afero.NewMemMapFs()
	onboarding.AFS = &afero.Afero{Fs: onboarding.FS}
	onboarding.AFS.Mkdir("/etc/nunet", 0777)

	provisioned := onboarding.GetTotalProvisioned()
	availableMemory, availableCPU := provisioned.Memory, provisioned.CPU
	var minimumCardanoMemory uint64 = 10000
	var minimumCardanoCPU uint64 = 6000
	var useCPU, useMemory uint64
	if availableCPU < float64(minimumCardanoCPU) {
		useCPU = uint64(availableCPU)
	} else {
		useCPU = 5900
	}
	if uint64(availableMemory) < uint64(minimumCardanoMemory) {
		useMemory = uint64(availableMemory)
	} else {
		useMemory = 9900
	}

	out, _ := exec.Command("./maint-scripts/nunet-dms/usr/bin/nunet",
		"onboard",
		"-c", fmt.Sprint(useCPU),
		"-m", fmt.Sprint(useMemory),
		"-a", "testaddress",
		"-n", "nunet-test",
		"--cardano").Output()
	s.Contains(string(out), "cardano node requires 10000MB of RAM and 6000MHz CPU")
}

func (s *MyTestSuite) TestNunetInfoNoMetadataCLI() {
	onboarding.FS = afero.NewMemMapFs()
	onboarding.AFS = &afero.Afero{Fs: onboarding.FS}
	onboarding.AFS.Mkdir("/etc/nunet", 0777)

	out, _ := exec.Command("./maint-scripts/nunet-dms/usr/bin/nunet", "info").Output()
	s.Contains(string(out), "metadata.json does not exists or not readable")
}

func (s *MyTestSuite) TestNunetInfoCLICPU() {
	gpu, _ := onboarding.Check_gpu()
	if len(gpu) == 0 {

		onboarding.FS = afero.NewMemMapFs()
		onboarding.AFS = &afero.Afero{Fs: onboarding.FS}
		onboarding.AFS.Mkdir("/etc/nunet", 0777)
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
		"public_key": "test-address"
		}`
		onboarding.AFS.WriteFile("/etc/nunet/metadataV2.json", []byte(expectedJsonString), 0644)
		out, _ := exec.Command("./maint-scripts/nunet-dms/usr/bin/nunet", "info").Output()
		var expectedMetadata, receivedMetadata models.MetadataV2

		err := json.Unmarshal(out, &receivedMetadata)
		s.Nil(err)
		err = json.Unmarshal([]byte(expectedJsonString), &expectedMetadata)
		s.Nil(err)
		s.Equal(expectedMetadata, receivedMetadata)
	} else {
		s.T().Skip("Skipping tests because GPUs are available.")
	}
}

func (s *MyTestSuite) TestNunetInfoCLIGPU() {
	gpu, _ := onboarding.Check_gpu()
	if len(gpu) > 0 {
		onboarding.FS = afero.NewMemMapFs()
		onboarding.AFS = &afero.Afero{Fs: onboarding.FS}
		onboarding.AFS.Mkdir("/etc/nunet", 0777)
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
		out, _ := exec.Command("./maint-scripts/nunet-dms/usr/bin/nunet", "info").Output()
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
