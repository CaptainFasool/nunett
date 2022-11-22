package statsdb

import (
	"context"
	"flag"
	"log"
	"time"

	"gitlab.com/nunet/device-management-service/adapter"
	"gitlab.com/nunet/device-management-service/models"
	pb "gitlab.com/nunet/device-management-service/statsdb/event_listener_spec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "65.108.205.60:50051", "the address to connect to")
)

// GetPeerID returns self Node ID from the adapter
func GetPeerID() string {
	NodeID, err := adapter.GetPeerID()

	if err != nil {
		log.Print("unable to get Node ID")
	}

	return NodeID
}

// GetTimestamp returns current unix timestamp
func GetTimestamp() int64 {
	return time.Now().Unix()
}

// NOTE:- **Not removing the hardcoded data kept at the moment for testing purpose **
// var New_device_onboard_params = models.NewDeviceOnboarded{
// 	PeerID:        "test-peer-id-1234567890",
// 	CPU:           4342.0,
// 	RAM:           1223.0,
// 	Network:       76.0,
// 	DedicatedTime: 12.0,
// 	Timestamp:     6876344.0,
// }
// var Service_call_params = models.ServiceCall{
// 	CallID:              6.808776152033109e+16,
// 	PeerIDOfServiceHost: "test-peer-id-1234567890",
// 	ServiceID:           "test-service-id",
// 	CPUUsed:             2345.0,
// 	MaxRAM:              3231.0,
// 	MemoryUsed:          2313.0,
// 	NetworkBwUsed:       23.0,
// 	TimeTaken:           12.0,
// 	Status:              "accepted",
// 	Timestamp:           5.5,
// }

// var Service_run_params = models.ServiceRun{
// 	CallID:              6.808776152033109e+16,
// 	PeerIDOfServiceHost: "test-peer-id-123456778",
// 	Status:              "success",
// 	Timestamp:           5.5,
// }

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
