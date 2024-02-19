package api

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p/core/peer"
	"gitlab.com/nunet/device-management-service/internal/klogger"
	"gitlab.com/nunet/device-management-service/libp2p"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ListPeersHandler  godoc
//
//	@Summary		Return list of peers currently connected to
//	@Description	Gets a list of peers the libp2p node can see within the network and return a list of peers
//	@Tags			p2p
//	@Produce		json
//	@Success		200	{string}	string
//	@Router			/peers [get]
func ListPeersHandler(c *gin.Context) {
	peers, err := libp2p.ListPeers()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, peers)
}

// ListDHTPeersHandler  godoc
//
//	@Summary		Return list of peers which have sent a dht update
//	@Description	Gets a list of peers the libp2p node has received a dht update from
//	@Tags			p2p
//	@Produce		json
//	@Success		200	{string}	string
//	@Router			/peers/dht [get]
func ListDHTPeersHandler(c *gin.Context) {
	reqCtx := c.Request.Context()
	peers, err := libp2p.ListDHTPeers(reqCtx)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if len(peers) == 0 {
		c.JSON(200, gin.H{"message": "no peers found"})
		return
	}
	c.JSON(200, peers)
}

// ListKadDHTPeersHandler  godoc
//
//	@Summary		Return list of peers which have sent a dht update
//	@Description	Gets a list of peers the libp2p node has received a dht update from
//	@Tags			p2p
//	@Produce		json
//	@Success		200	{string}	string
//	@Router			/peers/kad-dht [get]
func ListKadDHTPeersHandler(c *gin.Context) {
	reqCtx := c.Request.Context()

	peers, err := libp2p.ListKadDHTPeers(c, reqCtx)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if len(peers) == 0 {
		c.JSON(200, gin.H{"message": "no peers found"})
		klogger.Logger.Error("No peers found")
		return
	}
	c.JSON(200, peers)
}

// SelfPeerInfoHandler  godoc
//
//	@Summary		Return self peer info
//	@Description	Gets self peer info of libp2p node
//	@Tags			p2p
//	@Produce		json
//	@Success		200	{string}	string
//	@Router			/peers/self [get]
func SelfPeerInfoHandler(c *gin.Context) {
	self, err := libp2p.SelfPeerInfo()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, self)
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
	chats, err := libp2p.IncomingChatRequests()
	if err != nil {
		klogger.Logger.Error("List chat handler Error: " + err.Error())
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, chats)
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
	err := libp2p.ClearIncomingChatRequests()
	if err != nil {
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
	reqCtx := c.Request.Context()
	id := c.Query("peerID")

	stream, err := libp2p.CreateChatStream(reqCtx, id)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	libp2p.StartChat(c.Writer, c.Request, stream, id)
}

// JoinChatHandler  godoc
//
//	@Summary		Join chat with a peer
//	@Description	Join a chat session started by a peer
//	@Tags			chat
//	@Success		200
//	@Router			/peers/chat/join [get]
func JoinChatHandler(c *gin.Context) {
	streamID := c.Query("streamID")
	if streamID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "stream ID not provided"})
		return
	}
	stream, err := strconv.Atoi(streamID)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("invalid stream ID: %w", err)})
		return
	}
	libp2p.JoinChat(c.Writer, c.Request, stream)
}

// DumpDHTHandler  godoc
//
//	@Summary		Return a dump of the dht
//	@Description	Returns entire DHT content
//	@Tags			p2p
//	@Produce		json
//	@Success		200	{string}	string
//	@Router			/dht [get]
func DumpDHTHandler(c *gin.Context) {
	reqCtx := c.Request.Context()
	dht, err := libp2p.DumpDHT(reqCtx)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if len(dht) == 0 {
		c.JSON(200, gin.H{"message": "empty DHT"})
		return
	}
	c.JSON(200, dht)
}

// DefaultDepReqPeerHandler  godoc
//
//	@Summary		Manage default deplyment request receiver peer
//	@Description	Set peer as the default receipient of deployment requests by setting the peerID parameter on GET request.
//	@Description	Show peer set as default deployment request receiver by sending a GET request without any parameters.
//	@Description	Remove default deployment request receiver by sending a GET request with peerID parameter set to '0'.
//	@Tags			peers
//	@Success		200
//	@Router			/peers/depreq [get]
func DefaultDepReqPeerHandler(c *gin.Context) {
	id := c.Query("peerID")
	reqCtx := c.Request.Context()

	msg, err := libp2p.DefaultDepReqPeer(reqCtx, id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": msg})
}

// ClearFileTransferRequestsHandler  godoc
// @Summary      Clear file transfer requests
// @Description  Clear file transfer request streams from peers
// @Tags         file
// @Produce      json
// @Success      200
// @Router       /peers/file/clear [get]
func ClearFileTransferRequestsHandler(c *gin.Context) {
	reqCtx := c.Request.Context()
	span := trace.SpanFromContext(reqCtx)
	span.SetAttributes(attribute.String("URL", "/peers/file/clear"))

	err := libp2p.ClearIncomingFileRequests()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "successfully cleared inbound file transfer requests"})
}

// ListFileTransferRequestsHandler  godoc
// @Summary      List file transfer requests
// @Description  Get a list of file transfer requests from peers
// @Tags         file
// @Produce      json
// @Success      200
// @Router       /peers/file [get]
func ListFileTransferRequestsHandler(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/file"))

	req, err := libp2p.IncomingFileTransferRequests()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, req)
}

// SendFileTransferHandler  godoc
// @Summary      Send a file to a peer
// @Description  Initiate file transfer to a peer. filePath and peerID are required arguments.
// @Tags         file
// @Success      200
// @Router       /peers/file/send [get]
func SendFileTransferHandler(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/file/send"))

	id := c.Query("peerID")
	if len(id) == 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": "peer ID not provided"})
		return
	}
	if id == libp2p.GetP2P().Host.ID().String() {
		c.AbortWithStatusJSON(400, gin.H{"error": "peer ID cannot be self peer ID"})
		return
	}
	p, err := peer.Decode(id)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid peer string ID: could not decode string ID to peer ID"})
		return
	}

	path := c.Query("filePath")
	if len(path) == 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": "filepath not provided"})
		return
	}

	err = libp2p.InitiateTransferFile(c, c.Writer, c.Request, p, path)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, nil)
}

// AcceptFileTransferHandler  godoc
// @Summary      Accept incoming file transfer
// @Description  Accept an incoming file transfer. Incoming file transfer stream ID is a required parameter.
// @Tags         file
// @Success      200
// @Router       /peers/file/accept [get]
func AcceptFileTransferHandler(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/file/accept"))

	streamID := c.Query("streamID")
	if streamID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "stream ID not provided"})
		return
	}

	stream, err := strconv.Atoi(streamID)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("invalid stream ID: %s", streamID)})
		return
	}

	if stream != 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": fmt.Sprintf("unknown stream ID: %d", stream)})
		return
	}

	err = libp2p.AcceptPeerFileTransfer(c, c.Writer, c.Request)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not accept file transfer with peer"})
		return
	}
	c.JSON(200, nil)
}
