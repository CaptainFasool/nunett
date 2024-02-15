package libp2p

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/internal"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/klogger"
	"golang.org/x/net/context"

	"gitlab.com/nunet/device-management-service/models"
	"go.uber.org/zap"
)

var clients = make(map[internal.WebSocketConnection]string)

// TODO: Move this struct somewhere else
type SelfPeer struct {
	ID    string                `json:"id"`
	Addrs []multiaddr.Multiaddr `json:"addrs"`
}

func DeviceStatus() (bool, error) {
	if p2p.Host == nil {
		return false, fmt.Errorf("host node has not yet been initialized")
	}
	peerData, err := p2p.Host.Peerstore().Get(p2p.Host.ID(), "peer_info")
	if err != nil {
		return false, fmt.Errorf("could not retrieve data from peer: %w", err)
	}
	p, ok := peerData.(models.PeerData)
	if !ok {
		return false, fmt.Errorf("failed to type assert peer data for peer ID: %s", p2p.Host.ID().String())
	}
	return p.IsAvailable, nil
}

func ChangeDeviceStatus(ctx context.Context, status bool) error {
	if p2p.Host == nil {
		return fmt.Errorf("host node has not yet been initialized")
	}

	peerData, err := p2p.Host.Peerstore().Get(p2p.Host.ID(), "peer_info")
	if err != nil {
		return fmt.Errorf("could not retrieve data from peer: %w", err)
	}

	p, ok := peerData.(models.PeerData)
	if !ok {
		return fmt.Errorf("failed to type assert peer data for peer ID: %s", p2p.Host.ID().String())
	}

	if p.IsAvailable == status {
		return fmt.Errorf("no change in device status")
	}

	p.IsAvailable = status

	err = p2p.Host.Peerstore().Put(p2p.Host.ID(), "peer_info", peerData)
	if err != nil {
		return fmt.Errorf("failed to put peer data into peerstore: %w", err)
	}

	// update database value
	var p2pInfo models.Libp2pInfo
	res := db.DB.Find(&p2pInfo)
	if res.Error != nil {
		return fmt.Errorf("failed to retrieve libp2p info from database: %w", res.Error)
	}

	p2pInfo.Available = status
	res = db.DB.Save(&p2pInfo)
	if res.Error != nil {
		return fmt.Errorf("failed to update libp2p info in database: %w", res.Error)
	}
	UpdateKadDHT()
	return nil
}

func ListPeers() ([]peer.AddrInfo, error) {
	if p2p.Host == nil {
		return nil, fmt.Errorf("host node has not yet been initialized")
	}
	klogger.Logger.Info("List peers executed by " + p2p.Host.ID().String())
	// nil slice (not initialized)
	if p2p.peers == nil {
		klogger.Logger.Error("Peers haven't yet been fetched.")
		return nil, fmt.Errorf("Peers haven't yet been fetched.")
	}
	// empty slice (initialized)
	if len(p2p.peers) == 0 {
		klogger.Logger.Error("No Peers Yet.")
		return nil, fmt.Errorf("no peers yet")
	}
	klogger.Logger.Info("List of peers" + p2p.Host.ID().String())
	return p2p.peers, nil
}

func ListDHTPeers(ctx context.Context) ([]peer.ID, error) {
	if p2p.Host == nil {
		return nil, fmt.Errorf("host node has not yet been initialized")
	}

	klogger.Logger.Info("List DHT peers executed by " + p2p.Host.ID().String())

	_, debug := os.LookupEnv("NUNET_DEBUG_VERBOSE")
	var dhtPeers []peer.ID
	for _, peer := range p2p.peers {
		_, err := p2p.Host.Peerstore().Get(peer.ID, "peer_info")
		if err != nil && debug {
			zlog.ErrorContext(ctx, fmt.Sprintf("could not retrieve DHT content for peer %s", peer.String()), zap.Error(err))
			continue
		}
		if peer.ID == p2p.Host.ID() {
			continue
		}
		dhtPeers = append(dhtPeers, peer.ID)
	}

	if len(dhtPeers) == 0 {
		klogger.Logger.Error("No peers found")
	}

	logResp, err := json.Marshal(dhtPeers)
	if err != nil {
		// log with context
		zlog.ErrorContext(ctx, "failed to json marshal DHT peers: %w", zap.Error(err))
		klogger.Logger.Error("failed to json marshal DHT peers")
	}
	klogger.Logger.Info("Response: " + string(logResp))

	return dhtPeers, nil
}

