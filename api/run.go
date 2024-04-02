package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	kLogger "gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/libp2p/machines"
	"gitlab.com/nunet/device-management-service/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RequestServiceHandler  godoc
//
//	@Summary		Informs parameters related to blockchain to request to run a service on NuNet
//	@Description	RequestServiceHandler searches the DHT for non-busy, available devices with appropriate metadata. Then informs parameters related to blockchain to request to run a service on NuNet.
//	@Tags			run
//	@Param			deployment_request	body		models.DeploymentRequest	true	"Deployment Request"
//	@Success		200					{object}	machines.fundingRespToSPD
//	@Router			/run/request-service [post]
func RequestServiceHandler(c *gin.Context) {
	if c.Request.ContentLength == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewEmptyBodyProblem())
		return
	}

	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/run/request-service"))
	kLogger.Info("Handle request service", span)

	var depReq models.DeploymentRequest
	err := c.ShouldBindJSON(&depReq)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewValidationProblem(err))
		return
	}

	resp, err := machines.RequestService(c.Request.Context(), depReq)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "failed to request service"})
		return
	}
	c.JSON(200, resp)
}

// DeploymentRequestHandler  godoc
//
//	@Summary		Websocket endpoint responsible for sending deployment request and receiving deployment response.
//	@Description	Loads deployment request from the DB after a successful blockchain transaction has been made and passes it to compute provider.
//	@Tags			run
//	@Success		200	{string}	string
//	@Router			/run/deploy [get]
func DeploymentRequestHandler(c *gin.Context) {
	reqCtx := c.Request.Context()

	span := trace.SpanFromContext(reqCtx)
	span.SetAttributes(attribute.String("URL", "/run/deploy"))
	kLogger.Info("Handle deployment request", span)

	err := machines.DeploymentRequest(c, reqCtx, c.Writer, c.Request)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	// TODO: Original func did not return a success response. Should we return it?
}

// ListCheckpointHandler godoc
//
//	@Summary		Returns a list of absolute path to checkpoint files.
//	@Description	ListCheckpointHandler scans data_dir/received_checkpoints and lists all the tar.gz files which can be used to resume a job. Returns a list of objects with absolute path and last modified date.
//	@Tags			run
//	@Success		200					{object}	[]libp2p.checkpoint
//	@Router			/run/checkpoints [get]
func ListCheckpointHandler(c *gin.Context) {
	checkpoints, err := libp2p.ListCheckpoints()
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "failed to get checkpoint list"})
		return
	}
	c.JSON(200, checkpoints)
}
