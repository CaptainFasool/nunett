package statsdb

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"gitlab.com/nunet/device-management-service/firecracker/telemetry"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	pb "gitlab.com/nunet/device-management-service/statsdb/event_listener_spec"
	"gitlab.com/nunet/device-management-service/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getAddress() string {
	channelName := utils.GetChannelName()
	var (
		addr string

		// statsdb Address
		nunetStagingAddr string = "stats-event-listener.staging.nunet.io:80"
		nunetTestAddr    string = "stats-event-listener.test.nunet.io:80"
		nunetEdgeAddr    string = "stats-event-listener.edge.nunet.io:80"
		nunetTeamAddr    string = "stats-event-listener.team.nunet.io:80"
		nunetDevAddr     string = "stats-event-listener.dev.nunet.io:80"
		localAddr        string = "127.0.0.1:50051"
	)

	if channelName == "nunet-staging" {
		addr = nunetStagingAddr
	} else if channelName == "nunet-test" {
		addr = nunetTestAddr
	} else if channelName == "nunet-edge" {
		addr = nunetEdgeAddr
	} else if channelName == "nunet-team" {
		addr = nunetTeamAddr
	} else if channelName == "" { // XXX -- setting empty(not yet onboarded) to dev endpoint
		addr = nunetDevAddr
	} else {
		addr = localAddr
	}
	return addr
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GetCallID returns a call ID to track the deployement request
func GetCallID() int {
	randFloat := rand.Float64()
	return int(randFloat * 1e12)
}

// GetTimestamp returns current unix timestamp
func GetTimestamp() int64 {
	return time.Now().Unix()
}

// GetPeerID read PeerID from metadataV2.json maybe it's not equivalent to libp2p.GetP2P().Host.ID().Pretty()
func GetPeerID() (string, error) {
	metadata, err := utils.ReadMetadataFile()
	if err != nil {
		zlog.Sugar().Errorf("could not read metadata: %v", err)
		return "", fmt.Errorf("could not read metadata: %v", err)
	}

	metadata.NodeID = libp2p.GetP2P().Host.ID().Pretty()
	file, _ := json.MarshalIndent(metadata, "", " ")
	err = os.WriteFile("/etc/nunet/metadataV2.json", file, 0644)
	if err != nil {
		zlog.Sugar().Errorf("couldn't write metadata file: %v", err)
		return "", fmt.Errorf("couldn't write metadata file: %v", err)
	}

	return metadata.NodeID, nil
}