func ListKadDHTPeers(ctx context.Context) ([]string, error) {
	_, debug := os.LookupEnv("NUNET_DEBUG_VERBOSE")
	if p2p.Host == nil {
		return nil, fmt.Errorf("host node has not yet been initialized")
	}
	klogger.Logger.Info("List Kademlia DHT peers executed by " + p2p.Host.ID().String())

	var dhtPeers []string
	for _, peer := range p2p.peers {
		var updates models.KadDHTMachineUpdate
		var peerInfo models.PeerData

		// Add custom namespace to the key
		key := customNamespace + peer.ID.String()
		bytes, err := p2p.DHT.GetValue(ctx, key)
		if err != nil {
			if debug {
				zlog.Sugar().Errorf(fmt.Sprintf("could not retrieve dht content for peer: %s", peer.String()))
			}
			continue
		}
		err = json.Unmarshal(bytes, &updates)
		if err != nil {
			if debug {
				zlog.Sugar().Errorf("could not unmarshal updates: %w", err)
			}
			continue
		}
		err = json.Unmarshal(updates.Data, &peerInfo)
		if err != nil {
			if debug {
				zlog.Sugar().Errorf("could not unmarshal peer info: %w", err)
			}
			continue
		}
		dhtPeers = append(dhtPeers, peerInfo.PeerID)
	}
	return dhtPeers, nil
}

func SelfPeerInfo() (*SelfPeer, error) {
	start := time.Now().UnixMilli()
	if p2p.Host == nil {
		return nil, fmt.Errorf("host node has not yet been initialized")
	}

	klogger.Logger.Info(" result : Self Peer ID " + p2p.Host.ID().String())
	zlog.Sugar().Infof("----------klogger time taken=%d ms", time.Now().UnixMilli()-start)

	var self SelfPeer
	self.ID = p2p.Host.ID().String()
	self.Addrs = p2p.Host.Addrs()

	zlog.Sugar().Infof("----------overall time taken=%d ms", time.Now().UnixMilli()-start)
	return &self, nil
}

func StartChat(w http.ResponseWriter, r *http.Request, s network.Stream, id string) {
	ws, err := internal.UpgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		zlog.Sugar().Errorf("failed to set websocket upgrade: %w\n", err)
		return
	}

	welcomeMsg := fmt.Sprintf("Enter the message that you wish to send to %s with stream ID: %s and press return.", id, s.ID())

	err = ws.WriteMessage(websocket.TextMessage, []byte(welcomeMsg))
	if err != nil {
		zlog.Sugar().Errorf(err.Error())
	}

	conn := internal.WebSocketConnection{Conn: ws}
	clients[conn] = id

	reader := bufio.NewReader(s)
	writer := bufio.NewWriter(s)

	go SockReadStreamWrite(&conn, s, writer)
	go StreamReadSockWrite(&conn, s, reader)
}

func JoinChat(w http.ResponseWriter, r *http.Request, id int) error {
	if id >= len(inboundChatStreams) {
		return fmt.Errorf("unknown stream ID: %d", id)
	}

	ws, err := internal.UpgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		zlog.Sugar().Errorf("failed to set websocket upgrade: %w\n", err)
	}

	welcomeMsg := "Enter the message that you wish to send and press return."

	err = ws.WriteMessage(websocket.TextMessage, []byte(welcomeMsg))
	if err != nil {
		zlog.Sugar().Errorf(err.Error())
	}

	conn := internal.WebSocketConnection{Conn: ws}
	clients[conn] = strconv.Itoa(id)

	stream := inboundChatStreams[id]

	// remove the stream element from slice and updates it
	copy(inboundChatStreams[id:], inboundChatStreams[id+1:])
	inboundChatStreams[len(inboundChatStreams)-1] = nil
	inboundChatStreams = inboundChatStreams[:len(inboundChatStreams)-1]

	reader := bufio.NewReader(stream)
	writer := bufio.NewWriter(stream)

	go SockReadStreamWrite(&conn, stream, writer)
	go StreamReadSockWrite(&conn, stream, reader)
	return nil
}

func CreateChatStream(ctx context.Context, id string) (network.Stream, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("empty peer ID string")
	}
	if id == p2p.Host.ID().String() {
		return nil, fmt.Errorf("peer ID cannot be self peer ID")
	}

	p, err := peer.Decode(id)
	if err != nil {
		zlog.Sugar().Errorf("could not decode string ID to peer ID: %w", err)
		return nil, fmt.Errorf("could not decode string ID to peer ID: %w", err)
	}

	stream, err := p2p.Host.NewStream(ctx, p, protocol.ID(ChatProtocolID))
	if err != nil {
		zlog.Sugar().ErrorfContext(ctx, "could not create stream with peer: %w", err)
		return nil, fmt.Errorf("could not create stream with peer: %w", err)
	}
	return stream, nil
}

