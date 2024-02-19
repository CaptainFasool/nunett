package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/telemetry"
)

//	 GetFreeResources godoc
//
//		@Summary		Returns the amount of free resources available
//		@Description	Checks and returns the amount of free resources available
//		@Tags			telemetry
//		@Produce		json
//		@Success		200
//		@Router			/telemetry/free [get]
func HandleGetFreeResources(c *gin.Context) {
	reqCtx := c.Request.Context()
	free, err := telemetry.GetFreeResource(reqCtx)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, free)
}
