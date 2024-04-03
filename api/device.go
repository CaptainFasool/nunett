package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	kLogger "gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DeviceStatusHandler  godoc
//
// @Summary			Retrieve device status
// @Description		Retrieve device status whether paused/offline (unable to receive job deployments) or online
// @Tags			device
// @Produce			json
// @Success			200	{string}	string
// @Router			/device/status [get]
func DeviceStatusHandler(c *gin.Context) {
	status, err := libp2p.DeviceStatus()
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "could not retrieve device status"})
		return
	}
	var payload struct {
		IsAvailable bool `json:"is_available" binding:"boolean"`
	}
	payload.IsAvailable = status
	c.JSON(200, gin.H{"device": payload})
}

// ChangeDeviceStatusHandler  godoc
//
// @Summary			Change device status between online/offline
// @Description		Change device status to online (able to receive jobs) or offline (unable to receive jobs).
// @Tags			device
// @Produce			json
// @Success			200	{string}	string
// @Router			/device/status [post]
func ChangeDeviceStatusHandler(c *gin.Context) {
	if c.Request.ContentLength == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewEmptyBodyProblem())
		return
	}

	type deviceStatus struct {
		IsAvailable bool `json:"is_available" binding:"required,boolean"`
	}
	var status deviceStatus
	err := c.ShouldBindJSON(&status)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewValidationProblem(err))
		return
	}

	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/device/status"))
	span.SetAttributes(attribute.String("MachineUUID", utils.GetMachineUUID()))
	kLogger.Info("Pause job onboarding", span)

	err = libp2p.ChangeDeviceStatus(status.IsAvailable)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"device": status})
}
