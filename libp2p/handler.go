package libp2p

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"gitlab.com/nunet/device-management-service/internal"
	"gitlab.com/nunet/device-management-service/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

	if p2p.Host == nil {
		c.JSON(500, gin.H{"error": "Host Node hasn't yet been initialized."})
		return
	}
	var dhtPeers []peer.ID
	for _, peer := range p2p.Host.Peerstore().Peers() {
		_, err := p2p.Host.Peerstore().Get(peer, "peer_info")
		if err != nil {
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
		zlog.Sugar().Errorf("could not create stream with peer: %v", err)
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
		zlog.Sugar().Error(err)
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
		zlog.Sugar().Error(err)
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

	if p2p.Host == nil {
		c.JSON(500, gin.H{"error": "Host Node hasn't yet been initialized."})
		return
	}

	dhtContent := []models.PeerData{}
	for _, peer := range p2p.Host.Peerstore().Peers() {
		peerData, err := p2p.Host.Peerstore().Get(peer, "peer_info")
		if err != nil {
			zlog.Sugar().Infof("UpdateAvailableResources error: %s", err.Error())
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

	file, err := os.Open(filePath)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "couldn't open file"})
		zlog.Sugar().Errorf("couldn't open file: %v", err)
		return
	}

	if _, debugMode := os.LookupEnv("NUNET_DEBUG"); debugMode {
		zlog.Sugar().Debugf("sending '%s' to %s", filePath, peerID)
	}

	stream, err := p2p.Host.NewStream(c, p, protocol.ID(FileTransferProtocolID))
	if _, debugMode := os.LookupEnv("NUNET_DEBUG"); debugMode {
		zlog.Sugar().Debugf("stream : to %v", stream)
	}
	if err != nil {
		zlog.Sugar().Errorf("could not create stream with peer for file transfer: %v", err)
		c.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("Could not create stream with peer - %v", err)})
		return
	}

	w := bufio.NewWriter(stream)

	go FileReadStreamWrite(file, stream, w)
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

	file, err := os.Create("/tmp/libp2pfile")
	if err != nil {
		zlog.Sugar().Errorf("failed to create file: %v", err)
		return
	}

	r := bufio.NewReader(inboundFileStream)

	go StreamReadFileWrite(file, inboundFileStream, r)
}
