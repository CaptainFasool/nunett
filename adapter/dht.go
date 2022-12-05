package adapter

import (
	"context"
	"encoding/json"
	fmt "fmt"
	"time"

	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func fetchDhtContents() (*DhtContents, error) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(utils.AdapterGrpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

// func FetchDht() ([]byte, error) {
// 	content, err := fetchDht()
// 	if err != nil {
// 		return nil, err
// 	}
// 	b := []byte(content)

// 	return b, nil
// }

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

	var peerWithMachingSpec []Peer

	for _, peer := range peers {
		prAvRes := peer.AvailableResources
		if prAvRes.CpuHz > constraints.CPU && prAvRes.Ram > constraints.RAM {
			peerWithMachingSpec = append(peerWithMachingSpec, peer)
		}
	}

	return peerWithMachingSpec
}

// SendMessage takes in a nodeID of a node from the P2P network and posts a message
// to it. `message` is supposed to be a JSON marshalled in string.
func SendMessage(nodeID string, message string) (string, error) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(utils.AdapterGrpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func UpdateAvailableResoruces() (string, error) {
	conn, err := grpc.Dial(utils.AdapterGrpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	client := NewNunetAdapterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var freeResources models.FreeResources
	if err := db.DB.Where("id = ?", 1).First(&freeResources).Error; err != nil {
		panic(err)

	}

	marshaledRes, err := json.Marshal(freeResources)
	if err != nil {
		panic(err)
	}

	r, err := client.UpdateDHT(ctx, &DHTUpdateContent{
		HasGpu:             "",
		AllowCardano:       "",
		GpuInfo:            "",
		ServicesIndex:      "",
		AvailableResources: string(marshaledRes),
	})

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return r.GetResponse(), nil
}

func getSelfNodeID() (string, error) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(utils.AdapterGrpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	client := NewNunetAdapterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := client.GetSelfPeer(ctx, &GetPeerParams{})

	if err != nil {
		return "", err
	}

	return r.GetNodeId(), nil
}

func UpdateMachinesTable() {
	machines, err := FetchMachines()
	if err != nil {
		panic(err)
	}

	nodeId, _ := getSelfNodeID()

	var machine_index models.Machine
	if res := db.DB.Find(&machine_index); res.RowsAffected == 0 {
		result := db.DB.Create(&machine_index)
		if result.Error != nil {
			panic(result.Error)
		}
	}

	var available_resources models.AvailableResources
	if err := db.DB.Where("id = ?", 1).First(&available_resources).Error; err != nil {
		panic(err)
	}

	var freeResources models.FreeResources
	if err := db.DB.Where("id = ?", 1).First(&freeResources).Error; err != nil {
		panic(err)
	}

	var peerinfo models.PeerInfo
	if res := db.DB.Find(&peerinfo); res.RowsAffected == 0 {
		result := db.DB.Create(&peerinfo)
		if result.Error != nil {
			panic(result.Error)
		}
	}
	// var ip models.IP = models.IP(machines[nodeId].PeerInfo.Address)

	peerinfo.NodeID = nodeId
	peerinfo.Mid = machines[nodeId].PeerInfo.Mid
	peerinfo.Key = machines[nodeId].PeerInfo.Key
	peerinfo.PublicKey = machines[nodeId].PeerInfo.PublicKey
	// peerinfo.Address = machines[nodeId].PeerInfo.Address
	result := db.DB.Model(&models.PeerInfo{}).Where("id = ?", peerinfo.ID).Updates(peerinfo)
	if result.Error != nil {
		panic(result.Error)
	}

	machine_index.NodeId = nodeId
	machine_index.PeerInfo = int(peerinfo.ID)
	machine_index.AvailableResources = int(available_resources.ID)
	machine_index.FreeResources = int(freeResources.ID)
	// machine_index.Ip_Addr = machines[nodeId].PeerInfo.Address
	// TODO: Add tokenomics address and tokenomics blockchain fields to DB.

	result = db.DB.Model(&models.Machine{}).Where("id = ?", 1).Updates(machine_index)
	if result.Error != nil {
		panic(result.Error)
	}

}

// GetPeerID returns self NodeID from the adapter
func GetPeerID() (string, error) {
	return getSelfNodeID()
}
