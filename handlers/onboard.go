package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding"
)

// Onboarded      godoc
// @Summary      Get current device info.
// @Description  Responds with metadata of current provideer
// @Tags         onboard
// @Produce      json
// @Success      200  {array}        models.Metadata
// @Router       /onboard [get]
func Onboarded(c *gin.Context) {
	// read the info
	content, err := ioutil.ReadFile("/etc/nunet/metadata.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	// deserialize to json
	var metadata models.Metadata
	err = json.Unmarshal(content, &metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, metadata)
}

// Onboarded      godoc
// @Summary      Runs the onboarding process.
// @Description  Onboard runs onboarding script given the amount of resources to onboard.
// @Tags         onboard
// @Produce      json
// @Success      200  {array}  models.Metadata
// @Router       /onboard [post]
func Onboard(c *gin.Context) {
	hostname, err := os.Hostname()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	currentTime := time.Now().Unix()

	totalCpu := onboarding.GetTotalProvisioned().CPU
	totalMem := onboarding.GetTotalProvisioned().Memory
	numCores := onboarding.GetTotalProvisioned().NumCores

	// create metadata
	var metadata models.MetadataV2

	metadata.Name = hostname
	metadata.UpdateTimestamp = currentTime
	metadata.Resource.MemoryMax = int64(totalMem)
	metadata.Resource.TotalCore = int64(numCores)
	metadata.Resource.CpuMax = int64(totalCpu)

	// read the request body to fill rest of the fields

	// get capacity user want to rent to NuNet
	var capacityForNunet models.CapacityForNunet
	c.BindJSON(&capacityForNunet)

	if (capacityForNunet.Memory > int64(totalMem)) &&
		capacityForNunet.CPU > int64(totalCpu) {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": errors.New("wrong capacity provided")})
		return
	}

	cardanoPassive := "no"
	if capacityForNunet.Cardano {
		if capacityForNunet.Memory < 10000 || capacityForNunet.CPU < 6000 {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": errors.New("cardano node requires 10000MB of RAM and 6000MHz CPU")})
			return
		}
		cardanoPassive = "yes"
	}

	metadata.Reserved.Memory = capacityForNunet.Memory
	metadata.Reserved.CPU = capacityForNunet.CPU

	metadata.Available.Memory = int64(totalMem) - capacityForNunet.Memory
	metadata.Available.CPU = int64(totalCpu) - capacityForNunet.CPU

	metadata.Network = capacityForNunet.Channel
	metadata.PublicKey = capacityForNunet.PaymentAddress

	file, _ := json.MarshalIndent(metadata, "", " ")
	err = ioutil.WriteFile("/etc/nunet/metadataV2.json", file, 0644)
	if err != nil {
		fmt.Println(err)
	}

	// create client config
	clientData := struct {
		Name           string
		Network        string
		ReservedCPU    int64
		ReservedMemory int64
		Cardano        string
	}{
		Name:           hostname,
		Network:        capacityForNunet.Channel,
		ReservedCPU:    metadata.Reserved.CPU,
		ReservedMemory: metadata.Reserved.Memory,
		Cardano:        cardanoPassive,
	}

	clientFile, err := os.Create("/etc/nunet/clientV2.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	tmpl, err := template.New("client").Parse(models.ClientTemplate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = tmpl.Execute(clientFile, clientData)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	// create adapter-definition.json
	var adapterPrefix string
	var dockerImageTag string
	var deploymentType string
	var tokenomicsApiName string
	if metadata.Network == "nunet-development" {
		dockerImageTag = "test"
		adapterPrefix = "testing-nunet-adapter"
		deploymentType = "test"
		tokenomicsApiName = "testing-tokenomics"
		if os.Getenv("OS_RELEASE") == "raspbian" {
			dockerImageTag = "arm-test"
		}
	}
	if metadata.Network == "nunet-private-alpha" {
		dockerImageTag = "latest"
		adapterPrefix = "nunet-adapter"
		deploymentType = "prod"
		tokenomicsApiName = "tokenomics"
		if os.Getenv("OS_RELEASE") == "raspbian" {
			dockerImageTag = "arm-latest"
		}
	}
	adapterData := struct {
		Datacenters       string
		AdapterPrefix     string
		ClientName        string
		DockerTag         string
		DeploymentType    string
		TokenomicsApiName string
	}{
		Datacenters:       metadata.Network,
		AdapterPrefix:     adapterPrefix,
		ClientName:        hostname,
		DockerTag:         dockerImageTag,
		DeploymentType:    deploymentType,
		TokenomicsApiName: tokenomicsApiName,
	}

	adapterFile, err := os.Create("/etc/nunet/adapter-definitionV2.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	tmpl, err = template.New("adapter").Parse(models.AdapterTemplate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = tmpl.Execute(adapterFile, adapterData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, metadata)
}

// DeviceUsage streams device resources usage to client.
func DeviceUsage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "will stream device usage"})
}

// SetPreferences set preferences.
func SetPreferences(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "will set preferences to distributed persistent system"})
}

// Onboarded      godoc
// @Summary      Returns provisioned capacity on host.
// @Description  Get total memory capacity in MB and CPU capacity in MHz.
// @Tags         onboard
// @Produce      json
// @Success      200  {object}  models.Provisioned
// @Router       /provisioned [get]
func ProvisionedCapacity(c *gin.Context) {
	c.JSON(http.StatusOK, onboarding.GetTotalProvisioned())
}

// Onboarded      godoc
// @Summary      Create a new payment address.
// @Description  Create a payment address from public key. Return payment address and private key.
// @Tags         onboard
// @Produce      json
// @Success      200  {object}  models.Provisioned
// @Router       /address/new [get]
func CreatePaymentAddress(c *gin.Context) {
	pair, err := onboarding.GetAddressAndPrivateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"message": "Error creating address"})
	}
	c.JSON(http.StatusOK, pair)
}

// https://github.com/swaggo/swag
