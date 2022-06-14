package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

	if capacityForNunet.Memory > int64(totalMem) && capacityForNunet.CPU > int64(totalCpu) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.New("wrong capacity provided")})
		return
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
