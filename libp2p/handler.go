package libp2p

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"gitlab.com/nunet/device-management-service/internal"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/klogger"
	kLogger "gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/nunet/device-management-service/models"
	"go.uber.org/zap"
)

var clients = make(map[internal.WebSocketConnection]string)

// DeviceStatusHandler  godoc
//
// @Summary		    Retrieve device status
// @Description	    Retrieve device status whether paused/offline (unable to receive job deployments) or online
// @Tags			device
// @Produce		    json
// @Success		    200	{string}	string
// @Router			/device/status [get]
func DeviceStatusHandler(c *gin.Context) {
	if p2p.Host == nil {
		c.JSON(500, gin.H{"error": "Host Node hasn't yet been initialized."})
		return
	}

	// Fetch the raw peer data
	peerDataRaw, err := p2p.Host.Peerstore().Get(p2p.Host.ID(), "peer_info")
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Couldn't retrieve data - error: %v - Try again in a few minutes", err)})
		return
	}

	peerData, ok := peerDataRaw.(models.PeerData)
	if !ok {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to type assert peer data for peer: %s", p2p.Host.ID().String())})
		return
	}

	c.JSON(200, gin.H{"online": peerData.IsAvailable})
}

// DeviceStatusChangeHandler  godoc
//
// @Summary		    Change device status between online/offline
// @Description	    Change device status to online (able to receive jobs) or offline (unable to receive jobs).
// @Tags			device
// @Produce		    json
// @Success		    200	{string}	string
// @Router			/device/status [post]
func DeviceStatusChangeHandler(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/device/pause"))
	span.SetAttributes(attribute.String("MachineUUID", utils.GetMachineUUID()))
	kLogger.Info("Pause job onboarding", span)

	if p2p.Host == nil {
		c.JSON(500, gin.H{"error": "Host Node hasn't yet been initialized."})
		return
	}

	// Fetch the raw peer data
	peerDataRaw, err := p2p.Host.Peerstore().Get(p2p.Host.ID(), "peer_info")
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Couldn't retrieve data - error: %v - Try again in a few minutes", err)})
		return
	}

	peerData, ok := peerDataRaw.(models.PeerData)
	if !ok {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to type assert peer data for peer: %s", p2p.Host.ID().String())})
		return
	}

	var deviceStatus struct {
		IsAvailable bool `json:"is_available"`
	}

	if c.Request.ContentLength == 0 {
		c.JSON(500, gin.H{"error": "no data provided"})
		return
	}
	if err := c.ShouldBindJSON(&deviceStatus); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if peerData.IsAvailable == deviceStatus.IsAvailable {
		c.JSON(500, gin.H{"error": "no change in device status"})
		return
	}

	peerData.IsAvailable = deviceStatus.IsAvailable
	p2p.Host.Peerstore().Put(p2p.Host.ID(), "peer_info", peerData)

	UpdateKadDHT()

	if deviceStatus.IsAvailable {
		c.JSON(200, gin.H{"message": "Device status successfully changed to online"})
	} else {
		c.JSON(200, gin.H{"message": "Device status successfully changed to offline"})
	}
}

// ListPeers  godoc
//
//	@Summary		Return list of peers currently connected to
//	@Description	Gets a list of peers the libp2p node can see within the network and return a list of peers
//	@Tags			p2p
//	@Produce		json
//	@Success		200	{string}	string
//	@Router			/peers [get]
func ListPeers(c *gin.Context) {
	if p2p.Host == nil {
		c.JSON(500, gin.H{"error": "Host Node hasn't yet been initialized."})
		return
	}
	klogger.Logger.Info("List peers executed by " + p2p.Host.ID().String())
	if p2p.peers == nil {
		c.JSON(500, gin.H{"error": "Peers haven't yet been fetched."})
		klogger.Logger.Error("Peers haven't yet been fetched.")

		return
	}
	if len(p2p.peers) == 0 {
		c.JSON(500, gin.H{"error": "No Peers Yet."})
		klogger.Logger.Error("No Peers Yet.")

		return
	}

	klogger.Logger.Info("List of peers" + p2p.Host.ID().String())

	c.JSON(200, p2p.peers)

}

