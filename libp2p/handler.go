package libp2p

import (
	"bufio"
	"context"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
)

const dhtUpdateProtocol = protocol.ID("/nunet/dms/dht/0.0.1")

var Node host.Host
var DHT *dht.IpfsDHT
var Ctx context.Context

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

func CheckOnboarding() {
	// Checks for saved metadata and create a new host
	var libp2pInfo models.Libp2pInfo
	result := db.DB.Where("id = ?", 1).Find(&libp2pInfo)
	if result.Error != nil {
		panic(result.Error)
	}
	if libp2pInfo.PrivateKey != nil {
		// Recreate private key
		priv, err := crypto.UnmarshalPrivateKey(libp2pInfo.PrivateKey)
		if err != nil {
			panic(err)
		}
		RunNode(priv)
		go UpdateDHT()
	}
}

func RunNode(priv crypto.PrivKey) {
	ctx := context.Background()

	host, dht, err := NewHost(ctx, 9000, priv)
	if err != nil {
		panic(err)
	}
	host.SetStreamHandler(dhtUpdateProtocol, dhtUpdateHandler)
	err = Bootstrap(ctx, host, dht)
	if err != nil {
		panic(err)
	}
	go Discover(ctx, host, dht, "nunet")
	if _, err := host.Peerstore().Get(host.ID(), "peer_info"); err != nil {
		peerInfo := models.PeerData{}
		peerInfo.PeerID = host.ID().String()
		host.Peerstore().Put(host.ID(), "peer_info", peerInfo)
	}

	Node = host
	DHT = dht
	Ctx = ctx
}

func sendDHTUpdate(peerInfo models.PeerData, s network.Stream) {
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

func dhtUpdateHandler(s network.Stream) {
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
	Node.Peerstore().Put(peerID, "peer_info", peerInfo)
}

func UpdateDHT() {
	// Get existing entry from Peerstore
	var PeerInfo models.PeerData
	PeerInfo.PeerID = Node.ID().String()
	peerData, err := Node.Peerstore().Get(Node.ID(), "peer_info")
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
	Node.Peerstore().Put(Node.ID(), "peer_info", PeerInfo)

	// Broadcast updated peer_info to connected nodes
	for _, peer := range Node.Network().Peers() {
		stream, err := Node.NewStream(Ctx, peer, dhtUpdateProtocol)
		// Ignoring error because some peers might not support specified protocol
		if err != nil {
			continue
		}
		sendDHTUpdate(PeerInfo, stream)
		stream.Close()

	}
}

// ListPeers  godoc
// @Summary      Return list of peers currently connected to
// @Description  Gets a list of peers the libp2p node can see within the network and return a list of peers
// @Tags         run
// @Produce      json
// @Success      200  {string}	string
// @Router       /peers [get]
func ListPeers(c *gin.Context) {

	peers, err := getPeers(Ctx, Node, DHT, "nunet")
	if err != nil {
		c.JSON(500, gin.H{"error": "can not fetch peers"})
		panic(err)
	}
	c.JSON(200, peers)

}
