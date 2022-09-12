package gpu

import (
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/adapter"
)

// SearchDevice searches the DHT for non-busy, available devices with "has_gpu" metadata. Search results returns a list of available devices along with the resource capacity.
func SearchDevice(c *gin.Context) {
	jsonBytes, err := adapter.FetchDht()
	if err != nil {
		panic(err)
	}

	var dht adapter.DHT

	err = json.Unmarshal(jsonBytes, &dht)
	if err != nil {
		log.Fatalf("Error unmarshalling data")
	}

	peers := adapter.PeersWithGPU(dht.PeerMeta)
	peers = adapter.PeersNonBusy(peers)

	c.JSON(200, peers)
}

func SendDeploymentRequest(c *gin.Context) {
	// Send message to nodeID with REQUESTING_PEER_PUBKEY
	nodeId := c.Param("nodeID")
	deploymentType := c.Query("deployment_type")

	response, err := adapter.SendMessage(nodeId, deploymentType)
	if err != nil {
		log.Fatalf("Error sending message")
	}

	c.JSON(200, response)
}
