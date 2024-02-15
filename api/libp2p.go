package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/internal"
	"gitlab.com/nunet/device-management-service/internal/klogger"
	kLogger "gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MISSING: Fix zlog bug

// HandleDeviceStatus  godoc
//
// @Summary		    Retrieve device status
// @Description	    Retrieve device status whether paused/offline (unable to receive job deployments) or online
// @Tags			device
// @Produce		    json
// @Success		    200	{string}	string
// @Router			/device/status [get]
func HandleDeviceStatus(c *gin.Context) {
	status, err := libp2p.DeviceStatus()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"online": status})
}

// HandleChangeDeviceStatus  godoc
//
// @Summary		    Change device status between online/offline
// @Description	    Change device status to online (able to receive jobs) or offline (unable to receive jobs).
// @Tags			device
// @Produce		    json
// @Success		    200	{string}	string
// @Router			/device/status [post]
func HandleChangeDeviceStatus(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/device/pause"))
	span.SetAttributes(attribute.String("MachineUUID", utils.GetMachineUUID()))
	kLogger.Info("Pause job onboarding", span)

	if c.Request.ContentLength == 0 {
		c.JSON(400, gin.H{"error": "no data provided"})
		return
	}

	var status struct {
		IsAvailable bool `json:"is_available"`
	}

	err := c.ShouldBindJSON(&status)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	reqCtx := c.Request.Context()
	err = libp2p.ChangeDeviceStatus(reqCtx, status.IsAvailable)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if status.IsAvailable {
		c.JSON(200, gin.H{"message": "device status successfully changed to online"})
	} else {
		c.JSON(200, gin.H{"message": "device status successfully changed to offline"})
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
func HandleListPeers(c *gin.Context) {
	peers, err := libp2p.ListPeers()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}
	c.JSON(200, peers)
}

// SelfPeerInfo  godoc
//
//	@Summary		Return self peer info
//	@Description	Gets self peer info of libp2p node
//	@Tags			p2p
//	@Produce		json
//	@Success		200	{string}	string
//	@Router			/peers/self [get]
func HandleSelfPeerInfo(c *gin.Context) {
	self, err := libp2p.SelfPeerInfo()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, self)
}

