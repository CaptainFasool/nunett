package spo

import (
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/adapter"
)

// SearchDevice  godoc
// @Summary      Search devices on DHT with allow_cardano attribute set.
// @Description  SearchDevice searches the DHT for non-busy, available devices with "allow_cardano" metadata. Search results returns a list of available devices along with the resource capacity.
// @Tags         spo
// @Produce      json
// @Success      200  {array}  adapter.Peer
// @Router       /spo/devices [get]
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

	peers := adapter.PeersWithCardanoAllowed(dht.PeerMeta)
	peers = adapter.PeersNonBusy(peers)

	c.JSON(200, peers)
}

// SendDeploymentRequest  godoc
// @Summary      Send deployment request to one of the peer.
// @Description  Sends deployment request message to one of the peer on the message exchange.
// @Tags         spo
// @Produce      json
// @Success      200  {string}	string	"sent"
// @Router       /spo/deploy/:nodeID [post]
func SendDeploymentRequest(c *gin.Context) {
	// Send message to nodeID with REQUESTING_PEER_PUBKEY
	nodeId := c.Param("nodeID")
	// auto: will use a cardano firecracker golden image and takes in configuration parameters.
	// manual: will use a generic ubuntu firecracker golden image with docker installed in it to allow
	// the SPO to remotely connect and setup a cardano node with docker inside firecracker.
	deploymentType := c.Query("deployment_type")

	response, err := adapter.SendMessage(nodeId, deploymentType)
	if err != nil {
		log.Fatalf("Error sending message")
	}

	c.JSON(200, response)
}