// ListPeers  godoc
// ssss
//
//	@Summary		Return list of peers which have sent a dht update
//	@Description	Gets a list of peers the libp2p node has received a dht update from
//	@Tags			p2p
//	@Produce		json
//	@Success		200	{string}	string
//	@Router			/peers/dht [get]
func ListDHTPeers(c *gin.Context) {
	if p2p.Host == nil {
		c.JSON(500, gin.H{"error": "Host Node hasn't yet been initialized."})
		return
	}
	klogger.Logger.Info("List DHT peers executed by " + p2p.Host.ID().String())

	var dhtPeers []peer.ID
	for _, peer := range p2p.peers {
		_, err := p2p.Host.Peerstore().Get(peer.ID, "peer_info")
		if err != nil {
			if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
				zlog.ErrorContext(c.Request.Context(), fmt.Sprintf("coultn't retrieve dht content for peer: %s", peer.String()), zap.Error(err))
			}
			continue
		}
		if peer.ID == p2p.Host.ID() {
			continue
		}
		dhtPeers = append(dhtPeers, peer.ID)
	}

	if len(dhtPeers) == 0 {
		c.JSON(200, gin.H{"message": "No peers found"})
		klogger.Logger.Error("No peers found")
		return
	}

	dhtPeersJ, err := json.Marshal(dhtPeers)
	if err != nil {
		zlog.ErrorContext(c.Request.Context(), "failed to json marshal dhtPeers: %v", zap.Error(err))
		klogger.Logger.Error("failed to json marshal dhtPeers")
	}
	klogger.Logger.Info("Response: " + string(dhtPeersJ))

	c.JSON(200, dhtPeers)
}

// ListKadDHTPeers  godoc
//
//	@Summary		Return list of peers which have sent a dht update
//	@Description	Gets a list of peers the libp2p node has received a dht update from
//	@Tags			p2p
//	@Produce		json
//	@Success		200	{string}	string
//	@Router			/peers/kad-dht [get]
func ListKadDHTPeers(c *gin.Context) {
	if p2p.Host == nil {
		c.JSON(500, gin.H{"error": "Host Node hasn't yet been initialized."})
		return
	}
	klogger.Logger.Info("List Kademlia DHT peers executed by " + p2p.Host.ID().String())

	var dhtPeers []string

	for _, peer := range p2p.peers {
		var updates models.KadDHTMachineUpdate
		var peerInfo models.PeerData

		// Add custom namespace to the key
		namespacedKey := customNamespace + peer.ID.String()
		bytes, err := p2p.DHT.GetValue(c, namespacedKey)
		if err != nil {
			if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
				zlog.Sugar().Errorf(fmt.Sprintf("Couldn't retrieve dht content for peer: %s", peer.String()))
			}
			continue
		}
		err = json.Unmarshal(bytes, &updates)
		if err != nil {
			if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
				zlog.Sugar().Errorf("Error unmarshalling value: %v", err)

			}
			continue
		}
		err = json.Unmarshal(updates.Data, &peerInfo)
		if err != nil {
			if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
				zlog.Sugar().Errorf("Error unmarshalling value: %v", err)
			}
			continue
		}

		dhtPeers = append(dhtPeers, peerInfo.PeerID)
	}

	if len(dhtPeers) == 0 {
		c.JSON(200, gin.H{"message": "No peers found"})
		klogger.Logger.Error("No peers found")
		return
	}

	dhtPeersJ, err := json.Marshal(dhtPeers)
	if err != nil {
		zlog.ErrorContext(c.Request.Context(), "failed to json marshal dhtPeers: %v", zap.Error(err))
	}
	klogger.Logger.Info("Response: " + string(dhtPeersJ))

	c.JSON(200, dhtPeers)
}

// SelfPeerInfo  godoc
//
//	@Summary		Return self peer info
//	@Description	Gets self peer info of libp2p node
//	@Tags			p2p
//	@Produce		json
//	@Success		200	{string}	string
//	@Router			/peers/self [get]
func SelfPeerInfo(c *gin.Context) {
	stTime := time.Now().UnixMilli()
	if p2p.Host == nil {
		c.JSON(500, gin.H{"error": "Host Node hasn't yet been initialized."})
		return
	}
	klogger.Logger.Info(" result : Self Peer ID " + p2p.Host.ID().String())
	zlog.Sugar().Infof("----------klogger time taken=%d ms", time.Now().UnixMilli()-stTime)

	out := struct {
		ID    string
		Addrs []multiaddr.Multiaddr
	}{
		p2p.Host.ID().String(),
		p2p.Host.Addrs(),
	}
	zlog.Sugar().Infof("----------overall time taken=%d ms", time.Now().UnixMilli()-stTime)

	c.JSON(200, out)
}

