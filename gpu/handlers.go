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

// SearchDevice  godoc
// @Summary      Search devices on DHT with has_gpu attribute set..
// @Description  SearchDevice searches the DHT for non-busy, available devices with "has_gpu" metadata. Search results returns a list of available devices along with the resource capacity.
// @Tags         gpu
// @Produce      json
// @Success      200  {array}  adapter.Peer
// @Router       /gpu/devices [get]
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

// SendDeploymentRequest  godoc
// @Summary      Send deployment request to one of the peer.
// @Description  Sends deployment request message to one of the peer on the message exchange. Request include details such as docker image name, capacity required etc.
// @Tags         gpu
// @Produce      json
// @Success      200  {string}  string  "sent"
// @Router       /gpu/deploy/:nodeID [post]
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

// ReceiveDeploymentRequest  godoc
// @Summary      Receive the deployment message and do the needful.
// @Description  Receives the deployment message from the message exchange. And do following docker based actions in the sequence: Pull image, rnu container, get logs, delete container, delete image, send log to the requester.
// @Tags         gpu
// @Produce      json
// @Success      200  {string}	string
// @Router       /gpu/deploy/receive [get]
func ReceiveDeploymentRequest(c *gin.Context) {
	// The current implementation of the system is to poll message exchange for any
	// new message. This should be replaced with something similar observer pattern
	// where when a message is received, this endpoint is triggered.
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
	// TODO: Send message to the requesting peer instead of below stub.
	c.JSON(200, logOutput)
}
