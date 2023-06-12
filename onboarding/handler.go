package onboarding

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/firecracker/telemetry"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/heartbeat"
	"gitlab.com/nunet/device-management-service/internal/klogger"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"

	"github.com/spf13/afero"
)

var FS afero.Fs = afero.NewOsFs()
var AFS *afero.Afero = &afero.Afero{Fs: FS}

// GetMetadata      godoc
//
//	@Summary		Get current device info.
//	@Description	Responds with metadata of current provideer
//	@Tags			onboarding
//	@Produce		json
//	@Success		200	{object}	models.Metadata
//	@Router			/onboarding/metadata [get]
func GetMetadata(c *gin.Context) {
	// check if the request has any body data
	// if it has return that body  and skip the below code
	// just for the test cases
	// read the info
	content, err := AFS.ReadFile(fmt.Sprintf("%s/metadataV2.json", config.GetConfig().General.MetadataPath))
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

// ProvisionedCapacity      godoc
//
//	@Summary		Returns provisioned capacity on host.
//	@Description	Get total memory capacity in MB and CPU capacity in MHz.
//	@Tags			onboarding
//	@Produce		json
//	@Success		200	{object}	models.Provisioned
//	@Router			/onboarding/provisioned [get]
func ProvisionedCapacity(c *gin.Context) {
	totalProvisioned := GetTotalProvisioned()
	totalPJ, err := json.Marshal(totalProvisioned)
	if err != nil {
		zlog.Sugar().ErrorfContext(c.Request.Context(), "couldn't marshal totalProvisioned to json: %v", string(totalPJ))
	}
	c.JSON(http.StatusOK, GetTotalProvisioned())
}

// CreatePaymentAddress      godoc
//
//	@Summary		Create a new payment address.
//	@Description	Create a payment address from public key. Return payment address and private key.
//	@Tags			onboarding
//	@Produce		json
//	@Success		200	{object}	models.BlockchainAddressPrivKey
//	@Router			/onboarding/address/new [get]
func CreatePaymentAddress(c *gin.Context) {
	// send telemetry data
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

// Status      godoc
//
//	@Summary		Onboarding status and other metadata.
//	@Description	Returns json with 5 parameters: onboarded, error, machine_uuid, metadata_path, database_path.
//					  `onboarded` is true if the device is onboarded, false otherwise.
//					  `error` is the error message if any related to onboarding status check
//					  `machine_uuid` is the UUID of the machine
//					  `metadata_path` is the path to metadataV2.json only if it exists
//					  `database_path` is the path to nunet.db only if it exists
//	@Tags			onboarding
//	@Produce		json
//	@Success		200	{object} models.OnboardingStatus
//	@Router			/onboarding/status [get]
func Status(c *gin.Context) {
	onboarded, err := utils.IsOnboarded()
	var metadataPath string
	var dbPath string
	if metadataPath = fmt.Sprintf("%s/metadataV2.json", config.GetConfig().General.MetadataPath); !fileExists(metadataPath) {
		metadataPath = ""
	}
	if dbPath = fmt.Sprintf("%s/nunet.db", config.GetConfig().General.MetadataPath); !fileExists(dbPath) {
		dbPath = ""
	}

	resp := models.OnboardingStatus{
		Onboarded:    onboarded,
		Error:        fmt.Sprintf("%v", err),
		MachineUUID:  utils.GetMachineUUID(),
		MetadataPath: metadataPath,
		DatabasePath: dbPath,
	}
	c.JSON(http.StatusOK, resp)
}

// Onboard      godoc
//
//	@Summary		Runs the onboarding process.
//	@Description	Onboard runs onboarding script given the amount of resources to onboard.
//	@Tags			onboarding
//	@Produce		json
//	@Success		200	{object}	models.Metadata
//	@Router			/onboarding/onboard [post]
func Onboard(c *gin.Context) {
	// get capacity user want to rent to NuNet
	capacityForNunet := models.CapacityForNunet{ServerMode: true}
	if err := c.BindJSON(&capacityForNunet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	_, err := AFS.Stat(config.GetConfig().General.MetadataPath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": fmt.Sprintf("%s does not exist. is nunet installed correctly?", config.GetConfig().General.MetadataPath)})
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

	// validate the public (payment) address
	if err := ValidateAddress(capacityForNunet.PaymentAddress); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// validate dedicated capacity to NuNet (should be between 10% to 90%)
	if err := validateCapacityForNunet(capacityForNunet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metadata.AllowCardano = false
	if capacityForNunet.Cardano {
		if capacityForNunet.Memory < 10000 || capacityForNunet.CPU < 6000 {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "cardano node requires 10000MB of RAM and 6000MHz CPU"})
			klogger.Logger.Error("onboarding error : wrong capacity provided")
			return
		}
		metadata.AllowCardano = true
	}

	gpu_info, err := Check_gpu()
	if err != nil {
		zlog.Sugar().Errorf("Unable to detect GPU: %v ", err.Error())
	}
	metadata.GpuInfo = gpu_info

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

	metadata.Plugins = capacityForNunet.Plugins

	file, _ := json.MarshalIndent(metadata, "", " ")
	err = AFS.WriteFile(fmt.Sprintf("%s/metadataV2.json", config.GetConfig().General.MetadataPath), file, 0644)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "could not write metadata.json"})
		return
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
	if res := db.DB.WithContext(c.Request.Context()).Find(&availableRes); res.RowsAffected == 0 {
		result := db.DB.WithContext(c.Request.Context()).Create(&available_resources)
		if result.Error != nil {
			zlog.Panic(result.Error.Error())
		}
	} else {
		result := db.DB.WithContext(c.Request.Context()).Model(&models.AvailableResources{}).Where("id = ?", 1).Updates(available_resources)
		if result.Error != nil {
			zlog.Panic(result.Error.Error())
		}
	}

	priv, pub, err := libp2p.GenerateKey(0)
	if err != nil {
		zlog.Panic(err.Error())
	}
	libp2p.SaveNodeInfo(priv, pub, capacityForNunet.ServerMode)

	telemetry.CalcFreeResources()
	libp2p.RunNode(priv, capacityForNunet.ServerMode)

	_, err = heartbeat.NewToken(libp2p.GetP2P().Host.ID().String(), capacityForNunet.Channel)
	if err != nil {
		zlog.Sugar().Errorf("unable to get new telemetry token: %v", err)
	}
	klogger.Logger.Info("device onboarded")

	_, err = utils.RegisterLogbin(utils.GetMachineUUID(), libp2p.GetP2P().Host.ID().String())
	if err != nil {
		zlog.Sugar().Errorf("unable to register with logbin: %v", err)
	}

	c.JSON(http.StatusCreated, metadata)
}

// Config        godoc
//
//	@Summary	changes the amount of resources of onboarded device .
//	@Tags		onboarding
//	@Produce	json
//	@Success	200	{object}	models.Metadata
//	@Router		/onboarding/resource-config [post]
func ResourceConfig(c *gin.Context) {
	klogger.Logger.Info("device resource change started")
	// check if request body is empty
	if c.Request.ContentLength == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body is empty"})
		return
	}

	// _, err := AFS.Stat(fmt.Sprintf("%s/metadataV2.json", config.GetConfig().General.MetadataPath))
	if onboarded, err := utils.IsOnboarded(); !onboarded {
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "problem with machine onboarding: " + err.Error()})
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "machine not onboarded"})
		}
		return
	}

	// reading the request body
	capacityForNunet := models.CapacityForNunet{}
	if err := c.BindJSON(&capacityForNunet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	// validate dedicated capacity to NuNet (should be between 10% to 90%)
	if err := validateCapacityForNunet(capacityForNunet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// read metadataV2.json file and update it with new resources
	metadata, err := utils.ReadMetadataFile()
	if err != nil {
		zlog.Sugar().Errorf("could not read metadata: %v", err)
	}
	metadata.Reserved.CPU = capacityForNunet.CPU
	metadata.Reserved.Memory = capacityForNunet.Memory

	// read the existing data and update it with new resources
	var availableRes models.AvailableResources
	if res := db.DB.WithContext(c.Request.Context()).First(&availableRes); res.RowsAffected == 0 {
		zlog.Error("availableRes table does not exist")
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "Could not proceed, please check have you onboarded your machine on Nunet"})
		return
	}
	availableRes.TotCpuHz = int(capacityForNunet.CPU)
	availableRes.Ram = int(capacityForNunet.Memory)
	db.DB.Save(&availableRes)

	file, _ := json.MarshalIndent(metadata, "", " ")
	err = AFS.WriteFile(fmt.Sprintf("%s/metadataV2.json", config.GetConfig().General.MetadataPath), file, 0644)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "could not write metadata.json"})
		return
	}
	klogger.Logger.Info("device resource changed")
	telemetry.CalcFreeResources()
	c.JSON(http.StatusOK, metadata)
}

