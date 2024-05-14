package api

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/nunet/device-management-service/executor/firecracker"
	"gitlab.com/nunet/device-management-service/models"
)

type CustomVM struct {
	KernelImagePath string `json:"kernel_image_path"`
	FilesystemPath  string `json:"filesystem_path"`
	VCPUCount       int    `json:"vcpu_count"`
	MemSizeMib      int    `json:"mem_size_mib"`
	TapDevice       string `json:"tap_device"`
}

type DefaultVM struct {
	KernelImagePath string `json:"kernel_image_path"`
	FilesystemPath  string `json:"filesystem_path"`
	PublicKey       string `json:"public_key"`
	NodeID          string `json:"node_id"`
}

//	StartCustomHandler godoc
//
// @Summary		Start a VM with custom configuration.
// @Description	This endpoint is an abstraction of all primitive endpoints. When invokend, it calls all primitive endpoints in a sequence.
// @Tags			vm
// @Produce		json
// @Param			body	body		firecracker.CustomVM	true	"body"
// @Success		200		{object}	string					"VM started successfully."
// @Failure		400		{object}	string					"invalid request body"
// @Failure		500		{object}	string					"could not create database table"
// @Failure		500		{object}	string					"could not initialize virtual machine"
// @Failure		500		{object}	string					"failed to configure drives"
// @Failure		500		{object}	string					"failed to configure machine config"
// @Failure		500		{object}	string					"failed to configure network-interfaces"
// @Failure		500		{object}	string					"failed to setup MMDS"
// @Failure		500		{object}	string					"failed to pass MMDS message"
// @Failure		500		{object}	string					"unable to start virtual machine"
// @Router			/vm/start-custom [post]
func StartCustomHandler(c *gin.Context) {
	reqCtx := c.Request.Context()
	span := trace.SpanFromContext(reqCtx)
	span.SetAttributes(attribute.String("URL", "/vm/start-custom"))

	var body CustomVM
	err := c.BindJSON(&body)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid request body"})
		return
	}

	fe := firecracker.NewFirecrackerEngineBuilder(body.FilesystemPath).
		WithKernelImage(body.KernelImagePath).
		Build()

	fer := &models.ExecutionRequest{
		JobID:       "test_job",
		ExecutionID: "test_execution",
		EngineSpec:  fe,
		Resources: &models.ExecutionResources{
			CPU:    float64(body.VCPUCount),
			Memory: uint64(body.MemSizeMib * 1024 * 1024),
		},
	}

	fc, err := firecracker.NewExecutor(c.Request.Context(), "manual-start-custom")
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	err = fc.Start(c.Request.Context(), fer)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "VM started successfully"})
}

// StartDefaultHandler godoc
//
//	@Summary		Start a VM with default configuration.
//	@Description	Kernel file and filesystem file needs to be passed in body. This endpoint is an abstraction of all primitive endpoints.
//	@Tags			vm
//	@Produce		json
//	@Param			body	body		firecracker.DefaultVM	true	"body"
//	@Success		200		{object}	string					"VM started successfully."
//	@Failure		400		{object}	string					"invalid request body"
//	@Failure		500		{object}	string					"could not initialize virtual machine"
//	@Failure		500		{object}	string					"failed to confiugre boot source"
//	@Failure		500		{object}	string					"failed to configure drives"
//	@Failure		500		{object}	string					"failed to configure machineConfig"
//	@Failure		500		{object}	string					"failed to configure network-interfaces"
//	@Failure		500		{object}	string					"failed to setup MMDS"
//	@Failure		500		{object}	string					"failed to pass MMDS message"
//	@Failure		500		{object}	string					"unable to start virtual machine"
//	@Router			/vm/start-default [post]
func StartDefaultHandler(c *gin.Context) {
	reqCtx := c.Request.Context()
	span := trace.SpanFromContext(reqCtx)
	span.SetAttributes(attribute.String("URL", "/vm/start-default"))

	var body DefaultVM
	err := c.BindJSON(&body)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid request body"})
		return
	}

	fe := firecracker.NewFirecrackerEngineBuilder(body.FilesystemPath).
		WithKernelImage(body.KernelImagePath).
		Build()

	fer := &models.ExecutionRequest{
		JobID:       "test_job",
		ExecutionID: "test_execution",
		EngineSpec:  fe,
		Resources: &models.ExecutionResources{
			CPU:    1,
			Memory: 1024 * 1024 * 1024,
		},
	}

	fc, err := firecracker.NewExecutor(c.Request.Context(), "manual-start-default")
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	err = fc.Start(c.Request.Context(), fer)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "VM started successfully"})
}
