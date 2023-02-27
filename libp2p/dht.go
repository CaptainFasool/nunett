package libp2p

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
)

func (p2p DMSp2p) autoRelay(ctx context.Context) {
	for {
		peers, err := p2p.DHT.GetClosestPeers(ctx, p2p.Host.ID().String())
		if err != nil {
			zlog.Sugar().Infof("GetClosestPeers error: %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		for _, p := range peers {
			addrs := p2p.Host.Peerstore().Addrs(p)
			if len(addrs) == 0 {
				continue
			}
			make(chan peer.AddrInfo) <- peer.AddrInfo{
				ID:    p,
				Addrs: addrs,
			}
		}
	}
}

func (p2p DMSp2p) BootstrapNode(ctx context.Context) error {
	Bootstrap(ctx, p2p.Host, p2p.DHT)
	go p2p.autoRelay(ctx)

	return nil
}

func Bootstrap(ctx context.Context, node host.Host, idht *dht.IpfsDHT) error {
	if err := idht.Bootstrap(ctx); err != nil {
		return err
	}

	for _, p := range dht.GetDefaultBootstrapPeerAddrInfos() {
		if err := node.Connect(ctx, p); err != nil {
			zlog.Sugar().Errorf("failed to connect to bootstrap node %v\n", p.ID)
		} else {
			zlog.Sugar().Infof("Connected to Bootstrap Node %v\n", p.ID)
		}
	}

	zlog.Info("Done Bootstrapping")
	return nil
}

func DhtUpdateHandler(s network.Stream) {
	peerInfo := models.PeerData{}
	var peerID peer.ID
	data, err := io.ReadAll(s)
	if err != nil {
		zlog.Sugar().Infof("DHTUpdateHandler error: %s", err.Error())
	}
	err = json.Unmarshal(data, &peerInfo)
	if err != nil {
		zlog.Sugar().Infof("DHTUpdateHandler error: %s", err.Error())
	}
	// Update Peerstore
	peerID, err = peer.Decode(peerInfo.PeerID)
	if err != nil {
		zlog.Sugar().Infof("DHTUpdateHandler error: %s", err.Error())
	}
	p2p.Host.Peerstore().Put(peerID, "peer_info", peerInfo)
}

func SendDHTUpdate(peerInfo models.PeerData, s network.Stream) {
	w := bufio.NewWriter(s)
	data, err := json.Marshal(peerInfo)
	if err != nil {
		zlog.Sugar().Infof("SendDHTUpdate error: %s", err.Error())
	}
	n, err := w.Write(data)
	if n != len(data) {
		zlog.Sugar().Infof("SendDHTUpdate error: %s", err.Error())
	}
	if err != nil {
		zlog.Sugar().Infof("SendDHTUpdate error: %s", err.Error())
	}
	err = w.Flush()
	if err != nil {
		zlog.Sugar().Infof("SendDHTUpdate error: %s", err.Error())
	}
}

func UpdateDHT() {
	// Get existing entry from Peerstore
	var PeerInfo models.PeerData
	PeerInfo.PeerID = p2p.Host.ID().String()
	peerData, err := p2p.Host.Peerstore().Get(p2p.Host.ID(), "peer_info")
	if err != nil {
		zlog.Sugar().Infof("UpdateAvailableResources error: %s", err.Error())
	}
	if Data, ok := peerData.(models.PeerData); ok {
		PeerInfo = models.PeerData(Data)
	}

	//Get freeResources from DB
	var freeResources models.FreeResources
	if err := db.DB.Where("id = ?", 1).First(&freeResources).Error; err != nil {
		panic(err)

	}
	// Update Free Resources
	PeerInfo.AvailableResources = freeResources

	// Update Services
	var services []models.Services
	result := db.DB.Find(&services)
	if result.Error != nil {
		panic(result.Error)
	}
	PeerInfo.Services = services

	// Update peerstore with updated data
	p2p.Host.Peerstore().Put(p2p.Host.ID(), "peer_info", PeerInfo)

	ctx := context.Background()
	defer ctx.Done()

	// Broadcast updated peer_info to connected nodes
	for _, peer := range p2p.Host.Network().Peers() {
		stream, err := p2p.Host.NewStream(ctx, peer, DHTProtocolID)
		// Ignoring error because some peers might not support specified protocol
		if err != nil {
			continue
		}
		SendDHTUpdate(PeerInfo, stream)
		stream.Close()

	}

}

func fetchDhtContents(node host.Host) []models.PeerData {
	var dhtContent []models.PeerData
	var PeerInfo models.PeerData
	for _, peer := range node.Peerstore().Peers() {
		if node.ID() == peer {
			// skip if peerID matches our ID
			continue
		}
		peerData, err := node.Peerstore().Get(peer, "peer_info")
		if err != nil {
			// skip error since we might not have an entry for all peers
			continue
		}
		if Data, ok := peerData.(models.PeerData); ok {
			PeerInfo = models.PeerData(Data)
		}
		dhtContent = append(dhtContent, PeerInfo)
	}
	return dhtContent
}

// FetchMachines returns Machines on DHT.
func FetchMachines(node host.Host) models.Machines {
	machines := make(models.Machines)
	dhtContent := fetchDhtContents(node)
	for _, peerData := range dhtContent {
		id := peerData.PeerID
		machines[id] = peerData
	}

	return machines
}

// FetchAvailableResources returns AvailableResources on DHT.
func FetchAvailableResources(node host.Host) []models.FreeResources {

	var availableResources []models.FreeResources
	dhtContent := fetchDhtContents(node)
	for _, peerData := range dhtContent {
		availableResources = append(availableResources, peerData.AvailableResources)
	}

	return availableResources
}

// PeersWithCardanoAllowed is a filter function which returns a slice of
// PeerData based on allow_cardano metadata on peer.
func PeersWithCardanoAllowed(peers []models.PeerData) []models.PeerData {
	var cardanoAllowedPeers []models.PeerData

	for _, peer := range peers {
		if peer.AllowCardano {
			cardanoAllowedPeers = append(cardanoAllowedPeers, peer)
		}
	}
	return cardanoAllowedPeers
}

// PeersWithGPU is a filter function which returns a slice of
// PeerData based on has_gpu metadata on peer.
func PeersWithGPU(peers []models.PeerData) []models.PeerData {
	var peersWithGPU []models.PeerData

	for _, peer := range peers {
		if peer.HasGpu {
			peersWithGPU = append(peersWithGPU, peer)
		}
	}
	return peersWithGPU
}

// PeersWithMatchingSpec takes in a depReq which has minimum spec specified to
// run a job. Then it matches it against the peers available.
func PeersWithMatchingSpec(peers []models.PeerData, depReq models.DeploymentRequest) []models.PeerData {
	constraints := depReq.Constraints

	var peerWithMachingSpec []models.PeerData

	for _, peer := range peers {
		prAvRes := peer.AvailableResources
		if prAvRes.TotCpuHz > constraints.CPU && prAvRes.Ram > constraints.RAM {
			peerWithMachingSpec = append(peerWithMachingSpec, peer)
		}
	}
	return peerWithMachingSpec
}

// FilterPeers searches for available compute providers given specific parameters in depReq.
func FilterPeers(depReq models.DeploymentRequest, node host.Host) []models.PeerData {
	machines := FetchMachines(node)

	var peers []models.PeerData

	for _, val := range machines {
		peers = append(peers, val)
	}

	peers = PeersWithMatchingSpec(peers, depReq)
	if depReq.ServiceType == "ml-training-gpu" {
		peers = PeersWithGPU(peers)
	}

	if depReq.ServiceType == "cardano_node" {
		peers = PeersWithCardanoAllowed(peers)
	}

	return peers
}
