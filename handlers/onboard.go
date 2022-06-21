package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding"
)

// GetMetadata      godoc
// @Summary      Get current device info.
// @Description  Responds with metadata of current provideer
// @Tags         onboard
// @Produce      json
// @Success      200  {array}        models.Metadata
// @Failure      500  {object}		 httputil.HTTPError
// @Router       /metadata [get]
func GetMetadata(c *gin.Context) {
	// read the info
	content, err := ioutil.ReadFile("/etc/nunet/metadataV2.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "metadata.json does not exists or not readable"})
		return
	}

	// deserialize to json
	var metadata models.Metadata
	err = json.Unmarshal(content, &metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "unable to parse metadata.json"})
		return
	}

	c.JSON(http.StatusOK, metadata)
}

// Onboard      godoc
// @Summary      Runs the onboarding process.
// @Description  Onboard runs onboarding script given the amount of resources to onboard.
// @Tags         onboard
// @Produce      json
// @Success      200  {array}  models.Metadata
// @Router       /onboard [post]
func Onboard(c *gin.Context) {
	hostname, _ := os.Hostname()

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
		c.JSON(http.StatusConflict,
			gin.H{"error": "wrong capacity provided"})
		return
	}

	cardanoPassive := "no"
	if capacityForNunet.Cardano {
		if capacityForNunet.Memory < 10000 || capacityForNunet.CPU < 6000 {
			c.JSON(http.StatusConflict,
				gin.H{"error": "cardano node requires 10000MB of RAM and 6000MHz CPU"})
			return
		}
		cardanoPassive = "yes"
	}

	if capacityForNunet.Channel != "nunet-development" &&
		capacityForNunet.Channel != "nunet-private-alpha" {
		c.JSON(http.StatusConflict,
			gin.H{"error": "only nunet-development and nunet-private-alpha is supported at the moment"})
		return
	}

	metadata.Reserved.Memory = capacityForNunet.Memory
	metadata.Reserved.CPU = capacityForNunet.CPU

	metadata.Available.Memory = int64(totalMem) - capacityForNunet.Memory
	metadata.Available.CPU = int64(totalCpu) - capacityForNunet.CPU

	metadata.Network = capacityForNunet.Channel
	metadata.PublicKey = capacityForNunet.PaymentAddress

	file, _ := json.MarshalIndent(metadata, "", " ")
	err := ioutil.WriteFile("/etc/nunet/metadataV2.json", file, 0644)
	if err != nil {
		panic(err)
	}

	onboarding.CreateClientConfig(c, &metadata, &capacityForNunet, hostname)
	onboarding.CreateAdapterConfig(c, &metadata, cardanoPassive, hostname)

	var adapterPrefix string
	if metadata.Network == "nunet-development" {
		adapterPrefix = "testing-nunet-adapter"
	}
	if metadata.Network == "nunet-private-alpha" {
		adapterPrefix = "nunet-adapter"
	}

	jobName := adapterPrefix + "-" + hostname
	onboarding.RunNomadJob(c, jobName)

	c.JSON(http.StatusCreated, metadata)
}

// Echo responds back with same JSON it has received.
func Echo(c *gin.Context) {
	var json map[string]interface{}
	c.BindJSON(&json)
	c.JSON(http.StatusOK, json)
}

// DeviceUsage streams device resources usage to client.
func DeviceUsage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "will stream device usage"})
}

// SetPreferences set preferences.
func SetPreferences(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "will set preferences to distributed persistent system"})
}

// ProvisionedCapacity      godoc
// @Summary      Returns provisioned capacity on host.
// @Description  Get total memory capacity in MB and CPU capacity in MHz.
// @Tags         onboard
// @Produce      json
// @Success      200  {object}  models.Provisioned
// @Router       /provisioned [get]
func ProvisionedCapacity(c *gin.Context) {
	c.JSON(http.StatusOK, onboarding.GetTotalProvisioned())
}

// CreatePaymentAddress      godoc
// @Summary      Create a new payment address.
// @Description  Create a payment address from public key. Return payment address and private key.
// @Tags         onboard
// @Produce      json
// @Failure      500  {object}
// @Success      200  {object}  models.AddressPrivKey
// @Router       /address/new [get]
func CreatePaymentAddress(c *gin.Context) {
	pair, err := onboarding.GetAddressAndPrivateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"message": "error creating address"})
	}
	c.JSON(http.StatusOK, pair)
}

// https://github.com/swaggo/swag
