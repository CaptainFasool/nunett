package spo

import (
	"encoding/json"
	"log"

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
		log.Fatalf("Error unmarshalling data")
	}

	peers := adapter.PeersWithCardanoAllowed(dht.PeerMeta)
	peers = adapter.PeersNonBusy(peers)

	c.JSON(200, peers)
}

func DeployAuto(c *gin.Context) {

}

func DeployManual(c *gin.Context) {

}