// HandleListChat  godoc
//
//	@Summary		List chat requests
//	@Description	Get a list of chat requests from peers
//	@Tags			chat
//	@Produce		json
//	@Success		200
//	@Router			/peers/chat [get]
func HandleListChat(c *gin.Context) {
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

// HandleDumpDHT  godoc
//
//	@Summary		Return a dump of the dht
//	@Description	Returns entire DHT content
//	@Tags			p2p
//	@Produce		json
//	@Success		200	{string}	string
//	@Router			/dht [get]
func HandleDumpDHT(c *gin.Context) {
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

// HandleClearFileTransferRequests  godoc
// @Summary      Clear file transfer requests
// @Description  Clear file transfer request streams from peers
// @Tags         file
// @Produce      json
// @Success      200
// @Router       /peers/file/clear [get]
func HandleClearFileTransferRequests(c *gin.Context) {
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

// HandleListDHTPeers  godoc
//
//	@Summary		Return list of peers which have sent a dht update
//	@Description	Gets a list of peers the libp2p node has received a dht update from
//	@Tags			p2p
//	@Produce		json
//	@Success		200	{string}	string
//	@Router			/peers/dht [get]
func HandleListDHTPeers(c *gin.Context) {
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

// HandleListKadDHTPeers  godoc
//
//	@Summary		Return list of peers which have sent a dht update
//	@Description	Gets a list of peers the libp2p node has received a dht update from
//	@Tags			p2p
//	@Produce		json
//	@Success		200	{string}	string
//	@Router			/peers/kad-dht [get]
func HandleListKadDHTPeers(c *gin.Context) {
	reqCtx := c.Request.Context()

	peers, err := libp2p.ListKadDHTPeers(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}
	if len(peers) == 0 {
		c.JSON(200, gin.H{"message": "no peers found"})
		klogger.Logger.Error("No peers found")
		return
	}
	resp, err := json.Marshal(peers)
	if err != nil {
		zlog.ErrorContext(reqCtx, "failed to json marshal DHT peers slice: %w", zap.Error(err))
	}
	klogger.Logger.Info("Response: " + string(resp))
	c.JSON(200, peers)
}

// HandleDefaultDepReqPeer  godoc
//
//	@Summary		Manage default deplyment request receiver peer
//	@Description	Set peer as the default receipient of deployment requests by setting the peerID parameter on GET request.
//	@Description	Show peer set as default deployment request receiver by sending a GET request without any parameters.
//	@Description	Remove default deployment request receiver by sending a GET request with peerID parameter set to '0'.
//	@Tags			peers
//	@Success		200
//	@Router			/peers/depreq [get]
func HandleDefaultDepReqPeer(c *gin.Context) {
	id := c.Query("peerID")
	reqCtx := c.Request.Context()

	msg, err := libp2p.DefaultDepReqPeer(reqCtx, id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": msg})
}

// HandleListFileRequests  godoc
// @Summary      List file transfer requests
// @Description  Get a list of file transfer requests from peers
// @Tags         file
// @Produce      json
// @Success      200
// @Router       /peers/file [get]
func HandleListFileRequests(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/file"))

	req, err := libp2p.IncomingFileTransferRequests()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, req)
}

// HandleInitiateFileTransfer  godoc
// @Summary      Send a file to a peer
// @Description  Initiate file transfer to a peer. filePath and peerID are required arguments.
// @Tags         file
// @Success      200
// @Router       /peers/file/send [get]
func HandleInitiateFileTransfer(c *gin.Context) {
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

	path := c.Query("filePath")
	if len(path) == 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": "filepath not provided"})
		return
	}

	err := libp2p.InitiateTransferFile(c, c.Writer, c.Request, id, path)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, nil)
}

// HandleAcceptFileTransfer  godoc
// @Summary      Accept incoming file transfer
// @Description  Accept an incoming file transfer. Incoming file transfer stream ID is a required parameter.
// @Tags         file
// @Success      200
// @Router       /peers/file/accept [get]
func HandleAcceptFileTransfer(c *gin.Context) {
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

	path, transferCh, err := libp2p.AcceptFileTransfer(c, libp2p.CurrentFileTransfer)
	if err != nil {
		zlog.Sugar().Errorf("error: could not accept file transfer - %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// TODO: Create function just for handling WS upgrade and progress of the transfer
	// so that both InitiateTransferFile and AcceptFileTransfer can reuse it

	// upgrade to websocket and stream transfer progress
	ws, err := internal.UpgradeConnection.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Sugar().Errorf("failed to set websocket upgrade: %w\n", err)
		return
	}

	ws.WriteJSON(gin.H{"message": "File Transfer Accepted."})
	for p := range transferCh {
		ws.WriteJSON(gin.H{
			"remaining_time": fmt.Sprintf("%v seconds", p.Remaining().Round(time.Second)),
			"percentage":     fmt.Sprintf("%.2f %%", p.Percent()),
			"size":           fmt.Sprintf("%.2f MB", p.N()/1048576),
		})
	}
	ws.WriteMessage(1, []byte("transfer complete"))
	ws.WriteMessage(1, []byte("File saved to: "+path))
	ws.Close()
}

// HandleJoinChat  godoc
//
//	@Summary		Join chat with a peer
//	@Description	Join a chat session started by a peer
//	@Tags			chat
//	@Success		200
//	@Router			/peers/chat/join [get]
func HandleJoinChat(c *gin.Context) {
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

// DEBUG
func HandleManualDHTUpdate(c *gin.Context) {
	go libp2p.UpdateKadDHT()
	libp2p.GetDHTUpdates(c)
}

// DEBUG
func HandleCleanupPeer(c *gin.Context) {
	id := c.Query("peerID")

	if id == "" {
		c.JSON(400, gin.H{"error": "peer ID not provided"})
		return
	}
	if id == libp2p.GetP2P().Host.ID().String() {
		c.JSON(400, gin.H{"error": "peerID can not be self peerID"})
		return
	}
	err := libp2p.CleanupPeer(id)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("unable to cleanup peer: %w", err)})
	}
	c.JSON(200, gin.H{"message": fmt.Sprintf("successfully cleaned up peer: %s", id)})
}

// DEBUG
func PingPeerHandler(c *gin.Context) {
	reqCtx := c.Request.Context()
	id := c.Query("peerID")

	if id == "" {
		c.JSON(400, gin.H{"error": "peerID not provided"})
		return
	}
	if id == libp2p.GetP2P().Host.ID().String() {
		c.JSON(400, gin.H{"error": "peerID can not be self peerID"})
		return
	}
	status, err := libp2p.PingPeer(reqCtx, id)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("could not ping peer: %w", err), "peer_in_dht": status})
		return
	}
	c.JSON(200, gin.H{"message": fmt.Sprintf("ping successful with peer %s", id), "peer_in_dht": status})
}

// DEBUG
func HandleDumpKademliaDHT(c *gin.Context) {
	reqCtx := c.Request.Context()
	dht, err := libp2p.DumpKademliaDHT(reqCtx)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if len(dht) == 0 {
		c.JSON(200, gin.H{"message": "empty DHT"})
	} else {
		c.JSON(200, dht)
	}
}
