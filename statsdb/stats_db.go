package statsdb

import (
	"context"
	"math/rand"
	"time"

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
func GetCallID() float32 {
	return rand.Float32()
}

// GetTimestamp returns current unix timestamp
func GetTimestamp() int64 {
	return time.Now().Unix()
}

// NewDeviceOnboarded sends the newly onboarded telemetry info to the stats db via grpc call.
func NewDeviceOnboarded(inputData models.NewDeviceOnboarded) {
	conn, err := grpc.Dial(getAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zlog.Sugar().Errorf("did not connect: %v", err)
	}

	client := pb.NewEventListenerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res, err := client.NewDeviceOnboarded(ctx, &pb.NewDeviceOnboardedInput{
		PeerId:        inputData.PeerID,
		Cpu:           inputData.CPU,
		Ram:           inputData.RAM,
		Network:       inputData.Network,
		DedicatedTime: inputData.DedicatedTime,
		Timestamp:     inputData.Timestamp,
	})

	if err != nil {
		zlog.Sugar().Errorf("connection failed: %v", err)
	}
	zlog.Sugar().Infof("Responding: %s", res.PeerId)
}

// ServiceCall sends the info of the service call made to dms to stats via grpc call.
func ServiceCall(inputData models.ServiceCall) {
	conn, err := grpc.Dial(getAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zlog.Sugar().Errorf("did not connect: %v", err)
	}

	client := pb.NewEventListenerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res, err := client.ServiceCall(ctx, &pb.ServiceCallInput{
		CallId:              inputData.CallID,
		PeerIdOfServiceHost: inputData.PeerIDOfServiceHost,
		ServiceId:           inputData.ServiceID,
		CpuUsed:             inputData.CPUUsed, MaxRam: inputData.MaxRAM,
		MemoryUsed:    inputData.MemoryUsed,
		NetworkBwUsed: inputData.NetworkBwUsed,
		TimeTaken:     inputData.TimeTaken,
		Status:        inputData.Status,
		Timestamp:     inputData.Timestamp,
	})

	if err != nil {
		zlog.Sugar().Errorf("connection failed: %v", err)
	}
	zlog.Sugar().Infof("Responding: %s", res.Response)

}

// ServiceStatus updates the status of service process on host machine to stats db via gRPC call
func ServiceStatus(inputData models.ServiceStatus) {
	conn, err := grpc.Dial(getAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zlog.Sugar().Errorf("did not connect: %v", err)
	}

	client := pb.NewEventListenerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res, err := client.ServiceStatus(ctx, &pb.ServiceStatusInput{
		CallId:              inputData.CallID,
		PeerIdOfServiceHost: inputData.PeerIDOfServiceHost,
		ServiceId:           inputData.ServiceID,
		Status:              inputData.Status,
		Timestamp:           inputData.Timestamp,
	})

	if err != nil {
		zlog.Sugar().Errorf("connection failed: %v", err)
	}
	zlog.Sugar().Infof("Responding: %s", res.Response)
}

// HeartBeat pings the statsdb in every 10s for detacting live status of device via grpc call.
func HeartBeat(peerID string) {
	for {
		conn, err := grpc.Dial(getAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			zlog.Sugar().Errorf("did not connect: %v", err)
		}

		client := pb.NewEventListenerClient(conn)

		res, err := client.HeartBeat(context.Background(), &pb.HeartBeatInput{
			PeerId: peerID,
		})

		if err != nil {
			zlog.Sugar().Errorf("connection failed: %v", err)
		}
		zlog.Sugar().Infof("Responding: %s", res.PeerId)

		time.Sleep(10 * time.Second)
	}

}

// DeviceResourceChange sends the reonboarding info with new data to statsdb via grpc call.
func DeviceResourceChange(inputData *models.MetadataV2) {

	DeviceResourceParams := pb.DeviceResource{
		Cpu:           float32(inputData.Reserved.CPU),
		Ram:           float32(inputData.Reserved.Memory),
		Network:       0.0,
		DedicatedTime: 0.0,
	}

	conn, err := grpc.Dial(getAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zlog.Sugar().Errorf("did not connect: %v", err)
	}

	client := pb.NewEventListenerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res, err := client.DeviceResourceChange(ctx, &pb.DeviceResourceChangeInput{
		PeerId:                   inputData.NodeID,
		ChangedAttributeAndValue: &DeviceResourceParams,
		Timestamp:                float32(GetTimestamp()),
	})

	if err != nil {
		zlog.Sugar().Errorf("connection failed: %v", err)
	}

	zlog.Sugar().Infof("Responding: %s", res.PeerId)
}

// NtxPayment sends the payment info of the service process on host machine to statsdb via grpc call.
func NtxPayment(inputData models.NtxPayment) {
	conn, err := grpc.Dial(getAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zlog.Sugar().Errorf("did not connect: %v", err)
	}

	client := pb.NewEventListenerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res, err := client.NtxPayment(ctx, &pb.NtxPaymentInput{
		CallId:            inputData.CallID,
		ServiceId:         inputData.ServiceID,
		AmountOfNtx:       int32(inputData.AmountOfNtx),
		PeerId:            inputData.PeerID,
		SuccessFailStatus: inputData.SuccessFailStatus,
		Timestamp:         inputData.Timestamp,
	})

	if err != nil {
		zlog.Sugar().Errorf("connection failed: %v", err)
	}
	zlog.Sugar().Infof("Responding: %s", res.Response)
}
