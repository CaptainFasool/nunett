package libp2p

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
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
			zlog.Sugar().Errorf("failed to connect to bootstrap node %s - %v", p.ID.String(), err)
		} else {
			zlog.Sugar().Infof("Connected to Bootstrap Node %s", p.ID.String())
		}
	}

	zlog.Info("Done Bootstrapping")
	return nil
}

// Cleans up old peers from DHT
func CleanupOldPeers() {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	for _, node := range p2p.Host.Peerstore().Peers() {
		peerData, err := p2p.Host.Peerstore().Get(node, "peer_info")
		if err != nil {
			continue
		}
		if node == p2p.Host.ID() {
			continue
		}
		if Data, ok := peerData.(models.PeerData); ok {
			targetPeer, err := peer.Decode(Data.PeerID)
			if err != nil {
				zlog.Sugar().Errorf("Error decoding peer ID: %v", err)
				continue
			}
			pingResult, pingCancel := Ping(ctx, targetPeer)
			result := <-pingResult
			if result.Error == nil {
				if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
					zlog.Sugar().Infof("Peer is reachable. PeerID: %s", Data.PeerID)
					pingCancel()
					continue
				}
			} else {
				if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
					zlog.Sugar().Infof("Peer - %s is unreachable. Removing from Peerstore.", Data.PeerID)
					p2p.Host.Peerstore().Put(node, "peer_info", nil)
				}
			}
			pingCancel()
		}
	}
}

// UpdateKadDHT updates the Kad-DHT with the current node's peer info
func UpdateKadDHT() {
	// Get existing entry from Peerstore
	var PeerInfo models.PeerData
	PeerInfo.PeerID = p2p.Host.ID().String()
	peerData, err := p2p.Host.Peerstore().Get(p2p.Host.ID(), "peer_info")
	if err != nil {
		zlog.Sugar().Errorf("UpdateDHT error: unable to read self peer_info: %v", err)
	}
	if Data, ok := peerData.(models.PeerData); ok {
		PeerInfo = models.PeerData(Data)
	}

	//Get freeResources from DB
	var freeResources models.FreeResources
	if err := db.DB.Where("id = ?", 1).First(&freeResources).Error; err != nil {
		zlog.Sugar().Errorf("UpdateDHT error: unable to read free resources %v", err)
		return
	}
	// Update Free Resources
	PeerInfo.AvailableResources = freeResources

	// Update Services
	var services []models.Services
	result := db.DB.Where("job_status = ?", "running").Find(&services)
	if result.Error != nil {
		zlog.Sugar().Errorf("UpdateDHT error: Unable to read services: %v", err)
		return
	}
	PeerInfo.Services = services

	bytes, err := json.Marshal(PeerInfo)
	if err != nil {
		zlog.Sugar().Infof("UpdateDHT error: %v", err)
	}
	signature, err := signData(p2p.Host.Peerstore().PrivKey(p2p.Host.ID()), bytes)
	if err != nil {
		zlog.Sugar().Infof("Unable to sign DHT update: %v", err)
	}
	type update struct {
		Data      []byte `json:"data"`
		Signature []byte `json:"signature"`
	}
	dhtUpdate := update{
		Data:      bytes,
		Signature: signature,
	}

	dht, err := json.Marshal(dhtUpdate)
	if err != nil {
		zlog.Sugar().Infof("UpdateDHT error: %v", err)
	}

	// Store updated data in DHT
	peerID := p2p.Host.ID().String()

	// Add custom namespace to the key
	namespacedKey := customNamespace + peerID

	err = p2p.DHT.PutValue(context.Background(), namespacedKey, dht)
	if err != nil {
		zlog.Sugar().Infof("UpdateDHT error: %v", err)
	}
}

