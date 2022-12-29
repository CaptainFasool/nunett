package statsdb

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"time"

	"gitlab.com/nunet/device-management-service/models"
	pb "gitlab.com/nunet/device-management-service/statsdb/event_listener_spec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "stats-event-listener.dev.nunet.io:80", "the address to connect to")
)

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
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	client := pb.NewEventListenerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
		log.Fatalf("connection failed: %v", err)
	}
	log.Printf("Responding: %s", res.PeerId)
}

// ServiceCall sends the info of the service call made to dms to stats via grpc call.
func ServiceCall(inputData models.ServiceCall) {
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	client := pb.NewEventListenerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
		log.Fatalf("connection failed: %v", err)
	}
	log.Printf("Responding: %s", res.Response)

}

// ServiceRun sends the status info of the service process on host machine to stats via grpc call.
func ServiceRun(inputData models.ServiceRun) {
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	client := pb.NewEventListenerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.ServiceRun(ctx, &pb.ServiceRunInput{
		CallId:              inputData.CallID,
		PeerIdOfServiceHost: inputData.PeerIDOfServiceHost,
		Status:              inputData.Status,
		Timestamp:           inputData.Timestamp,
	})

	if err != nil {
		log.Fatalf("connection failed: %v", err)
	}
	log.Printf("Responding: %s", res.Response)
}

// HeartBeat pings the statsdb in every 10s for detacting live status of device via grpc call.
func HeartBeat(peerID string) {
	for {
		conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}

		client := pb.NewEventListenerClient(conn)

		res, err := client.HeartBeat(context.Background(), &pb.HeartBeatInput{
			PeerId: peerID,
		})

		if err != nil {
			log.Fatalf("connection failed: %v", err)
		}
		log.Printf("Responding: %s", res.PeerId)

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

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	client := pb.NewEventListenerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.DeviceResourceChange(ctx, &pb.DeviceResourceChangeInput{
		PeerId:                   inputData.NodeID,
		ChangedAttributeAndValue: &DeviceResourceParams,
		Timestamp:                float32(GetTimestamp()),
	})

	if err != nil {
		log.Fatalf("connection failed: %v", err)
	}

	log.Printf("Responding: %s", res.PeerId)
}
