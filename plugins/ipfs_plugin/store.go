package ipfs_plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "gitlab.com/nunet/device-management-service/plugins/ipfs_plugin/grpc/ipfs_plugin"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	port = "31001"
)

var (
	gRPCClient pb.IPFSClient
	conn       *grpc.ClientConn
	mu         sync.Mutex
)

func UseSnapshotsFeatIPFS(jobID string, scheduleSec int) {
	// TODO 1: Go routine to store data
	// TODO 2: Go routine to receive CIDs and distribute CIDs
	return
}

func UseOutputFeatIPFS(jobID string) error {
	// Store Data
	ipfsPlug := NewIPFSPlugin(context.Background())
	storeResponse, err := ipfsPlug.storeOutputIPFS(jobID)
	if err != nil {
		zlog.Sugar().Error(err)
		return err
	}

	// Distributing CIDs to topic
	ipfsPlug.ts.Publish(storeResponse.CID)
	return nil
}

func storeSnapshotsIPFS(jobID string, scheduleSec int) {
	return
}

func (p *IPFSPlugin) storeOutputIPFS(jobID string) (*pb.StoreOutputResponse, error) {
	storeResponse, err := storeOutputRPC(jobID)
	if err != nil {
		return nil, err
	}
	zlog.Sugar().Info("Returned CID for output stored on IPFS: %v ", storeResponse.CID)

	return storeResponse, nil
}

func storeOutputRPC(jobID string) (*pb.StoreOutputResponse, error) {
	zlog.Sugar().Infof("Sending gRPC /store call to IPFS-Plugin")
	client, err := newgRPCClient()
	if err != nil {
		zlog.Sugar().Errorf("Fail creating gRPC instance to IPFS-Plugin server: %v", err)
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	storeReq := pb.StoreOutputRequest{
		ContainerId: jobID,
	}

	res, err := client.StoreOutput(ctx, &storeReq)
	if err != nil {
		zlog.Sugar().Errorf("Failed when gRPC calling /store to IPFS-Plugin %v", err)
		return nil, err
	}

	storeRes := &pb.StoreOutputResponse{
		CID: res.GetCID(),
	}

	zlog.Sugar().Infof("Store response: %v", storeRes)
	return storeRes, nil
}

func pinBasedOnCidRPC(cid string) error {
	zlog.Sugar().Infof("Sending gRPC /PinByCID call to IPFS-Plugin")
	client, err := newgRPCClient()
	if err != nil {
		zlog.Sugar().Errorf("Fail creating gRPC instance to IPFS-Plugin server: %v", err)
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	storeReq := pb.PinByCIDRequest{
		CID: cid,
	}

	_, err = client.PinByCID(ctx, &storeReq)
	if err != nil {
		if errStatus, ok := status.FromError(err); ok {
			return fmt.Errorf(
				"(%v) Something went wrong pinning data based on CID, Error message: %v",
				errStatus.Code(), errStatus.Message())
		} else {
			return fmt.Errorf("Error calling gRPC server of IPFS-Plugin, Error: %w", err)
		}
	}

	return nil
}

func newgRPCClient() (pb.IPFSClient, error) {
	mu.Lock()
	defer mu.Unlock()

	if gRPCClient != nil {
		return gRPCClient, nil
	}

	var err error
	conn, err = grpc.Dial(fmt.Sprintf("localhost:%s", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err == nil {
		gRPCClient = pb.NewIPFSClient(conn)
	}

	return gRPCClient, err
}
