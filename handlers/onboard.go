package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"

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

// Onboard runs onboarding script given the amount of resources to onboard.
func Onboard(c *gin.Context) {
	out, err := exec.Command("ls", "/home").Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": string(out)})
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
// @Summary      Returns provisined capacity on host.
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
