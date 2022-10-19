package machines

import (
	"encoding/json"
	"net/http"
	"time"

	"context"
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/adapter"
	"gitlab.com/nunet/device-management-service/models"
	"go.opentelemetry.io/otel"
)

// SendDeploymentRequest  godoc
// @Summary      Search devices on DHT with appropriate machines and sends a deployment request.
// @Description  SendDeploymentRequest searches the DHT for non-busy, available devices with appropriate metadata. Then sends a deployment request to the first machine
// @Tags         run
// @Produce      json
// @Success      200  {string}  string
// @Router       /run/deploy [post]
func SendDeploymentRequest(c *gin.Context) {
	// parse the body, get service type, and filter devices
	var depReq models.DeploymentRequest
	if err := c.ShouldBindJSON(&depReq); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "invalid deployment request body",
		})
	}

	// check if the pricing matched
	if estimatedNtx := CalculateStaticNtxGpu(depReq); estimatedNtx > float64(depReq.MaxNtx) {
		c.JSON(406, gin.H{"error": "nunet estimation price is greater than client price"})
		return
	}

	// filter peers based on passed criteria
	peers, err := FilterPeers(c, depReq)
	if err != nil {
		c.JSON(500, gin.H{"error": "no peers with specified spec was found"})
		panic(err)
	}

	//create a new span
	_, span := otel.Tracer("SendDeploymentRequest").Start(context.Background(), "SendDeploymentRequest")
	defer span.End()

	//extract a span context parameters , so that a child span constructed on the reciever side
	depReq.TraceInfo.SpanID = span.SpanContext().TraceID().String()
	depReq.TraceInfo.TraceID = span.SpanContext().SpanID().String()
	depReq.TraceInfo.TraceFlags = span.SpanContext().TraceFlags().String()
	depReq.TraceInfo.TraceStates = span.SpanContext().TraceState().String()

	// pick a peer from the list and send a message to the nodeID of the peer.
	selectedNode := peers[0]
	depReq.Timestamp = time.Now()
	out, err := json.Marshal(depReq)

	if err != nil {
		c.JSON(500, gin.H{"error": "error converting deployment request body to string"})
		panic(err)
	}

	response, err := adapter.SendMessage(selectedNode.PeerInfo.NodeID, string(out))
	if err != nil {
		c.JSON(500, gin.H{"error": "cannot send message to the peer"})
		panic(err)
	}

	c.JSON(200, response)
}

// ListPeers  godoc
// @Summary      Return list of peers currently connected to
// @Description  Gets a list of peers the adapter can see within the network and return a list of peer info
// @Tags         run
// @Produce      json
// @Success      200  {string}	string
// @Router       /peer/list [get]
func ListPeers(c *gin.Context) {
	response, err := adapter.FetchMachines()
	if err != nil {
		c.JSON(500, gin.H{"error": "can not fetch peers"})
		panic(err)
	}
	c.JSON(200, response)

}
