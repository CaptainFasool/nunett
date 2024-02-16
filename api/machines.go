package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	kLogger "gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/libp2p/machines"
	"gitlab.com/nunet/device-management-service/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//	 HandleRequestService  godoc
//
//		@Summary		Informs parameters related to blockchain to request to run a service on NuNet
//		@Description	RequestServiceHandler searches the DHT for non-busy, available devices with appropriate metadata. Then informs parameters related to blockchain to request to run a service on NuNet.
//		@Tags			run
//		@Param			deployment_request	body		models.DeploymentRequest	true	"Deployment Request"
//		@Success		200					{object}	fundingRespToSPD
//		@Router			/run/request-service [post]
func HandleRequestService(c *gin.Context) {
	reqCtx := c.Request.Context()

	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/run/request-service"))
	kLogger.Info("Handle request service", span)

	// receive deployment request
	var depReq models.DeploymentRequest
	err := c.BindJSON(&depReq)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Errorf("invalid payload data for deployment request: %w", err)})
		return
	}
	resp, err := machines.RequestService(reqCtx, depReq)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Errorf("failed to request service: %w", err)})
		return
	}
	c.JSON(200, resp)
}

// HandleDeploymentRequest  godoc
//
//	@Summary		Websocket endpoint responsible for sending deployment request and receiving deployment response.
//	@Description	Loads deployment request from the DB after a successful blockchain transaction has been made and passes it to compute provider.
//	@Tags			run
//	@Success		200	{string}	string
//	@Router			/run/deploy [get]
func HandleDeploymentRequest(c *gin.Context) {
	reqCtx := c.Request.Context()

	span := trace.SpanFromContext(reqCtx)
	span.SetAttributes(attribute.String("URL", "/run/deploy"))
	kLogger.Info("Handle deployment request", span)

	err := machines.DeploymentRequest(c, reqCtx, c.Writer, c.Request)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	// TODO: Original func did not return a success response. Should we return it?
}
