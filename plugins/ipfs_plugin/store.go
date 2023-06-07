package ipfs_plugin

import (
	"context"
	"fmt"
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
)

func StoreSnapshotsIPFS(jobID string, scheduleSec int) {
	// TODO: Call this from DMS when the job wants it
	return
}

func StoreOutputIPFS(jobID string) {
	// TODO: Call this from DMS when the job wants it
	return
}

func store() (pb.StoreResponse, error) {
	zlog.Sugar().Infof("Sending gRPC /store call to IPFS-Plugin")
	client, err := newgRPCClient()
	if err != nil {
		return pb.StoreResponse{}, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	storeReq := pb.StoreRequest{
		ContainerId: "test321",
	}

	res, err := client.Store(ctx, &storeReq)
	if err != nil {
		zlog.Sugar().Infof("Failed when gRPC calling /store to IPFS-Plugin %v", err)
		return pb.StoreResponse{}, err
	}

	storeRes := pb.StoreResponse{
		CID: res.GetCID(),
	}

	zlog.Sugar().Infof("Store response: %v", storeRes)
	return storeRes, nil
}

func newgRPCClient() (pb.IPFSPluginClient, error) {
	var err error
	if gRPCClient != nil {
		return gRPCClient, nil
	}

	conn, err = grpc.Dial(fmt.Sprintf("localhost:%s", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	gRPCClient = pb.NewIPFSPluginClient(conn)
	return gRPCClient, nil
}
