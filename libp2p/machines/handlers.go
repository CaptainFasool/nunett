package machines

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/peer"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/integrations/oracle"
	"gitlab.com/nunet/device-management-service/internal"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/statsdb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type wsMessage struct {
	Action  string          `json:"action"`
	Message json.RawMessage `json:"message"`
}

type BlockchainTxStatus struct {
	TransactionType   string `json:"transaction_type"`
	TransactionStatus string `json:"transaction_status"`
}

// HandleRequestService  godoc
// @Summary      Informs parameters related to blockchain to request to run a service on NuNet
// @Description  HandleRequestService searches the DHT for non-busy, available devices with appropriate metadata. Then informs parameters related to blockchain to request to run a service on NuNet.
// @Success      200  {string}  string
// @Router       /run/request-service [post]
func HandleRequestService(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/run/request-service"))

	// receive deployment request
	var depReq models.DeploymentRequest
	var depReqFlat models.DeploymentRequestFlat
	if err := c.BindJSON(&depReq); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "json: cannot unmarshal object into Go"})
		return
	}

	// Check if there is already a running job
	if libp2p.IsDepReqStreamOpen() {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "a service is already running; only 1 service is supported at the moment"})
		return
	}

	// since IsDepReqStreamOpen() is false, we can assume that there is no running job because we've lost connection to the peer
	if result := db.DB.Where("deleted_at IS NULL").Find(&depReqFlat).RowsAffected; result > 0 {
		zlog.Sugar().Infof(
			"Deployed job unknown status (outbound depReq not open anymore). Deleting DepReqFlat record (i=%d, n=%d) from DB",
			depReqFlat.ID, result)
		// XXX: Needs to be modified to take multiple deployment requests from same service provider
		// deletes all the record in table; deletes == mark as delete
		if err := db.DB.Where("deleted_at IS NULL").Delete(&depReqFlat).Error; err != nil {
			// TODO: Do not delete, update JobStatus
			zlog.Sugar().Errorf("%v", err)
		}
	}

	// add node_id and public_key in deployment request
	pKey, err := libp2p.GetPublicKey()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unable to obtain public key"})
		return
	}
	selfNodeID := libp2p.GetP2P().Host.ID().String()

	depReq.Params.LocalNodeID = selfNodeID
	depReq.Params.LocalPublicKey = pKey.Type().String()

	// check if the pricing matched
	estimatedNtx := CalculateStaticNtxGpu(depReq)
	zlog.Sugar().Infof("estimated ntx price %v", estimatedNtx)
	if estimatedNtx > float64(depReq.MaxNtx) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "nunet estimation price is greater than client price"})
		return
	}

	filteredPeers := FilterPeers(depReq, libp2p.GetP2P().Host)
	var onlinePeer models.PeerData
	var rtt time.Duration = 1000000000000000000
	ctx := context.Background()
	defer ctx.Done()
	for _, node := range filteredPeers {
		targetPeer, err := peer.Decode(node.PeerID)
		if err != nil {
			zlog.Sugar().Errorf("Error decoding peer ID: %v\n", err)
			return
		}
		res := libp2p.PingPeer(ctx, libp2p.GetP2P().Host, targetPeer)
		if res.Success {
			if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
				zlog.Sugar().Info("Peer is online.", "RTT", res.RTT, "PeerID", node.PeerID)
			}
			if res.RTT < rtt {
				rtt = res.RTT
				onlinePeer = node
			}
		} else {
			if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
				zlog.Sugar().Info("Peer -  ", node.PeerID, " is offline.")
			}
		}
	}
	if onlinePeer.PeerID == "" {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "no peers found with matched specs"})
		return
	}
	computeProvider := onlinePeer
	if len(filteredPeers) < 1 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "no peers found with matched specs"})
		return
	}
	zlog.Sugar().Debugf("compute provider: ", computeProvider)

	depReq.Params.RemoteNodeID = computeProvider.PeerID
	computeProviderPeerID, err := peer.Decode(computeProvider.PeerID)
	if err != nil {
		zlog.Sugar().Errorf("Error decoding peer ID: %v\n", err)
		return
	}
	computeProviderPubKey, err := computeProviderPeerID.ExtractPublicKey()
	if err != nil {
		zlog.Sugar().Errorf("unable to extract public key from peer id: %v", err)
		depReq.Params.RemotePublicKey = ""
	} else {
		depReq.Params.RemotePublicKey = computeProviderPubKey.Type().String()
		zlog.Sugar().Debugf("compute provider public key: ", computeProviderPubKey.Type().String())
	}
	// oracle inputs: service provider user address, max tokens amount, type of blockchain (cardano or ethereum)
	zlog.Sugar().Infof("sending fund contract request to oracle")
	fcr, err := oracle.FundContractRequest()
	if err != nil {
		zlog.Sugar().Infof("sending fund contract request to oracle failed")
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "cannot connect to oracle"})
		return
	}

	// Marshal struct to JSON
	depReqBytes, _ := json.Marshal(depReq)
	// Convert JSON bytes to string
	depReqStr := string(depReqBytes)

	depReqFlat = models.DeploymentRequestFlat{} // reset
	depReqFlat.DeploymentRequest = depReqStr
	depReqFlat.JobStatus = "awaiting"

	if err := db.DB.Create(&depReqFlat).Error; err != nil {
		zlog.Sugar().Infof("cannot write to database")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "cannot write to database"})
		return
	}

	// oracle outputs: compute provider user address, estimated price, signature, oracle message
	fundingRespToWebapp := struct {
		ComputeProviderAddr string  `json:"compute_provider_addr"`
		EstimatedPrice      float64 `json:"estimated_price"`
		Signature           string  `json:"signature"`
		OracleMessage       string  `json:"oracle_message"`
	}{
		ComputeProviderAddr: computeProvider.TokenomicsAddress,
		EstimatedPrice:      estimatedNtx,
		Signature:           fcr.Signature,
		OracleMessage:       fcr.OracleMessage,
	}
	c.JSON(200, fundingRespToWebapp)
}