// Offboard      godoc
// @Summary      Runs the offboarding process.
// @Description  Offboard runs the offboarding script to remove resources associated with a device.
// @Tags         onboarding
// @Success      200  "Successfully Onboarded"
// @Router       /onboarding/offboard [delete]
func Offboard(c *gin.Context) {
	klogger.Logger.Info("device offboarding started")
	force, _ := strconv.ParseBool(c.DefaultQuery("force", "false"))
	if onboarded, err := utils.IsOnboarded(); !onboarded {
		if err != nil {
			if !force {
				c.JSON(http.StatusBadRequest, gin.H{"error": "problem with state: " + err.Error()})
				klogger.Logger.Error("offboarding error : " + err.Error())
				return
			} else {
				zlog.Sugar().Errorf("problem with onboarding state: %v", err)
				zlog.Info("continuing with offboarding because forced")
			}
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "machine not onboarded"})
			klogger.Logger.Error("offboarding error : machine not onboarded")
			return
		}
	}

	// remove the metadata file
	err := os.Remove(fmt.Sprintf("%s/metadataV2.json", config.GetConfig().General.MetadataPath))
	if err != nil {
		if !force {
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				gin.H{"error": "failed to delete metadata file"})
			return
		} else {
			zlog.Sugar().Errorf("failed to delete metadata file - problem with onboarding state: %v", err)
			zlog.Info("continuing with offboarding because forced")
		}
	}

	// delete the available resources from database
	var availableRes models.AvailableResources
	result := db.DB.WithContext(c.Request.Context()).Where("id = ?", 1).Delete(&availableRes)
	if result.Error != nil {
		zlog.Error(result.Error.Error())
	} else if result.RowsAffected == 0 {
		zlog.Error("No rows affected")
		if !force {
			return
		}
	}

	telemetry.DeleteCalcFreeResources()
	err = libp2p.ShutdownNode()
	if err != nil && !force {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unable to properly shutdown the node"})
		return
	}
	klogger.Logger.Info("device offboarded successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Successfully Offboarded", "forced": force})
}

func fileExists(filename string) bool {
	info, err := AFS.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func validateCapacityForNunet(capacityForNunet models.CapacityForNunet) error {
	totalCpu := GetTotalProvisioned().CPU
	totalMem := GetTotalProvisioned().Memory

	if capacityForNunet.CPU > int64(totalCpu*9/10) || capacityForNunet.CPU < int64(totalCpu/10) {
		return fmt.Errorf("CPU should be between 10%% and 90%% of the available CPU (%d and %d)", int64(totalCpu/10), int64(totalCpu*9/10))
	}

	if capacityForNunet.Memory > int64(totalMem*9/10) || capacityForNunet.Memory < int64(totalMem/10) {
		return fmt.Errorf("memory should be between 10%% and 90%% of the available memory (%d and %d)", int64(totalMem/10), int64(totalMem*9/10))
	}

	return nil
}
