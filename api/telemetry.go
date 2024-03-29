package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/dms/resources"
)

// GetFreeResourcesHandler godoc
//
//	@Summary		Returns the amount of free resources available
//	@Description	Checks and returns the amount of free resources available
//	@Tags			telemetry
//	@Produce		json
//	@Success		200
//	@Router			/telemetry/free [get]
func GetFreeResourcesHandler(c *gin.Context) {
	reqCtx := c.Request.Context()
	free, err := resources.GetFreeResource(reqCtx)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, free)
}
