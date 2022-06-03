package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"

	"device-management-service/models"

	"github.com/gin-gonic/gin"
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

// https://github.com/swaggo/swag
