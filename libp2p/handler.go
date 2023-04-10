package libp2p

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"gitlab.com/nunet/device-management-service/internal"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var clients = make(map[internal.WebSocketConnection]string)

// ListPeers  godoc
// @Summary      Return list of peers currently connected to
// @Description  Gets a list of peers the libp2p node can see within the network and return a list of peers
// @Tags         p2p
// @Produce      json
// @Success      200  {string}	string
// @Router       /peers [get]
func ListPeers(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers"))
	span.SetAttributes(attribute.String("PeerID", p2p.Host.ID().String()))
	span.SetAttributes(attribute.String("MachineUUID", utils.GetMachineUUID()))

	if p2p.Host == nil {
		c.JSON(500, gin.H{"error": "Host Node hasn't yet been initialized."})
		return
	}
	peers, err := p2p.getPeers(c, "nunet")
	if err != nil {
		c.JSON(500, gin.H{"error": "can not fetch peers"})
		zlog.Sugar().Fatalf("Error Can Not Fetch Peers: %s\n", err.Error())
		return
	}
	c.JSON(200, peers)

}

// ListPeers  godoc
// @Summary      Return list of peers which have sent a dht update
// @Description  Gets a list of peers the libp2p node has received a dht update from
// @Tags         p2p
// @Produce      json
// @Success      200  {string}	string
// @Router       /peers/dht [get]
func ListDHTPeers(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/dht"))
	span.SetAttributes(attribute.String("PeerID", p2p.Host.ID().String()))
	span.SetAttributes(attribute.String("MachineUUID", utils.GetMachineUUID()))

	if p2p.Host == nil {
		c.JSON(500, gin.H{"error": "Host Node hasn't yet been initialized."})
		return
	}
	var dhtPeers []peer.ID
	for _, peer := range p2p.Host.Peerstore().Peers() {
		_, err := p2p.Host.Peerstore().Get(peer, "peer_info")
		if err != nil {
			zlog.ErrorContext(c.Request.Context(), fmt.Sprintf("coultn't retrieve dht content for peer: %s", peer.String()), zap.Error(err))
			continue
		}
		if peer == p2p.Host.ID() {
			continue
		}
		dhtPeers = append(dhtPeers, peer)
	}

	if len(dhtPeers) == 0 {
		c.JSON(200, gin.H{"message": "No peers found"})
		return
	}

	dhtPeersJ, err := json.Marshal(dhtPeers)
	if err != nil {
		zlog.ErrorContext(c.Request.Context(), "failed to json marshal dhtPeers: %v", zap.Error(err))
	}

	span.SetAttributes(attribute.String("Response", string(dhtPeersJ)))
	c.JSON(200, dhtPeers)
}

// SelfPeerInfo  godoc
// @Summary      Return self peer info
// @Description  Gets self peer info of libp2p node
// @Tags         p2p
// @Produce      json
// @Success      200  {string}	string
// @Router       /peers/self [get]
func SelfPeerInfo(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/self"))
	span.SetAttributes(attribute.String("PeerID", p2p.Host.ID().String()))
	span.SetAttributes(attribute.String("MachineUUID", utils.GetMachineUUID()))

	if p2p.Host == nil {
		c.JSON(500, gin.H{"error": "Host Node hasn't yet been initialized."})
		return
	}

	out := struct {
		ID    string
		Addrs []multiaddr.Multiaddr
	}{
		p2p.Host.ID().String(),
		p2p.Host.Addrs(),
	}

	c.JSON(200, out)
}

// ListChatHandler  godoc
// @Summary      List chat requests
// @Description  Get a list of chat requests from peers
// @Tags         chat
// @Produce      json
// @Success      200
// @Router       /peers/chat [get]
func ListChatHandler(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/chat"))
	span.SetAttributes(attribute.String("PeerID", p2p.Host.ID().String()))
	span.SetAttributes(attribute.String("MachineUUID", utils.GetMachineUUID()))

	chatRequests, err := incomingChatRequests()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, chatRequests)
}

// ClearChatHandler  godoc
// @Summary      Clear chat requests
// @Description  Clear chat request streams from peers
// @Tags         chat
// @Produce      json
// @Success      200
// @Router       /peers/chat/clear [get]
func ClearChatHandler(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/chat/clear"))
	span.SetAttributes(attribute.String("PeerID", p2p.Host.ID().String()))
	span.SetAttributes(attribute.String("MachineUUID", utils.GetMachineUUID()))

	if err := clearIncomingChatRequests(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Successfully Cleard Inbound Chat Requests."})
}

// StartChatHandler  godoc
// @Summary      Start chat with a peer
// @Description  Start chat session with a peer
// @Tags         chat
// @Success      200
// @Router       /peers/chat/start [get]
func StartChatHandler(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/chat/start"))
	span.SetAttributes(attribute.String("PeerID", p2p.Host.ID().String()))
	span.SetAttributes(attribute.String("MachineUUID", utils.GetMachineUUID()))

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
		zlog.Sugar().Errorf("Could not decode string ID to peerID:", err)
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

	welcomeMessage := fmt.Sprintf("Enter the message that you wish to send to %s and press return.", peerID)

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
// @Summary      Join chat with a peer
// @Description  Join a chat session started by a peer
// @Tags         chat
// @Success      200
// @Router       /peers/chat/join [get]
func JoinChatHandler(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/chat/join"))
	span.SetAttributes(attribute.String("PeerID", p2p.Host.ID().String()))
	span.SetAttributes(attribute.String("MachineUUID", utils.GetMachineUUID()))

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
// @Summary      Return a dump of the dht
// @Description  Returns entire DHT content
// @Tags         p2p
// @Produce      json
// @Success      200  {string}	string
// @Router       /dht [get]
func DumpDHT(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/dht"))
	span.SetAttributes(attribute.String("PeerID", p2p.Host.ID().String()))
	span.SetAttributes(attribute.String("MachineUUID", utils.GetMachineUUID()))

	if p2p.Host == nil {
		c.JSON(500, gin.H{"error": "Host Node hasn't yet been initialized."})
		return
	}

	dhtContent := []models.PeerData{}
	for _, peer := range p2p.Host.Peerstore().Peers() {
		peerData, err := p2p.Host.Peerstore().Get(peer, "peer_info")
		if err != nil {
			zlog.ErrorContext(c.Request.Context(), fmt.Sprintf("Coultn't retrieve dht content for peer: %s", peer.String()), zap.Error(err))
		}
		if peer == p2p.Host.ID() {
			continue
		}
		if Data, ok := peerData.(models.PeerData); ok {
			dhtContent = append(dhtContent, models.PeerData(Data))
		}
	}

	if len(dhtContent) == 0 {
		c.JSON(200, gin.H{"message": "No Content in DHT"})
		return
	}

	dhtContentJ, err := json.Marshal(dhtContent)
	if err != nil {
		zlog.ErrorContext(c.Request.Context(), "failed to json marshal dhtContent: %v", zap.Error(err))
	}

	span.SetAttributes(attribute.String("Response", string(dhtContentJ)))
	c.JSON(200, dhtContent)
}