// NewDeviceOnboarded sends the newly onboarded telemetry info to the stats db via grpc call.
func NewDeviceOnboarded(inputData models.NewDeviceOnboarded) {
	statsdbConn, err := grpc.Dial(getAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zlog.Sugar().Errorf("did not connect: %v", err)
		return
	}

	statsdbClient := pb.NewEventListenerClient(statsdbConn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res, err := statsdbClient.NewDeviceOnboarded(ctx, &pb.NewDeviceOnboardedInput{
		PeerId:        inputData.PeerID,
		Cpu:           inputData.CPU,
		Ram:           inputData.RAM,
		Network:       inputData.Network,
		DedicatedTime: inputData.DedicatedTime,
		Timestamp:     inputData.Timestamp,
	})

	if err != nil {
		zlog.Sugar().Errorf("connection failed: %v", err)
		return
	}
	zlog.Sugar().Infof("NewDeviceOnboarded is Responding: %s", res.PeerId)
}

// ServiceCall sends the info of the service call made to dms to stats via grpc call.
func ServiceCall(inputData models.ServiceCall) {
	statsdbConn, err := grpc.Dial(getAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zlog.Sugar().Errorf("did not connect: %v", err)
		return
	}

	statsdbClient := pb.NewEventListenerClient(statsdbConn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res, err := statsdbClient.ServiceCall(ctx, &pb.ServiceCallInput{
		CallId:              inputData.CallID,
		PeerIdOfComputeHost: libp2p.GetP2P().Host.ID().Pretty(),
		PeerIdOfServiceHost: inputData.PeerIDOfServiceHost,
		ServiceId:           inputData.ServiceID,
		CpuUsed:             inputData.CPUUsed, MaxRam: inputData.MaxRAM,
		MemoryUsed:    inputData.MemoryUsed,
		NetworkBwUsed: inputData.NetworkBwUsed,
		TimeTaken:     inputData.TimeTaken,
		Status:        inputData.Status,
		AmountOfNtx:   inputData.AmountOfNtx,
		Timestamp:     inputData.Timestamp,
	})

	if err != nil {
		zlog.Sugar().Errorf("connection failed: %v", err)
		return
	}
	zlog.Sugar().Infof("ServiceCall is Responding: %s", res.Response)

}

// ServiceStatus updates the status of service process on host machine to stats db via gRPC call
func ServiceStatus(inputData models.ServiceStatus) {
	statsdbConn, err := grpc.Dial(getAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zlog.Sugar().Errorf("did not connect: %v", err)
		return
	}

	statsdbClient := pb.NewEventListenerClient(statsdbConn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res, err := statsdbClient.ServiceStatus(ctx, &pb.ServiceStatusInput{
		CallId:              inputData.CallID,
		PeerIdOfServiceHost: inputData.PeerIDOfServiceHost,
		ServiceId:           inputData.ServiceID,
		Status:              inputData.Status,
		Timestamp:           inputData.Timestamp,
	})

	if err != nil {
		zlog.Sugar().Errorf("connection failed: %v", err)
		return
	}
	zlog.Sugar().Infof("ServiceStatus is Responding: %s", res.Response)
}

// HeartBeat pings the statsdb in every 10s for detacting live status of device via grpc call.
func HeartBeat(reOnboard bool) {
	peerID, err := GetPeerID()
	if err != nil {
		zlog.Sugar().Errorf("couldn't get PeerID: %v", err)
		return
	}
	if reOnboard {
		freeResource, err := telemetry.GetFreeResources()
		if err != nil {
			zlog.Sugar().Errorf("Error getting freeResources: %v", err)
		}
		DeviceResourceChange(freeResource)
	}
	addr := getAddress()
	for {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			zlog.Sugar().Errorf("did not connect: %v", err)
			return
		}

		client := pb.NewEventListenerClient(conn)

		res, err := client.HeartBeat(context.Background(), &pb.HeartBeatInput{
			PeerId: peerID,
		})

		if err != nil {
			zlog.Sugar().Errorf("connection failed: %v", err)
			return
		}
		zlog.Sugar().Infof("HeartBeat is Responding: %s", res.PeerId)

		time.Sleep(5 * time.Minute)
	}
}

// DeviceResourceChange sends the reonboarding info with new data to statsdb via grpc call.
func DeviceResourceChange(inputData models.FreeResources) {
	peerID, err := GetPeerID()
	if err != nil {
		zlog.Sugar().Errorf("couldn't get PeerID: %v", err)
		return
	}
	DeviceResourceParams := pb.DeviceResource{
		Cpu:           float32(inputData.TotCpuHz),
		Ram:           float32(inputData.Ram),
		Network:       0.0,
		DedicatedTime: 0.0,
	}

	statsdbConn, err := grpc.Dial(getAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zlog.Sugar().Errorf("did not connect: %v", err)
		return
	}

	statsdbClient := pb.NewEventListenerClient(statsdbConn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res, err := statsdbClient.DeviceResourceChange(ctx, &pb.DeviceResourceChangeInput{
		PeerId:                   peerID,
		ChangedAttributeAndValue: &DeviceResourceParams,
		Timestamp:                float32(GetTimestamp()),
	})

	if err != nil {
		zlog.Sugar().Errorf("connection failed: %v", err)
		return
	}

	zlog.Sugar().Infof("DeviceResourceChange is Responding: %s", res.PeerId)
}

// DeviceResourceConfig sends the info of change the configuration resources of onboarded device via GRPC call.
func DeviceResourceConfig(inputData models.MetadataV2) {
	peerID, err := GetPeerID()
	if err != nil {
		zlog.Sugar().Errorf("couldn't get PeerID: %v", err)
		return
	}
	DeviceResourceParams := pb.DeviceResource{
		Cpu:           float32(inputData.Reserved.CPU),
		Ram:           float32(inputData.Reserved.Memory),
		Network:       0.0,
		DedicatedTime: 0.0,
	}

	statsdbConn, err := grpc.Dial(getAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zlog.Sugar().Errorf("did not connect: %v", err)
		return
	}

	statsdbClient := pb.NewEventListenerClient(statsdbConn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res, err := statsdbClient.DeviceResourceConfig(ctx, &pb.DeviceResourceConfigInput{
		PeerId:                   peerID,
		ChangedAttributeAndValue: &DeviceResourceParams,
		Timestamp:                float32(GetTimestamp()),
	})

	if err != nil {
		zlog.Sugar().Errorf("connection failed: %v", err)
		return
	}

	zlog.Sugar().Infof("DeviceResourceConfig is Responding: %s", res.Response)
}

// NtxPayment sends the payment info of the service process on host machine to statsdb via grpc call.
func NtxPayment(inputData models.NtxPayment) {
	statsdbConn, err := grpc.Dial(getAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zlog.Sugar().Errorf("did not connect: %v", err)
		return
	}

	statsdbClient := pb.NewEventListenerClient(statsdbConn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res, err := statsdbClient.NtxPayment(ctx, &pb.NtxPaymentInput{
		CallId:            inputData.CallID,
		ServiceId:         inputData.ServiceID,
		AmountOfNtx:       int32(inputData.AmountOfNtx),
		PeerId:            inputData.PeerID,
		SuccessFailStatus: inputData.SuccessFailStatus,
		Timestamp:         inputData.Timestamp,
	})

	if err != nil {
		zlog.Sugar().Errorf("connection failed: %v", err)
		return
	}
	zlog.Sugar().Infof("NtxPayment is Responding: %s", res.Response)
}
