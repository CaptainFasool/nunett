package ipfs_plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "gitlab.com/nunet/device-management-service/plugins/ipfs_plugin/grpc/ipfs_plugin"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	port = "31001"
)

var (
	gRPCClient pb.IPFSPluginClient
	conn       *grpc.ClientConn
	mu         sync.Mutex
)

func UseSnapshotsFeatIPFS(jobID string, scheduleSec int) {
	// TODO 1: Go routine to store data
	// TODO 2: Go routine to receive CIDs and distribute CIDs
	return
}

func UseOutputFeatIPFS(jobID string) {
	// TODO 1: Store Data
	err := storeOutputIPFS(jobID)
	if err != nil {
		zlog.Sugar().Error(err)
	}

	// TODO 2: Distribute CID
}

func storeSnapshotsIPFS(jobID string, scheduleSec int) {
	return
}

func storeOutputIPFS(jobID string) error {
	storeResponse, err := store(jobID)
	if err != nil {
		return err
	}

	zlog.Sugar().Info("Returned CID for output stored on IPFS: %v ", storeResponse.CID)
	// TODO: distribute CID

	return nil
}

func store(jobID string) (pb.StoreResponse, error) {
	zlog.Sugar().Infof("Sending gRPC /store call to IPFS-Plugin")
	client, err := newgRPCClient()
	if err != nil {
		zlog.Sugar().Errorf("Fail creating gRPC instance to IPFS-Plugin server: %v", err)
		return pb.StoreResponse{}, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	storeReq := pb.StoreRequest{
		ContainerId: jobID,
	}

	res, err := client.Store(ctx, &storeReq)
	if err != nil {
		zlog.Sugar().Errorf("Failed when gRPC calling /store to IPFS-Plugin %v", err)
		return pb.StoreResponse{}, err
	}

	storeRes := pb.StoreResponse{
		CID: res.GetCID(),
	}

	zlog.Sugar().Infof("Store response: %v", storeRes)
	return storeRes, nil
}

func newgRPCClient() (pb.IPFSPluginClient, error) {
	mu.Lock()
	defer mu.Unlock()

	if gRPCClient != nil {
		return gRPCClient, nil
	}

	var err error
	conn, err = grpc.Dial(fmt.Sprintf("localhost:%s", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err == nil {
		gRPCClient = pb.NewIPFSPluginClient(conn)
	}

	return gRPCClient, err
}