func DumpDHT(ctx context.Context) ([]models.PeerData, error) {
	_, debug := os.LookupEnv("NUNET_DEBUG_VERBOSE")
	if p2p.Host == nil {
		return nil, fmt.Errorf("host node has not yet been initialized")
	}

	var dht []models.PeerData
	for _, peer := range p2p.Host.Peerstore().Peers() {
		info, err := p2p.Host.Peerstore().Get(peer, "peer_info")
		if err != nil {
			if debug {
				zlog.ErrorContext(ctx, fmt.Sprintf("could not retrieve DHT content for peer: %s", peer.String()), zap.Error(err))
			}
			continue
		}
		data, ok := info.(models.PeerData)
		if ok {
			dht = append(dht, models.PeerData(data))
		}
	}
	return dht, nil
}

// SUGGESTION: Define two functions SetDepReqPeer and GetDepReqPeer
// Current function have both SET and GET logic which make things confusing
func DefaultDepReqPeer(ctx context.Context, id string) (string, error) {
	target := config.GetConfig().Job.TargetPeer

	if id == "0" { // remove any previously set peer
		config.SetConfig("job.target_peer", "")
		return "", nil
	} else if id == "" && target == "" { // return nil target peer
		return "", nil
	} else if id == "" { // return current target peer
		return target, nil
	} else if id == p2p.Host.ID().String() {
		return "", fmt.Errorf("peer ID can not be self peer ID")
	}

	p, err := peer.Decode(id)
	if err != nil {
		zlog.Sugar().Errorf("could not decode string ID to peerID: %w", err)
		return "", fmt.Errorf("could not decode string ID to peer ID: %w", err)
	}
	_, err = p2p.Host.Peerstore().Get(p, "peer_info")
	if err != nil {
		return "", fmt.Errorf("could not get target peer ID from DHT: %w", err)
	}
	pingCh, cancel := Ping(ctx, p)
	defer cancel()

	msg := <-pingCh
	if msg.Error != nil {
		zlog.Sugar().Errorf("could not ping peer: %w", msg.Error)
		return "", fmt.Errorf("peer not online")
	}
	config.SetConfig("job.target_peer", id)
	return id, nil
}

// DEBUG ONLY
func ManualDHTUpdate(ctx context.Context) {
	go UpdateKadDHT()
	GetDHTUpdates(ctx)
}

// DEBUG ONLY
func CleanupPeer(id string) error {
	if id == "" {
		return fmt.Errorf("peer ID not provided")
	}
	if id == p2p.Host.ID().String() {
		return fmt.Errorf("peer ID cannot be self peer ID")
	}
	target, err := peer.Decode(id)
	if err != nil {
		zlog.Sugar().Errorf("Could not decode string ID to peerID: %v", err)
		return fmt.Errorf("could not decode string ID to peer ID: %w", err)
	}
	p2p.Host.Peerstore().RemovePeer(target)
	return nil
}

// DEBUG ONLY
func PingPeer(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, fmt.Errorf("peer ID not provided")
	}
	if id == p2p.Host.ID().String() {
		return false, fmt.Errorf("peer ID cannot be self peer ID")
	}

	target, err := peer.Decode(id)
	if err != nil {
		zlog.Sugar().Errorf("Could not decode string ID to peerID: %v", err)
		return false, fmt.Errorf("could not decode string ID to peer ID: %w", err)
	}

	var aval bool
	_, err = p2p.Host.Peerstore().Get(target, "peer_info")
	if err != nil {
		aval = false
	} else {
		aval = true
	}

	pingCh, cancel := Ping(ctx, target)
	defer cancel()
	result := <-pingCh
	zlog.Sugar().Infof("Pinged %s --> RTT: %s", target.String(), result.RTT)
	if result.Error != nil {
		return aval, fmt.Errorf("ping failed with peer %s: %w", id, result.Error)
	}
	return aval, nil
	// c.JSON(200, gin.H{"message": fmt.Sprintf("Successfully Pinged Peer: %s", peerID), "peer_in_dht": peerInDHT, "RTT": result.RTT})
}

