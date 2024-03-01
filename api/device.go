package api

import (
	"github.com/gin-gonic/gin"
	kLogger "gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DeviceStatusHandler  godoc
//
// @Summary		    Retrieve device status
// @Description	    Retrieve device status whether paused/offline (unable to receive job deployments) or online
// @Tags			device
// @Produce		    json
// @Success		    200	{string}	string
// @Router			/device/status [get]
func DeviceStatusHandler(c *gin.Context) {
	status, err := libp2p.DeviceStatus()
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "could not retrieve device status"})
		return
	}
	c.JSON(200, gin.H{"online": status})
}

// ChangeDeviceStatusHandler  godoc
//
// @Summary		    Change device status between online/offline
// @Description	    Change device status to online (able to receive jobs) or offline (unable to receive jobs).
// @Tags			device
// @Produce		    json
// @Success		    200	{string}	string
// @Router			/device/status [post]
func ChangeDeviceStatusHandler(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/device/pause"))
	span.SetAttributes(attribute.String("MachineUUID", utils.GetMachineUUID()))
	kLogger.Info("Pause job onboarding", span)

	var status struct {
		IsAvailable bool `json:"is_available"`
	}

	if c.Request.ContentLength == 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": "empty content data"})
		return
	}
	err := c.ShouldBindJSON(&status)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid payload data"})
		return
	}

	err = libp2p.ChangeDeviceStatus(status.IsAvailable)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "could not change device status"})
		return
	}

	var msg string
	if status.IsAvailable {
		msg = "Device status successfully changed to online"
	} else {
		msg = "Device status successfully changed to offline"
	}
	c.JSON(200, gin.H{"message": msg})
}