// ListChatHandler  godoc
//
//	@Summary		List chat requests
//	@Description	Get a list of chat requests from peers
//	@Tags			chat
//	@Produce		json
//	@Success		200
//	@Router			/peers/chat [get]
func ListChatHandler(c *gin.Context) {
	chatRequests, err := IncomingChatRequests()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		klogger.Logger.Error("List chat handler Error: " + err.Error())
		return
	}
	c.JSON(200, chatRequests)
}

// ClearChatHandler  godoc
//
//	@Summary		Clear chat requests
//	@Description	Clear chat request streams from peers
//	@Tags			chat
//	@Produce		json
//	@Success		200
//	@Router			/peers/chat/clear [get]
func ClearChatHandler(c *gin.Context) {
	if err := ClearIncomingChatRequests(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		klogger.Logger.Error("Clear chat handler Error: " + err.Error())
		return
	}

	c.JSON(200, gin.H{"message": "Successfully Cleard Inbound Chat Requests."})
}

// StartChatHandler  godoc
//
//	@Summary		Start chat with a peer
//	@Description	Start chat session with a peer
//	@Tags			chat
//	@Success		200
//	@Router			/peers/chat/start [get]
func StartChatHandler(c *gin.Context) {
	peerID := c.Query("peerID")

	if len(peerID) == 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": "peerID not provided"})
		return
	}
	if peerID == p2p.Host.ID().String() {
		c.AbortWithStatusJSON(400, gin.H{"error": "peerID can not be self peerID"})
		return
	}

	p, err := peer.Decode(peerID)
	if err != nil {
		zlog.Sugar().Errorf("could not decode string ID to peerID: %v", err)
		c.AbortWithStatusJSON(400, gin.H{"error": "Could not decode string ID to peerID"})
		return
	}

	stream, err := p2p.Host.NewStream(c, p, protocol.ID(ChatProtocolID))
	if err != nil {
		zlog.Sugar().ErrorfContext(c.Request.Context(), "Could not create stream with peer - %v", err)
		c.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Could not create stream with peer - %v", err)})
		return
	}

	ws, err := internal.UpgradeConnection.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Sugar().Errorf("Failed to set websocket upgrade: %+v\n", err)
		return
	}

	welcomeMessage := fmt.Sprintf("Enter the message that you wish to send to %s with stream ID: %s and press return.", peerID, stream.ID())

	err = ws.WriteMessage(websocket.TextMessage, []byte(welcomeMessage))
	if err != nil {
		zlog.Sugar().Errorf(err.Error())
	}

	conn := internal.WebSocketConnection{Conn: ws}
	clients[conn] = peerID

	r := bufio.NewReader(stream)
	w := bufio.NewWriter(stream)

	go SockReadStreamWrite(&conn, stream, w)
	go StreamReadSockWrite(&conn, stream, r)
}

// JoinChatHandler  godoc
//
//	@Summary		Join chat with a peer
//	@Description	Join a chat session started by a peer
//	@Tags			chat
//	@Success		200
//	@Router			/peers/chat/join [get]
func JoinChatHandler(c *gin.Context) {
	streamReqID := c.Query("streamID")
	if streamReqID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Stream ID not provided"})
		return
	}

	streamID, err := strconv.Atoi(streamReqID)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Invalid Stream ID: %v", streamReqID)})
		return
	}

	if streamID >= len(inboundChatStreams) {
		c.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Unknown Stream ID: %v", streamID)})
		return
	}

	ws, err := internal.UpgradeConnection.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Sugar().Errorf("Failed to set websocket upgrade: %+v\n", err)
		return
	}

	welcomeMessage := "Enter the message that you wish to send and press return."

	err = ws.WriteMessage(websocket.TextMessage, []byte(welcomeMessage))
	if err != nil {
		zlog.Sugar().Errorf(err.Error())
	}

	conn := internal.WebSocketConnection{Conn: ws}
	clients[conn] = streamReqID

	stream := inboundChatStreams[int(streamID)]
	copy(inboundChatStreams[streamID:], inboundChatStreams[streamID+1:])
	inboundChatStreams[len(inboundChatStreams)-1] = nil
	inboundChatStreams = inboundChatStreams[:len(inboundChatStreams)-1]

	r := bufio.NewReader(stream)
	w := bufio.NewWriter(stream)

	go SockReadStreamWrite(&conn, stream, w)
	go StreamReadSockWrite(&conn, stream, r)
}

