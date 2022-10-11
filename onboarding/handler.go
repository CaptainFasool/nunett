package onboarding

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/models"
)

// GetMetadata      godoc
// @Summary      Get current device info.
// @Description  Responds with metadata of current provideer
// @Tags         onboarding
// @Produce      json
// @Success      200  {array}        models.Metadata
// @Router       /onboarding/metadata [get]
func GetMetadata(c *gin.Context) {
	// read the info
	content, err := os.ReadFile("/etc/nunet/metadataV2.json")
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
// @Tags         onboarding
// @Produce      json
// @Success      200  {array}  models.Metadata
// @Router       /onboarding/onboard [post]
func Onboard(c *gin.Context) {
	// check if request body is empty
	if c.Request.ContentLength == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body is empty"})
		return
	}

	_, err := os.Stat("/etc/nunet")
	if os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "/etc/nunet does not exist. is nunet onboaded successfully?"})
		return
	}

	hostname, _ := os.Hostname()

	currentTime := time.Now().Unix()

	totalCpu := GetTotalProvisioned().CPU
	totalMem := GetTotalProvisioned().Memory
	numCores := GetTotalProvisioned().NumCores

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
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "wrong capacity provided"})
		return
	}

	cardanoPassive := "no"
	if capacityForNunet.Cardano {
		if capacityForNunet.Memory < 10000 || capacityForNunet.CPU < 6000 {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "cardano node requires 10000MB of RAM and 6000MHz CPU"})
			return
		}
		cardanoPassive = "yes"
	}

	if capacityForNunet.Channel != "nunet-development" &&
		capacityForNunet.Channel != "nunet-private-alpha" {
		c.JSON(http.StatusBadRequest,
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
	err = os.WriteFile("/etc/nunet/metadataV2.json", file, 0644)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "cound not write metadata.json"})
		return
	}

	go InstallRunAdapter(c, hostname, &metadata, cardanoPassive)

	c.JSON(http.StatusCreated, metadata)
}

// ProvisionedCapacity      godoc
// @Summary      Returns provisioned capacity on host.
// @Description  Get total memory capacity in MB and CPU capacity in MHz.
// @Tags         onboarding
// @Produce      json
// @Success      200  {object}  models.Provisioned
// @Router       /onboarding/provisioned [get]
func ProvisionedCapacity(c *gin.Context) {
	c.JSON(http.StatusOK, GetTotalProvisioned())
}

// CreatePaymentAddress      godoc
// @Summary      Create a new payment address.
// @Description  Create a payment address from public key. Return payment address and private key.
// @Tags         onboarding
// @Produce      json
// @Success      200  {object}  models.AddressPrivKey
// @Router       /onboarding/address/new [get]
func CreatePaymentAddress(c *gin.Context) {
	pair, err := GetAddressAndPrivateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"message": "error creating address"})
	}
	c.JSON(http.StatusOK, pair)
}

// https://github.com/swaggo/swag
