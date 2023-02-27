package onboarding

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/firecracker/telemetry"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"

	"github.com/spf13/afero"
)

var FS afero.Fs = afero.NewOsFs()
var AFS *afero.Afero = &afero.Afero{Fs: FS}

// GetMetadata      godoc
// @Summary      Get current device info.
// @Description  Responds with metadata of current provideer
// @Tags         onboarding
// @Produce      json
// @Success      200  {array}        models.Metadata
// @Router       /onboarding/metadata [get]
func GetMetadata(c *gin.Context) {
	// check if the request has any body data
	// if it has return that body  and skip the below code
	// just for the test cases

	// read the info
	content, err := AFS.ReadFile("/etc/nunet/metadataV2.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "metadata.json does not exists or not readable"})
		return
	}

	// deserialize to json
	var metadata models.MetadataV2
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

	_, err := AFS.Stat("/etc/nunet")
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
	metadata.Resource.CPUMax = int64(totalCpu)

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
		metadata.AllowCardano = true
	}

	if capacityForNunet.Channel != "nunet-staging" &&
		capacityForNunet.Channel != "nunet-test" &&
		capacityForNunet.Channel != "nunet-team" &&
		capacityForNunet.Channel != "nunet-edge" {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "channel name not supported! nunet-test, nunet-edge, nunet-team and nunet-staging are supported at the moment"})
		return
	}

	metadata.Reserved.Memory = capacityForNunet.Memory
	metadata.Reserved.CPU = capacityForNunet.CPU

	metadata.Available.Memory = int64(totalMem) - capacityForNunet.Memory
	metadata.Available.CPU = int64(totalCpu) - capacityForNunet.CPU

	metadata.Network = capacityForNunet.Channel
	metadata.PublicKey = capacityForNunet.PaymentAddress

	if !fileExists("/etc/nunet/metadataV2.json") {
		file, _ := json.MarshalIndent(metadata, "", " ")
		err = AFS.WriteFile("/etc/nunet/metadataV2.json", file, 0644)
		if err != nil {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "could not write metadata.json"})
			return
		}
	} else {
		content, err := AFS.ReadFile("/etc/nunet/metadataV2.json")
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": "metadata.json does not exists or not readable"})
			return
		}
		var metadata2 models.MetadataV2
		err = json.Unmarshal(content, &metadata2)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": "unable to parse metadata.json"})
			return
		}
		metadata.NodeID = metadata2.NodeID
		file, _ := json.MarshalIndent(metadata, "", " ")
		err = AFS.WriteFile("/etc/nunet/metadataV2.json", file, 0644)
		if err != nil {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "could not write metadata.json"})
			return
		}
	}

	// Add available resources to database.

	available_resources := models.AvailableResources{
		TotCpuHz:  int(capacityForNunet.CPU),
		CpuNo:     int(numCores),
		CpuHz:     hz_per_cpu(),
		PriceCpu:  0, // TODO: Get price of CPU
		Ram:       int(capacityForNunet.Memory),
		PriceRam:  0, // TODO: Get price of RAM
		Vcpu:      int(math.Floor((float64(capacityForNunet.CPU)) / hz_per_cpu())),
		Disk:      0,
		PriceDisk: 0,
	}

	var availableRes models.AvailableResources
	if res := db.DB.Find(&availableRes); res.RowsAffected == 0 {
		result := db.DB.Create(&available_resources)
		if result.Error != nil {
			panic(result.Error)
		}
	} else {
		result := db.DB.Model(&models.AvailableResources{}).Where("id = ?", 1).Updates(available_resources)
		if result.Error != nil {
			panic(result.Error)
		}
	}

	priv, pub, err := libp2p.GenerateKey(0)
	if err != nil {
		panic(err)
	}
	telemetry.CalcFreeResources()
	libp2p.SaveKey(priv, pub)
	libp2p.RunNode(priv)

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
// @Success      200  {object}  models.BlockchainAddressPrivKey
// @Router       /onboarding/address/new [get]
func CreatePaymentAddress(c *gin.Context) {
	blockChain := c.DefaultQuery("blockchain", "cardano")

	var pair *models.BlockchainAddressPrivKey
	var err error
	if blockChain == "ethereum" {
		pair, err = GetEthereumAddressAndPrivateKey()
	} else if blockChain == "cardano" {
		pair, err = GetCardanoAddressAndMnemonic()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"message": "error creating address"})
	}
	c.JSON(http.StatusOK, pair)
}

func fileExists(filename string) bool {
	info, err := AFS.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
