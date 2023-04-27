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

func (p2p DMSp2p) BootstrapNode(ctx context.Context) error {
	Bootstrap(ctx, p2p.Host, p2p.DHT)

	return nil
}

func Bootstrap(ctx context.Context, node host.Host, idht *dht.IpfsDHT) error {
	if err := idht.Bootstrap(ctx); err != nil {
		return err
	}

	for _, nb := range NuNetBootstrapPeers {
		p, _ := peer.AddrInfoFromP2pAddr(nb)
		if err := node.Connect(ctx, *p); err != nil {
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
	peerInfo.Timestamp = time.Now().Unix()
	var peerID peer.ID
	data, err := io.ReadAll(s)
	if err != nil {
		zlog.Sugar().Errorf("DHTUpdateHandler error: %v", err)
	}
	err = json.Unmarshal(data, &peerInfo)
	if err != nil {
		zlog.Sugar().Errorf("DHTUpdateHandler error: %v", err)
	}
	// Update Peerstore
	peerID, err = peer.Decode(peerInfo.PeerID)
	if err != nil {
		zlog.Sugar().Errorf("DHTUpdateHandler error: %v", err)
	}

	stringPeerInfo, err := json.Marshal(peerInfo)
	if err != nil {
		zlog.Sugar().Errorf("failed to json marshal peerInfo: %v", err)
	}
	zlog.Sugar().Debugf("dht update from: %s -> %v", peerID.String(), string(stringPeerInfo))

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

// Cleans up old peers from DHT
func CleanupOldPeers() {
	for _, peer := range p2p.Host.Peerstore().Peers() {
		peerData, err := p2p.Host.Peerstore().Get(peer, "peer_info")
		if err != nil {
			continue
		}
		if peer == p2p.Host.ID() {
			continue
		}
		if Data, ok := peerData.(models.PeerData); ok {
			if time.Now().Unix()-Data.Timestamp > 180 {
				p2p.Host.Peerstore().Put(peer, "peer_info", nil)
			}
		}
	}
}

func UpdateDHT() {
	// Get existing entry from Peerstore
	var PeerInfo models.PeerData
	PeerInfo.PeerID = p2p.Host.ID().String()
	peerData, err := p2p.Host.Peerstore().Get(p2p.Host.ID(), "peer_info")
	if err != nil {
		zlog.Sugar().Infof("UpdateDHT error: %s", err.Error())
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
	result := db.DB.Where("job_status = ?", "running").Find(&services)
	if result.Error != nil {
		panic(result.Error)
	}
	PeerInfo.Services = services

	// Update peerstore with updated data
	p2p.Host.Peerstore().Put(p2p.Host.ID(), "peer_info", PeerInfo)

	ctx := context.Background()
	defer ctx.Done()

	// Broadcast updated peer_info to connected nodes
	addr, err := p2p.getPeers(ctx, "nunet")

	p2p.peers = addr
	zlog.Sugar().Debugf("Obtained peers for DHT update: %v", addr)
	if err != nil {
		zlog.Sugar().Debugf("UpdateDHT error: %s", err.Error())
	}

	for _, addr := range addr {
		zlog.Sugar().Debugf("Attempting to Send DHT Update to: %s", addr.ID.String())
		peerID := addr.ID

		relayPeer <- addr

		// XXX wait 5 seconds for stream creation
		//     needs better implementation in the future
		streamCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		stream, err := p2p.Host.NewStream(streamCtx, peerID, DHTProtocolID)
		// Ignoring error because some peers might not support specified protocol
		if err != nil {
			zlog.Sugar().Debugf("UpdateDHT Create Stream error: %v --- peer: %v", err, peerID.String())
			continue
		}
		zlog.Sugar().Debugf("Sending DHT update to %s", peerID.String())
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