func fetchPeerStoreContents(node host.Host) []models.PeerData {
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

func fetchKadDhtContents(ctxt context.Context, resultChan chan models.PeerData) {
	zlog.Debug("Fetching DHT content for all peers")

	fetchCtx, _ := context.WithTimeout(ctxt, time.Minute)

	go func() {
		// Create a wait group to ensure all workers have finished
		var wg sync.WaitGroup

		// Create a buffered channel for the worker pool
		poolSize := 5 // Adjust the pool size as per your requirements
		workerPool := make(chan struct{}, poolSize)

		zlog.Sugar().Debugf("FetchKadDHTContents: Starting workers - number of p2p.peers: %d", len(p2p.peers))

		for _, p := range p2p.peers {
			zlog.Sugar().Debugf("FetchKadDHTContents: Waiting for worker slot for peer: %s", p.ID.String())
			workerPool <- struct{}{} // Acquire a worker slot from the pool
			zlog.Sugar().Debugf("FetchKadDHTContents: Acquired worker slot for peer: %s", p.ID.String())
			wg.Add(1) // Increment the wait group counter

			zlog.Sugar().Debugf("FetchKadDHTContents: Fetching DHT content for peer: %s ", p.ID.String())
			go func(peer peer.AddrInfo) {
				defer func() {
					<-workerPool // Release the worker slot
					wg.Done()    // Signal the wait group that the worker is done
					zlog.Sugar().Debugf("FetchKadDHTContents: Worker for %s finished", peer.ID.String())
				}()

				var updates models.KadDHTMachineUpdate

				// XXX: default 'IsAvailable' set to true here because older DMSs
				//      that don't have this parameter will by default have it as
				//      false and that will make them unable to receive jobs.
				//      NEEDS TO BE REMOVED ONCE MOST ARE UPDATED.
				peerInfo := models.PeerData{IsAvailable: true}

				// Add custom namespace to the key
				namespacedKey := customNamespace + peer.ID.String()
				bytes, err := p2p.DHT.GetValue(fetchCtx, namespacedKey)

				if err != nil {
					if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
						zlog.Sugar().Errorf(fmt.Sprintf("Couldn't retrieve dht content for peer: %s", peer.ID.String()))
					}
					return
				}
				err = json.Unmarshal(bytes, &updates)
				if err != nil {
					if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
						zlog.Sugar().Errorf("Error unmarshalling value: %v", err)
					}
					return
				}
				err = json.Unmarshal(updates.Data, &peerInfo)
				if err != nil {
					if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
						zlog.Sugar().Errorf("Error unmarshalling value: %v", err)
					}
					return
				}

				// Send the fetched value through the result channel
				resultChan <- peerInfo
			}(p)
		}

		zlog.Debug("FetchKadDHTContents: Waiting for workers to finish")
		wg.Wait()
		zlog.Debug("FetchKadDHTContents: All workers Done. Closing channel")
		close(resultChan)
	}()
}

// FetchMachines returns Machines on DHT.
func FetchMachines(node host.Host) models.Machines {
	machines := make(models.Machines)
	dhtContent := fetchPeerStoreContents(node)
	for _, peerData := range dhtContent {
		id := peerData.PeerID
		machines[id] = peerData
	}

	return machines
}

// FetchAvailableResources returns AvailableResources on DHT.
func FetchAvailableResources(node host.Host) []models.FreeResources {

	var availableResources []models.FreeResources
	dhtContent := fetchPeerStoreContents(node)
	for _, peerData := range dhtContent {
		availableResources = append(availableResources, peerData.AvailableResources)
	}

	return availableResources
}

// Filter function which returns a slice of the PeerData struct containing peers that are available.
func PeersWithAvailability(peers []models.PeerData) []models.PeerData {
	var availablePeers []models.PeerData

	for _, peer := range peers {
		if peer.IsAvailable {
			availablePeers = append(availablePeers, peer)
		}
	}
	return availablePeers
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

// Fetches peer info of peers from Kad-DHT and updates Peerstore.
func GetDHTUpdates(ctx context.Context) {
	if gettingDHTUpdate {
		zlog.Debug("GetDHTUpdates: Already Getting DHT Updates")
		return
	}
	gettingDHTUpdate = true
	zlog.Debug("GetDHTUpdates: Start Getting DHT Updates")

	machines := make(chan models.PeerData)
	fetchKadDhtContents(ctx, machines)

	for machine := range machines {
		zlog.Sugar().Debugf("GetDHTUpdates: Got machine: %v", machine.PeerID)
		targetPeer, err := peer.Decode(machine.PeerID)
		if err != nil {
			zlog.Sugar().Errorf("Error decoding peer ID: %v", err)
			gettingDHTUpdate = false
			continue
		}
		pingResult, pingCancel := Ping(ctx, targetPeer)
		res := <-pingResult
		if res.Error == nil {
			if _, verboseDebugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); verboseDebugMode {
				zlog.Sugar().Info("Peer is reachable.", "PeerID", machine.PeerID)
			}
			err := p2p.Host.Peerstore().Put(targetPeer, "peer_info", machine)
			if err != nil {
				zlog.Sugar().Errorf("Error putting peer info of %s in peerstore: %v", targetPeer.String(), err)
			}
		} else {
			if _, verboseDebugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); verboseDebugMode {
				zlog.Sugar().Info("Peer -  ", machine.PeerID, " is unreachable.")
			}
		}
		pingCancel()
	}
	gettingDHTUpdate = false
	doneGettingDHTUpdate <- true
	zlog.Debug("Done Getting DHT Updates")
}

func signData(hostPrivateKey crypto.PrivKey, data []byte) ([]byte, error) {
	signature, err := hostPrivateKey.Sign(data)
	if err != nil {
		return nil, err
	}
	return signature, nil
}