// DumpDHT  godoc
//
//	@Summary		Return a dump of the dht
//	@Description	Returns entire DHT content
//	@Tags			p2p
//	@Produce		json
//	@Success		200	{string}	string
//	@Router			/dht [get]
func DumpDHT(c *gin.Context) {
	if p2p.Host == nil {
		c.JSON(500, gin.H{"error": "Host Node hasn't yet been initialized."})
		return
	}

	dhtContent := []models.PeerData{}
	for _, peer := range p2p.Host.Peerstore().Peers() {
		peerData, err := p2p.Host.Peerstore().Get(peer, "peer_info")
		if err != nil {
			if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
				zlog.ErrorContext(c.Request.Context(), fmt.Sprintf("Coultn't retrieve dht content for peer: %s", peer.String()), zap.Error(err))
			}
		}

		if Data, ok := peerData.(models.PeerData); ok {
			dhtContent = append(dhtContent, models.PeerData(Data))
		}
	}

	if len(dhtContent) == 0 {
		c.JSON(200, gin.H{"message": "No Content in DHT"})
		return
	}

	c.JSON(200, dhtContent)
}

// DefaultDepReqPeer  godoc
//
//	@Summary		Manage default deplyment request receiver peer
//	@Description	Set peer as the default receipient of deployment requests by setting the peerID parameter on GET request.
//	@Description	Show peer set as default deployment request receiver by sending a GET request without any parameters.
//	@Description	Remove default deployment request receiver by sending a GET request with peerID parameter set to '0'.
//	@Tags			peers
//	@Success		200
//	@Router			/peers/depreq [get]
func DefaultDepReqPeer(c *gin.Context) {
	peerID := c.Query("peerID")

	if peerID == "0" {
		config.SetConfig("job.target_peer", "")
		c.JSON(200, gin.H{"message": "default peer successfully removed"})
		return
	}
	if peerID == "" {
		if config.GetConfig().Job.TargetPeer == "" {
			c.JSON(200, gin.H{"message": "no default peer set"})
		} else {
			c.JSON(200, gin.H{"message": fmt.Sprintf("default peer is %s", config.GetConfig().Job.TargetPeer)})
		}
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

	_, err = p2p.Host.Peerstore().Get(targetPeer, "peer_info")
	if err != nil {
		c.JSON(400, gin.H{"error": "Peer not in DHT yet.\nPlease use peerID from DHT Peers section of 'nunet peer list'"})
		return
	}

	pingResult, pingCancel := Ping(c.Request.Context(), targetPeer)
	defer pingCancel()
	result := <-pingResult
	if result.Error == nil {
		config.SetConfig("job.target_peer", peerID)
		c.JSON(200, gin.H{"message": fmt.Sprintf("Successfully set %s as default deployment request receiver.", peerID)})
	} else {
		zlog.Sugar().Errorf("Could not ping peer: %v", result.Error)
		c.JSON(400, gin.H{"error": "Peer not online."})
		return
	}
}

// DEBUG ONLY
func ManualDHTUpdateHandler(c *gin.Context) {
	go UpdateKadDHT()
	GetDHTUpdates(c)
}

// DEBUG ONLY
func CleanupPeerHandler(c *gin.Context) {
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
	p2p.Host.Peerstore().RemovePeer(targetPeer)
	c.JSON(200, gin.H{"message": fmt.Sprintf("Successfully cleaned up peer: %s", peerID)})
}

// DEBUG ONLY
func PingPeerHandler(c *gin.Context) {
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

	pingResult, pingCancel := Ping(c.Request.Context(), targetPeer)
	defer pingCancel()
	result := <-pingResult
	zlog.Sugar().Infof("Pinged %s --> RTT: %s", targetPeer.String(), result.RTT)
	if result.Error == nil {
		c.JSON(200, gin.H{"message": fmt.Sprintf("Successfully Pinged Peer: %s", peerID), "peer_in_dht": peerInDHT, "RTT": result.RTT})
	} else {
		c.JSON(400, gin.H{"message": fmt.Sprintf("Could not ping peer: %s -- %s", peerID, result.Error), "peer_in_dht": peerInDHT, "RTT": result.RTT})
	}
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

	result := PingPeer(c.Request.Context(), p2p.Host, targetPeer)
	zlog.Sugar().Infof("Pinged %s --> RTT: %s", targetPeer.String(), result.RTT)
	if result.Success {
		c.JSON(200, gin.H{"message": fmt.Sprintf("Successfully Pinged Peer: %s", peerID), "peer_in_dht": peerInDHT, "RTT": result.RTT})
	} else {
		c.JSON(400, gin.H{"message": fmt.Sprintf("Could not ping peer: %s -- %s", peerID, result.Error), "peer_in_dht": peerInDHT, "RTT": result.RTT})
		return
	}
}

// DEBUG ONLY
func DumpKademliaDHT(c *gin.Context) {
	if p2p.Host == nil {
		c.JSON(500, gin.H{"error": "Host Node hasn't yet been initialized."})
		return
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
				bytes, err := p2p.DHT.GetValue(c.Request.Context(), namespacedKey)
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

	if len(dhtContent) == 0 {
		c.JSON(200, gin.H{"message": "No Content in DHT"})
		return
	}

	c.JSON(200, dhtContent)
}

// ListFileRequestsHandler  godoc
// @Summary      List file transfer requests
// @Description  Get a list of file transfer requests from peers
// @Tags         file
// @Produce      json
// @Success      200
// @Router       /peers/file [get]
func ListFileRequestsHandler(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/file"))

	fileRequests, err := incomingFileTransferRequests()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, fileRequests)
}

// ClearFileTransferRequestseHandler  godoc
// @Summary      Clear file transfer requests
// @Description  Clear file transfer request streams from peers
// @Tags         file
// @Produce      json
// @Success      200
// @Router       /peers/file/clear [get]
func ClearFileTransferRequestseHandler(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/file/clear"))

	if err := clearIncomingFileRequests(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Successfully Cleard Inbound File Transfer Requests."})
}

// InitiateFileTransferHandler  godoc
// @Summary      Send a file to a peer
// @Description  Initiate file transfer to a peer. filePath and peerID are required arguments.
// @Tags         file
// @Success      200
// @Router       /peers/file/send [get]
func InitiateFileTransferHandler(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/file/send"))

	peerID := c.Query("peerID")

	if len(peerID) == 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": "peerID not provided"})
		return
	}
	if peerID == p2p.Host.ID().String() {
		c.AbortWithStatusJSON(400, gin.H{"error": "peerID can not be self peerID"})
		return
	}

	p, err := peer.Decode(peerID)
	if err != nil {
		zlog.Sugar().Errorf("could not decode string ID to peerID: %v", err)
		c.AbortWithStatusJSON(400, gin.H{"error": "Could not decode string ID to peerID"})
		return
	}

	filePath := c.Query("filePath")
	if len(filePath) == 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": "filePath not provided"})
		return
	}

	// upgrade to websocket and steam transfer progress
	ws, err := internal.UpgradeConnection.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Sugar().Errorf("Failed to set websocket upgrade: %+v\n", err)
		return
	}
	conn := internal.WebSocketConnection{Conn: ws}
	clients[conn] = peerID

	transferChan, err := sendFileToPeer(c, p, filePath)
	if err != nil {
		zlog.Sugar().Errorf("error: couldn't send file to peer - %v", err)
		ws.WriteJSON(gin.H{"error": err.Error()})
		ws.Close()
		return
	}

	ws.WriteJSON(gin.H{"message": "File Transfer Initiated. Transfer will start when peer accepts it."})
	for p := range transferChan {
		ws.WriteJSON(gin.H{
			"remaining_time": fmt.Sprintf("%v seconds", p.Remaining().Round(time.Second)),
			"percentage":     fmt.Sprintf("%.2f %%", p.Percent()),
			"size":           fmt.Sprintf("%.2f MB", p.N()/1048576),
		})
	}
	ws.WriteMessage(1, []byte("transfer complete"))
	ws.Close()

}

// AcceptFileTransferHandler  godoc
// @Summary      Accept incoming file transfer
// @Description  Accept an incoming file transfer. Incoming file transfre stream id is a required parameter.
// @Tags         file
// @Success      200
// @Router       /peers/file/accept [get]
func AcceptFileTransferHandler(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/file/accept"))

	streamReqID := c.Query("streamID")
	if streamReqID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Stream ID not provided"})
		return
	}

	streamID, err := strconv.Atoi(streamReqID)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Invalid Stream ID: %v", streamReqID)})
		return
	}

	if streamID != 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Unknown Stream ID: %v", streamID)})
		return
	}

	err = acceptFileTransfer()
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
}
