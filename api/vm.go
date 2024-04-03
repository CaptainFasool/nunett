package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/firecracker"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// StartCustomHandler godoc
//
//	@Summary		Start a VM with custom configuration.
//	@Description	This endpoint is an abstraction of all primitive endpoints. When invokend, it calls all primitive endpoints in a sequence.
//	@Tags			vm
//	@Produce		json
//	@Success		200
//	@Router			/vm/start-custom [post]
func StartCustomHandler(c *gin.Context) {
	if c.Request.ContentLength == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewEmptyBodyProblem())
		return
	}

	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/vm/start-custom"))

	var body firecracker.CustomVM
	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewValidationProblem(err))
		return
	}
	err = firecracker.StartCustom(c.Request.Context(), body)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "VM started successfully"})
}

// StartDefaultHandler godoc
//
//	@Summary		Start a VM with default configuration.
//	@Description	Everything except kernel files and filesystem file will be set by DMS itself.
//	@Tags			vm
//	@Produce		json
//	@Success		200
//	@Router			/vm/start-default [post]
func StartDefaultHandler(c *gin.Context) {
	if c.Request.ContentLength == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewEmptyBodyProblem())
		return
	}

	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/vm/start-default"))

	var body firecracker.DefaultVM
	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewValidationProblem(err))
		return
	}
	err = firecracker.StartDefault(c.Request.Context(), body)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "VM started successfully"})
}
