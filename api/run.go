package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/libp2p/machines"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type checkpoint struct {
	CheckpointDir string `json:"checkpoint_dir"`
	FilenamePath  string `json:"filename_path"`
	LastModified  int64  `json:"last_modified"`
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
//	@Success		200					{object}	[]checkpoint
//	@Router			/run/checkpoints [get]
func ListCheckpointHandler(c *gin.Context) {
	checkpoints, err := libp2p.ListCheckpoints()
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "failed to get checkpoint list"})
		return
	}
	c.JSON(200, checkpoints)
}
