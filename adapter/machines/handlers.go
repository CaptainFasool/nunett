package machines

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/adapter"
	"gitlab.com/nunet/device-management-service/models"
	"go.opentelemetry.io/otel"
	"context"
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
	var deploymentRequest models.DeploymentRequest
	
	c.BindJSON(&deploymentRequest)

	peers, err := SearchDevice(c, deploymentRequest.ServiceType)
	if err != nil {
		c.JSON(500, gin.H{"error": "no peers with specified spec was found"})
		panic(err)
	}

	//create a new span
	_, span := otel.Tracer("SendDeploymentRequest").Start(context.Background(), "SendDeploymentRequest")
	defer span.End()

	//extract a span context parameters , so that a child span constructed on the reciever side 
	deploymentRequest.TraceInfo.SpanID=span.SpanContext().TraceID().String()
	deploymentRequest.TraceInfo.TraceID=span.SpanContext().SpanID().String()
	deploymentRequest.TraceInfo.TraceFlags=span.SpanContext().TraceFlags().String()
	deploymentRequest.TraceInfo.TraceStates=span.SpanContext().TraceState().String()

	// pick a peer from the list and send a message to the nodeID of the peer.
	selectedNode := peers[0]
	deploymentRequest.Timestamp = time.Now()
	out, err := json.Marshal(deploymentRequest)
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

// ReceiveDeploymentRequest  godoc
// @Summary      Receive the deployment message and do the needful.
// @Description  Receives the deployment message from the message exchange. And do required actions based on the service_type.
// @Tags         run
// @Produce      json
// @Success      200  {string}	string
// @Router       /run/deploy/receive [get]
func ReceiveDeploymentRequest() {
	// The current implementation of the system is to poll message exchange for any
	// new message. This should be replaced with something similar observer pattern
	// where when a message is received, this endpoint is triggered.

	adapter.PollAdapter()
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
