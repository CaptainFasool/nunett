package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/firecracker"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// HandleStartCustom godoc
//
//	@Summary		Start a VM with custom configuration.
//	@Description	This endpoint is an abstraction of all primitive endpoints. When invokend, it calls all primitive endpoints in a sequence.
//	@Tags			vm
//	@Produce		json
//	@Success		200
//	@Router			/vm/start-custom [post]
func HandleStartCustom(c *gin.Context) {
	reqCtx := c.Request.Context()
	span := trace.SpanFromContext(reqCtx)
	span.SetAttributes(attribute.String("URL", "/vm/start-custom"))

	var body firecracker.CustomVM
	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	err = firecracker.StartCustom(reqCtx, body)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "VM started successfully"})
}

// HandleStartDefault godoc
//
//	@Summary		Start a VM with default configuration.
//	@Description	Everything except kernel files and filesystem file will be set by DMS itself.
//	@Tags			vm
//	@Produce		json
//	@Success		200
//	@Router			/vm/start-default [post]
func HandleStartDefault(c *gin.Context) {
	reqCtx := c.Request.Context()
	span := trace.SpanFromContext(reqCtx)
	span.SetAttributes(attribute.String("URL", "/vm/start-default"))

	var body firecracker.DefaultVM
	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	err = firecracker.StartDefault(reqCtx, body)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "VM started successfully"})
}
