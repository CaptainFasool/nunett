package spo

import (
	"encoding/json"
	"net/http"
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/adapter"
)

// SearchDevice searches the DHT for non-busy, available devices with "allow_cardano" metadata. Search results returns a list of available devices along with the resource capacity.
func SearchDevice(c *gin.Context) {
	jsonBytes, err := adapter.FetchDht()
	if err != nil {
		panic(err)
	}

	var dht adapter.DHT

	err = json.Unmarshal(jsonBytes, &dht)
	if err != nil {
		panic("Error unmarshalling data")
	}

	peers := adapter.PeersWithCardanoAllowed(dht.PeerMeta)
	peers = adapter.PeersNonBusy(peers)

	c.JSON(200, peers)
}

// auto: will use a cardano firecracker golden image and takes in configuration parameters.
// manual: will use a generic ubuntu firecracker golden image with docker installed in it to allow
// the SPO to remotely connect and setup a cardano node with docker inside firecracker.
func SendDeploymentRequest(c *gin.Context) {
	// Send message to nodeID with REQUESTING_PEER_PUBKEY
	nodeId := c.Param("nodeID")

	type DeploymentBody struct {
		DeploymentType string `json:"deployment_type,omitempty"`
	}

	body := DeploymentBody{}

	// deploymentType := c.Query("deployment_type")
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	response, err := adapter.SendMessage(nodeId, body.DeploymentType)
	if err != nil {
		panic("Error sending message")
	}

	c.JSON(200, response)
}
