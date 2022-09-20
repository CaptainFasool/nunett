package gpu

import (
	"context"
	"encoding/json"
	"io"
	"log"

	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/adapter"
)

// SearchDevice searches the DHT for non-busy, available devices with "has_gpu" metadata. Search results returns a list of available devices along with the resource capacity.
// TODO: Add Swagger comments
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

// TODO: Add Swagger comments
func SendDeploymentRequest(c *gin.Context) {
	// Send message to nodeID with REQUESTING_PEER_PUBKEY
	nodeId := c.Param("nodeID")

	bodyAsByteArray, _ := io.ReadAll(c.Request.Body)
	jsonBody := string(bodyAsByteArray)

	response, err := adapter.SendMessage(nodeId, jsonBody)

	if err != nil {
		log.Fatalf("Error sending message")
	}

	c.JSON(200, response)
}

// TODO: Add Swagger comments
// The current implementation of the system is to poll message exchange for any
// new message. This should be replaced with something similar observer pattern
// where when a message is received, this endpoint is triggered.
func ReceiveDeploymentRequest(c *gin.Context) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Pull the image
	// TODO: Where in the flow do we get the image name.
	imageName := "nvidia/cuda:10.0-base"

	PullImage(ctx, cli, imageName)

	// Run the container.
	// TODO: What command do we run inside container? Do does the Image comes pre-baked?
	contID := RunContainer(ctx, cli, imageName, []string{"nvidia-smi"})

	// Get the logs.
	logOutput := GetLogs(ctx, cli, contID)

	// Delete the container.
	DeleteContainer(ctx, cli, contID)

	// Remove the downloaded image.
	DeleteImage(ctx, cli, imageName)

	// Send back the logs.
	c.JSON(200, logOutput)
}
