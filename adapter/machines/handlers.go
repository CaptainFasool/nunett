package machines

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/adapter"
	"gitlab.com/nunet/device-management-service/models"
)

// SendDeploymentRequest  godoc
// @Summary      Search devices on DHT with appropriate machines and sends a deployment request.
// @Description  SendDeploymentRequest searches the DHT for non-busy, available devices with appropriate metadata. Then sends a deployment request to the first machine
// @Tags         run
// @Produce      json
// @Success      200  {array}  adapter.Peer
// @Router       /run/deploy [post]
func SendDeploymentRequest(c *gin.Context) {
	// parse the body, get service type, and filter devices
	var deploymentRequest models.DeploymentRequest
	c.BindJSON(&deploymentRequest)

	peers, err := SearchDevice(c, deploymentRequest.ServiceType)
	if err != nil {
		panic(err)
	}

	// pick a peer from the list and send a message to the nodeID of the peer.
	selectedNode := peers[0]

	out, err := json.Marshal(deploymentRequest)
	if err != nil {
		panic(err)
	}

	response, err := adapter.SendMessage(selectedNode.PeerID.NodeID, string(out))
	if err != nil {
		panic(err)
	}

	c.JSON(200, response)
}

// ReceiveDeploymentRequest  godoc
// @Summary      Receive the deployment message and do the needful.
// @Description  Receives the deployment message from the message exchange. And do required actions based on the service_type.
// @Tags         gpu
// @Produce      json
// @Success      200  {string}	string
// @Router       /run/deploy/receive [get]
func ReceiveDeploymentRequest(c *gin.Context) {
	// The current implementation of the system is to poll message exchange for any
	// new message. This should be replaced with something similar observer pattern
	// where when a message is received, this endpoint is triggered.

	// TODO: Check the service_type and act according to it.
}
