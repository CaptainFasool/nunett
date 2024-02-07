package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/internal/klogger"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding"
)

func HandleGetMetadata(c *gin.Context) {
	metadata, err := onboarding.GetMetadata()
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, metadata)
}

func HandleCreatePaymentAddress(c *gin.Context) {
	wallet := c.DefaultQuery("blockchain", "cardano")
	pair, err := onboarding.CreatePaymentAddress(wallet)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, pair)
}

func HandleOnboard(c *gin.Context) {
	capacity := models.CapacityForNunet{
		ServerMode: true,
	}
	err := c.BindJSON(&capacity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	reqCtx := c.Request.Context()
	metadata, err := onboarding.Onboard(reqCtx, capacity)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, metadata)
}

func HandleOffboard(c *gin.Context) {
	force, _ := strconv.ParseBool(c.DefaultQuery("force", "false"))

	reqCtx := c.Request.Context()
	err := onboarding.Offboard(reqCtx, force)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, nil)
}

func HandleOnboardStatus(c *gin.Context) {
	status, err := onboarding.Status()
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, status)
}

func HandleResourceConfig(c *gin.Context) {
	klogger.Logger.Info("device resource change started")
	if c.Request.ContentLength == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body is empty"})
		return
	}

	var capacity models.CapacityForNunet
	err := c.BindJSON(&capacity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	reqCtx := c.Request.Context()
	metadata, err := onboarding.ResourceConfig(reqCtx, capacity)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, metadata)
}