// DEBUG ONLY
func OldPingPeerHandler(c *gin.Context) {
	peerID := c.Query("peerID")

	if peerID == "" {
		c.JSON(400, gin.H{"error": "peerID not provided"})
		return
	}
	if peerID == p2p.Host.ID().String() {
		c.JSON(400, gin.H{"error": "peerID can not be self peerID"})
		return
	}

	targetPeer, err := peer.Decode(peerID)
	if err != nil {
		zlog.Sugar().Errorf("Could not decode string ID to peerID: %v", err)
		c.JSON(400, gin.H{"error": "Could not decode string ID to peerID"})
		return
	}

	var peerInDHT bool
	_, err = p2p.Host.Peerstore().Get(targetPeer, "peer_info")
	if err != nil {
		peerInDHT = false
	} else {
		peerInDHT = true
	}

	result := OldPingPeer(c.Request.Context(), p2p.Host, targetPeer)
	zlog.Sugar().Infof("Pinged %s --> RTT: %s", targetPeer.String(), result.RTT)
	if result.Success {
		c.JSON(200, gin.H{"message": fmt.Sprintf("Successfully Pinged Peer: %s", peerID), "peer_in_dht": peerInDHT, "RTT": result.RTT})
	} else {
		c.JSON(400, gin.H{"message": fmt.Sprintf("Could not ping peer: %s -- %s", peerID, result.Error), "peer_in_dht": peerInDHT, "RTT": result.RTT})
		return
	}
}

// DEBUG ONLY
func DumpKademliaDHT(ctx context.Context) ([]models.PeerData, error) {
	if p2p.Host == nil {
		return nil, fmt.Errorf("host node has not yet been initialized")
	}

	dhtContentChan := make(chan models.PeerData, len(p2p.peers))
	tasks := make(chan peer.AddrInfo, len(p2p.peers))

	var wg sync.WaitGroup
	workerCount := 5
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for peer := range tasks {
				var peerInfo models.PeerData
				var updates models.KadDHTMachineUpdate

				// Add custom namespace to the key
				namespacedKey := customNamespace + peer.ID.String()
				bytes, err := p2p.DHT.GetValue(ctx, namespacedKey)
				if err != nil {
					zlog.Sugar().Errorf(fmt.Sprintf("Couldn't retrieve dht content for peer: %s", peer.String()))
					continue
				}
				err = json.Unmarshal(bytes, &updates)
				if err != nil {
					zlog.Sugar().Errorf("Error unmarshalling value: %v", err)
				}
				err = json.Unmarshal(updates.Data, &peerInfo)
				if err != nil {
					zlog.Sugar().Errorf("Error unmarshalling value: %v", err)
					continue
				}
				dhtContentChan <- peerInfo
			}
		}()
	}

	// Send tasks to the workers
	for _, peer := range p2p.peers {
		tasks <- peer
	}
	close(tasks)

	wg.Wait()
	close(dhtContentChan)

	var dhtContent []models.PeerData
	for peerData := range dhtContentChan {
		dhtContent = append(dhtContent, peerData)
	}
	return dhtContent, nil
}

func InitiateTransferFile(ctx context.Context, w http.ResponseWriter, r *http.Request, id, path string) error {
	p, err := peer.Decode(id)
	if err != nil {
		zlog.Sugar().Errorf("could not decode string ID to peer ID: %w", err)
		return fmt.Errorf("could not decode string ID to peer ID: %w", err)
	}

	// upgrade to websocket and steam transfer progress
	ws, err := internal.UpgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		zlog.Sugar().Errorf("failed to set websocket upgrade: %w\n", err)
		return fmt.Errorf("failed to set websocket upgrade: %w", err)
	}
	// conn := internal.WebSocketConnection{Conn: ws}
	// clients[conn] = peerID

	transferCh, err := SendFileToPeer(ctx, p, path, FTMISC)
	if err != nil {
		zlog.Sugar().Errorf("error: could not send file to peer - %v", err)
		ws.Close()
		return fmt.Errorf("could not send file to peer: %w", err)
	}

	ws.WriteJSON(gin.H{"message": "File Transfer Initiated. Transfer will start when peer accepts it."})
	for p := range transferCh {
		ws.WriteJSON(gin.H{
			"remaining_time": fmt.Sprintf("%v seconds", p.Remaining().Round(time.Second)),
			"percentage":     fmt.Sprintf("%.2f %%", p.Percent()),
			"size":           fmt.Sprintf("%.2f MB", p.N()/1048576),
		})
	}
	ws.WriteMessage(1, []byte("transfer complete"))
	ws.Close()
	return nil
}
