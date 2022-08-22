package spo

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	pb "gitlab.com/nunet/device-management-service/adapter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SearchDevice searches the DHT for non-busy, available devices with "allow_cardano" metadata. Search results returns a list of available devices along with the resource capacity.
func SearchDevice(c *gin.Context) {
	// Set up a connection to the server.
	address := "localhost:9998"
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewNunetAdapterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := client.GetPeerList(ctx, &pb.GetPeerParams{})
	if err != nil {
		log.Fatalf("could not add: %v", err)
	}
	log.Printf("Add result: %v", r)

}

func DeployAuto(c *gin.Context) {

}

func DeployManual(c *gin.Context) {

}