// HandleDeploymentRequest  godoc
// @Summary      Websocket endpoint responsible for sending deployment request and receiving deployment response.
// @Description  Loads deployment request from the DB after a successful blockchain transaction has been made and passes it to compute provider.
// @Success      200  {string}  string
// @Router       /run/deploy [get]
func HandleDeploymentRequest(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/run/deploy"))

	ws, err := internal.UpgradeConnection.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Error(fmt.Sprintf("Failed to set websocket upgrade: %+v\n", err))
		return
	}

	err = ws.WriteMessage(websocket.TextMessage, []byte("{\"action\": \"connected\", \"message\": \"You are connected to DMS for docker (GPU/CPU) deployment.\"}"))
	if err != nil {
		zlog.Error(err.Error())
	}

	conn := internal.WebSocketConnection{Conn: ws}

	go listenToOutgoingMessage(c, &conn)
	go listenForIncomingMessage(c, &conn)
}

// listenToOutgoingMessage is for sending message to websocket server.
// If you want to add a new message to receive, see listenForDeploymentResponse.
func listenToOutgoingMessage(ctx *gin.Context, conn *internal.WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			zlog.Sugar().Warnf("closing sock after panic - %v", r)
			if conn != nil {
				conn.Close()
			}
		}
	}()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			zlog.Sugar().Warnf("unable to read from websocket  - panicking: %v", err.Error())
			panic(err)
		}
		zlog.Sugar().Debugf("Received message from websocket client: %v", string(p))
		handleWebsocketAction(ctx, p, conn)
	}
}

// listenForDeploymentStatus is for receiving message to websocket client.
// If you want to add a new message to send to server, see listenForDeploymentStatus.
func listenForIncomingMessage(ctx *gin.Context, conn *internal.WebSocketConnection) {
	for {
		// 1. check if DepResQueue has anything
		select {
		case rawMsg, ok := <-libp2p.DepResQueue:
			zlog.Sugar().Infof("Deployment response received. - msg: %v , ok: %v", rawMsg, ok)
			if ok {
				zlog.Info("Sending Received DepResp to connected websocket client")
				var depRes models.DeploymentResponse

				jsonDataMsg, _ := json.Marshal(rawMsg)
				json.Unmarshal(jsonDataMsg, &depRes)
				msg, _ := json.Marshal(depRes)

				var wsResponse wsMessage

				if strings.Contains(depRes.Content, libp2p.ContainerJobFinishedWithErrors) {
					zlog.Info("Finished: Sending Deployment failed to websock")
					err := conn.WriteJSON(map[string]string{"action": libp2p.JobFailed})
					if err != nil {
						zlog.Sugar().Errorf("unable to write to websocket: %v", err)
						break
					}
					zlog.Info("Closing websocket connection")
					conn.Close()
				} else if strings.Contains(depRes.Content, libp2p.ContainerJobFinishedWithoutErrors) {
					zlog.Info("Finished: Sending Deployment success to websock")
					err := conn.WriteJSON(map[string]string{"action": libp2p.JobCompleted})
					if err != nil {
						zlog.Sugar().Errorf("unable to write to websocket: %v", err)
						break
					}
					zlog.Info("Closing websocket connection")
					conn.Close()
				} else {
					zlog.Info("Sending Deployment status/response to websock")
					wsResponse = wsMessage{
						Action:  "deployment-response",
						Message: json.RawMessage(msg),
					}
					err := conn.WriteJSON(wsResponse)
					if err != nil {
						zlog.Sugar().Errorf("unable to write to websocket: %v", err)
						break
					}
				}
			} else {
				zlog.Info("Channel closed!")
			}
		}
	}
}

