package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/dms/onboarding"
	"gitlab.com/nunet/device-management-service/dms/resources"
	"gitlab.com/nunet/device-management-service/internal/klogger"
	"gitlab.com/nunet/device-management-service/models"
)

// ProvisionedCapacityHandler      godoc
//
//	@Summary		Returns provisioned capacity on host.
//	@Description	Get total memory capacity in MB and CPU capacity in MHz.
//	@Tags			onboarding
//	@Produce		json
//	@Success		200	{object}	models.Provisioned
//	@Router			/onboarding/provisioned [get]
func ProvisionedCapacityHandler(c *gin.Context) {
	c.JSON(200, resources.GetTotalProvisioned())
}

// GetMetadataHandler      godoc
//
//	@Summary		Get current device info.
//	@Description	Responds with metadata of current provideer
//	@Tags			onboarding
//	@Produce		json
//	@Success		200	{object}	models.Metadata
//	@Router			/onboarding/metadata [get]
func GetMetadataHandler(c *gin.Context) {
	metadata, err := onboarding.GetMetadata()
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, metadata)
}

// CreatePaymentAddressHandler      godoc
//
//	@Summary		Create a new payment address.
//	@Description	Create a payment address from public key. Return payment address and private key.
//	@Tags			onboarding
//	@Produce		json
//	@Success		200	{object}	models.BlockchainAddressPrivKey
//	@Router			/onboarding/address/new [get]
func CreatePaymentAddressHandler(c *gin.Context) {
	wallet := c.DefaultQuery("blockchain", "cardano")
	pair, err := onboarding.CreatePaymentAddress(wallet)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, pair)
}

// OnboardHandler      godoc
//
//	@Summary		Runs the onboarding process.
//	@Description	Onboard runs onboarding script given the amount of resources to onboard.
//	@Tags			onboarding
//	@Produce		json
//	@Success		200	{object}	models.Metadata
//	@Router			/onboarding/onboard [post]
func OnboardHandler(c *gin.Context) {
	capacity := models.CapacityForNunet{
		ServerMode:  true,
		IsAvailable: true,
	}
	err := c.ShouldBindJSON(&capacity)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewValidationProblem(err))
		return
	}

	reqCtx := c.Request.Context()
	metadata, err := onboarding.Onboard(reqCtx, capacity)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, metadata)
}

// OnboardStatusHandler      godoc
//
//	@Summary		Onboarding status and other metadata.
//	@Description	Returns json with 5 parameters: onboarded, error, machine_uuid, metadata_path, database_path.
//	@Description	`onboarded` is true if the device is onboarded, false otherwise.
//	@Description	`error` is the error message if any related to onboarding status check
//	@Description	`machine_uuid` is the UUID of the machine
//	@Description	`metadata_path` is the path to metadataV2.json only if it exists
//	@Description	`database_path` is the path to nunet.db only if it exists
//	@Tags			onboarding
//	@Produce		json
//	@Success		200	{object} models.OnboardingStatus
//	@Router			/onboarding/status [get]
func OnboardStatusHandler(c *gin.Context) {
	status, err := onboarding.Status()
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, status)
}

// OffboardHandler      godoc
//
//	@Summary		Runs the offboarding process.
//	@Description	Offboard runs offboarding process to remove the machine from the NuNet network.
//	@Tags		onboarding
//	@Produce		json
//	@Success      200              {string}  string    "device successfully offboarded"
//	@Router		/onboarding/offboard [post]
func OffboardHandler(c *gin.Context) {
	type offboardQuery struct {
		Force bool `form:"force" binding:"omitempty,boolean"`
	}

	var query offboardQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewValidationProblem(err))
		return
	}

	err = onboarding.Offboard(c.Request.Context(), query.Force)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "device successfully offboarded"})
}

// ResourceConfigHandler        godoc
//
//	@Summary	changes the amount of resources of onboarded device
//	@Tags		onboarding
//	@Produce	json
//	@Success	200	{object}	models.Metadata
//	@Router		/onboarding/resource-config [post]
func ResourceConfigHandler(c *gin.Context) {
	if c.Request.ContentLength == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewEmptyBodyProblem())
		return
	}

	klogger.Logger.Info("device resource change started")

	var resource models.ResourceConfig
	err := c.ShouldBindJSON(&resource)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewValidationProblem(err))
		return
	}

	metadata, err := onboarding.ResourceConfig(c.Request.Context(), resource)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, metadata)
}
