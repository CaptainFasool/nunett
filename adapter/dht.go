package adapter

import (
	"context"
	"encoding/json"
	"time"

	"gitlab.com/nunet/device-management-service/models"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func fetchDhtContents() (*DhtContents, error) {
	// Set up a connection to the server.
	address := "localhost:60777"
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := NewNunetAdapterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := client.GetDhtContent(ctx, &GetDhtParams{})

	if err != nil {
		return nil, err
	}

	return r, nil
}

// FetchMachines returns Machines on DHT.
func FetchMachines() (Machines, error) {
	dhtContent, err := fetchDhtContents()
	if err != nil {
		return nil, err
	}
	machinesByte := []byte(dhtContent.GetMachinesIndex())
	// machinesByte, err := os.ReadFile("/tmp/machine_index.json")

	var machines Machines

	err = json.Unmarshal(machinesByte, &machines)
	if err != nil {
		return nil, err
	}

	return machines, nil
}

// FetchAvailableResources returns AvailableResources on DHT.
// TODO: Return actual struct, not bytes; check FetchMachines
func FetchAvailableResources() ([]byte, error) {
	dhtContent, err := fetchDhtContents()
	if err != nil {
		return nil, err
	}
	b := []byte(dhtContent.GetAvailableResourcesIndex())

	return b, nil
}

// FetchServices returns Services on DHT.
// TODO: Return actual struct, not bytes; check FetchMachines
func FetchServices() ([]byte, error) {
	dhtContent, err := fetchDhtContents()
	if err != nil {
		return nil, err
	}
	b := []byte(dhtContent.GetServicesIndex())

	return b, nil
}

// PeersWithCardanoAllowed is a filter function which returns a slice of
// Peer based on allow_cardano metadata on peer.
func PeersWithCardanoAllowed(peers []Peer) []Peer {
	var cardanoAllowedPeers []Peer

	for idx, peer := range peers {
		if peer.AllowCardano == "True" {
			cardanoAllowedPeers = append(cardanoAllowedPeers, peer)
		}
		_ = idx
	}

	return cardanoAllowedPeers
}

// PeersWithGPU is a filter function which returns a slice of
// Peer based on has_gpu metadata on peer.
func PeersWithGPU(peers []Peer) []Peer {
	var peersWithGPU []Peer

	for idx, peer := range peers {
		if peer.HasGpu == "True" {
			peersWithGPU = append(peersWithGPU, peer)
		}
		_ = idx
	}

	return peersWithGPU
}

// PeersWithMatchingSpec takes in a depReq which has minimum spec specified to
// run a job. Then it matches it against the peers available.
func PeersWithMatchingSpec(peers []Peer, depReq models.DeploymentRequest) []Peer {
	constraints := depReq.Constraints

	var peersWithMachingSpec []Peer

	for _, peer := range peers {
		prAvRes := peer.AvailableResources
		if prAvRes.CpuHz > constraints.CPU && prAvRes.Ram > constraints.RAM {
			peersWithMachingSpec = append(peersWithMachingSpec, peer)
		}
	}

	return peersWithMachingSpec
}

// SendMessage takes in a nodeID of a node from the P2P network and posts a message
// to it. `message` is supposed to be a JSON marshalled in string.
func SendMessage(nodeID string, message string) (string, error) {
	// Set up a connection to the server.
	address := "localhost:9998"
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// log.Fatalf("did not connect: %v", err)
		return "", err
	}
	defer conn.Close()

	client := NewNunetAdapterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := client.SendMessage(ctx, &MessageParams{
		NodeId:         nodeID,
		MessageContent: message,
	})

	if err != nil {
		return "", err
	}

	return r.GetMessageResponse(), nil
}