func handleWebsocketAction(ctx *gin.Context, payload []byte, conn *internal.WebSocketConnection) {
	// convert json to go values
	var m wsMessage
	err := json.Unmarshal(payload, &m)
	if err != nil {
		zlog.Error(fmt.Sprintf("wrong message payload: %v", err))
	}

	switch m.Action {
	case "send-status":
		var txStatus BlockchainTxStatus
		err := json.Unmarshal(m.Message, &txStatus)
		if err != nil || txStatus.TransactionStatus != "success" {
			// Send event to Signoz
			span := trace.SpanFromContext(ctx.Request.Context())
			span.SetAttributes(attribute.String("URL", "/run/deploy"))
			span.SetAttributes(attribute.String("TransactionStatus", "error"))
			return
		}

		err = sendDeploymentRequest(ctx, conn)
		if err != nil {
			zlog.Error(err.Error())
		}
	}
}

func sendDeploymentRequest(ctx *gin.Context, conn *internal.WebSocketConnection) error {
	span := trace.SpanFromContext(ctx.Request.Context())
	span.SetAttributes(attribute.String("URL", "/run/deploy"))
	span.SetAttributes(attribute.String("TransactionStatus", "success"))
	defer span.End()

	// load depReq from the database
	var depReqFlat models.DeploymentRequestFlat
	var depReq models.DeploymentRequest

	// SELECTs the first record; first record which is not marked as delete
	if err := db.DB.First(&depReqFlat).Error; err != nil {
		zlog.Sugar().Errorf("%v", err)
	}

	err := json.Unmarshal([]byte(depReqFlat.DeploymentRequest), &depReq)
	if err != nil {
		zlog.Sugar().Errorf("%v", err)
	}

	//extract a span context parameters , so that a child span constructed on the reciever side
	depReq.TraceInfo.SpanID = span.SpanContext().TraceID().String()
	depReq.TraceInfo.TraceID = span.SpanContext().SpanID().String()
	depReq.TraceInfo.TraceFlags = span.SpanContext().TraceFlags().String()
	depReq.TraceInfo.TraceStates = span.SpanContext().TraceState().String()

	depReq.Timestamp = time.Now()

	zlog.Sugar().Debugf("deployment request: %v", depReq)

	// notify websocket client about the
	submitted := map[string]string{"action": libp2p.JobSubmitted}
	err = conn.WriteJSON(submitted)
	if err != nil {
		zlog.Sugar().Errorf("unable to write to websocket: %v", err)
	}

	depReqStream, err := libp2p.SendDeploymentRequest(ctx, depReq)
	if err != nil {
		return err
	}

	//TODO: Context handling and cancellation on all messaging related code
	//      most importantly, depreq/depres messaging
	//XXX: needs a lot of testing.
	go libp2p.DeploymentUpdateListener(depReqStream)

	return nil
}

// HandleSendStatus  godoc
// @Summary      Sends blockchain status of contract creation.
// @Description  HandleSendStatus is used by webapps to send status of blockchain activities. Such as if tokens have been put in escrow account and account creation.
// @Success      200  {string}  string
// @Router       /run/send-status [post]
func HandleSendStatus(c *gin.Context) {
	// TODO: This is a stub function. Replace the logic to talk with Oracle.
	rand.Seed(time.Now().Unix())

	body := BlockchainTxStatus{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cannot read payload body"})
		return
	}

	if body.TransactionType == "withdraw" && body.TransactionStatus == "success" {
		// Delete the entry
		if err := db.DB.Where("deleted_at IS NULL").Delete(&models.Services{}).Error; err != nil {
			zlog.Sugar().Errorln(err)
		}
	}

	serviceStatus := body.TransactionType + " with " + body.TransactionStatus

	var requestTracker models.RequestTracker
	res := db.DB.Where("id = ?", 1).Find(&requestTracker)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
	}

	ServiceStatusParams := models.ServiceStatus{
		CallID:              requestTracker.CallID,
		PeerIDOfServiceHost: requestTracker.NodeID,
		Status:              serviceStatus,
		Timestamp:           float32(statsdb.GetTimestamp()),
	}
	statsdb.ServiceStatus(ServiceStatusParams)

	requestTracker.Status = serviceStatus
	db.DB.Save(&requestTracker)

	c.JSON(200, gin.H{"message": fmt.Sprintf("transaction status %s acknowledged", body.TransactionStatus)})
}
