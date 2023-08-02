package machines

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/peer"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/integrations/oracle"
	"gitlab.com/nunet/device-management-service/internal"
	"gitlab.com/nunet/device-management-service/internal/config"
	kLogger "gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
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

type fundingRespToSPD struct {
	ComputeProviderAddr string  `json:"compute_provider_addr"`
	EstimatedPrice      float64 `json:"estimated_price"`
	Signature           string  `json:"signature"`
	OracleMessage       string  `json:"oracle_message"`
}

var depreqWsConn *internal.WebSocketConnection

// HandleRequestService  godoc
//
//	@Summary		Informs parameters related to blockchain to request to run a service on NuNet
//	@Description	HandleRequestService searches the DHT for non-busy, available devices with appropriate metadata. Then informs parameters related to blockchain to request to run a service on NuNet.
//	@Tags			run
//	@Param			deployment_request	body		models.DeploymentRequest	true	"Deployment Request"
//	@Success		200					{object}	fundingRespToSPD
//	@Router			/run/request-service [post]
func HandleRequestService(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/run/request-service"))
	kLogger.Info("Handle request service", span)

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

	depReq.Timestamp = time.Now().In(time.UTC)
	depReq.Params.LocalNodeID = selfNodeID
	depReq.Params.LocalPublicKey = pKey.Type().String()

	// check if the pricing matched
	estimatedNtx := CalculateStaticNtxGpu(depReq)
	zlog.Sugar().Infof("estimated ntx price %v", estimatedNtx)
	if estimatedNtx > float64(depReq.MaxNtx) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "nunet estimation price is greater than client price"})
		return
	}

	var filteredPeers []models.PeerData

	// XXX setting target peer if specified in config
	if config.GetConfig().Job.TargetPeer != "" {
		zlog.Debug("Going for target peer specified in config")
		machines := libp2p.FetchMachines(libp2p.GetP2P().Host)
		filteredPeers = libp2p.PeersWithMatchingSpec([]models.PeerData{machines[config.GetConfig().Job.TargetPeer]}, depReq)
	} else {
		zlog.Debug("Filtering peers - no default target peer specified")
		filteredPeers = FilterPeers(depReq, libp2p.GetP2P().Host)
	}

	if len(filteredPeers) < 1 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "no peers found with matched specs"})
		return
	}

	var onlinePeer models.PeerData
	var rtt time.Duration = 1000000000000000000
	for _, node := range filteredPeers {
		targetPeer, err := peer.Decode(node.PeerID)
		if err != nil {
			zlog.Sugar().Errorf("Error decoding peer ID: %v", err)
			return
		}
		pingResult, pingCancel := libp2p.Ping(c.Request.Context(), targetPeer)
		result := <-pingResult
		if result.Error == nil {
			if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
				zlog.Sugar().Info("Peer is online.", "RTT", result.RTT, "PeerID", node.PeerID)
			}
			if result.RTT < rtt {
				rtt = result.RTT
				onlinePeer = node
			}
		} else {
			if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
				zlog.Sugar().Infof("Peer - %s is offline. Error: %v", node.PeerID, result.Error)
			}
		}
		pingCancel()
	}

	if onlinePeer.PeerID == "" {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "no peers found with matched specs"})
		return
	}
	computeProvider := onlinePeer

	zlog.Sugar().Debugf("compute provider: %v", computeProvider)

	depReq.Params.RemoteNodeID = computeProvider.PeerID
	computeProviderPeerID, err := peer.Decode(computeProvider.PeerID)
	if err != nil {
		zlog.Sugar().Errorf("Error decoding peer ID: %v", err)
		return
	}
	computeProviderPubKey := libp2p.GetP2P().Host.Peerstore().PubKey(computeProviderPeerID)

	depReq.Params.RemotePublicKey = computeProviderPubKey.Type().String()
	zlog.Sugar().Debugf("compute provider public key: %s", computeProviderPubKey.Type().String())

	// oracle inputs: service provider user address, max tokens amount, type of blockchain (cardano or ethereum)
	zlog.Info("sending fund contract request to oracle")
	fcr, err := oracle.FundContractRequest()
	if err != nil {
		zlog.Info("sending fund contract request to oracle failed")
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
		zlog.Info("cannot write to database")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "cannot write to database"})
		return
	}

	// oracle outputs: compute provider user address, estimated price, signature, oracle message
	resp := fundingRespToSPD{
		ComputeProviderAddr: computeProvider.TokenomicsAddress,
		EstimatedPrice:      estimatedNtx,
		Signature:           fcr.Signature,
		OracleMessage:       fcr.OracleMessage,
	}
	c.JSON(200, resp)
	go outgoingDepReqWebsock()
}

// HandleDeploymentRequest  godoc
//
//	@Summary		Websocket endpoint responsible for sending deployment request and receiving deployment response.
//	@Description	Loads deployment request from the DB after a successful blockchain transaction has been made and passes it to compute provider.
//	@Tags			run
//	@Success		200	{string}	string
//	@Router			/run/deploy [get]
func HandleDeploymentRequest(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/run/deploy"))
	kLogger.Info("Handle deployment request", span)

	ws, err := internal.UpgradeConnection.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Sugar().Errorf("Failed to set websocket upgrade: %v", err)
		return
	}

	err = ws.WriteMessage(websocket.TextMessage, []byte("{\"action\": \"connected\", \"message\": \"You are connected to DMS for docker (GPU/CPU) deployment.\"}"))
	if err != nil {
		zlog.Error(err.Error())
	}

	depreqWsConn = &internal.WebSocketConnection{Conn: ws}

	go incomingDepReqWebsock(c)

}

// incomingDepReqWebsock listens for messages from websocket client
func incomingDepReqWebsock(ctx *gin.Context) {
	zlog.Info("Started listening for messages from websocket client")
	listen := true
	defer func() {
		if r := recover(); r != nil {
			zlog.Sugar().Warnf("closing sock after panic - %v", r)
			if depreqWsConn != nil {
				depreqWsConn.Close()
			}
			listen = false
		}
	}()

	for listen {
		zlog.Info("Listening for messages from websocket client")
		select {
		case <-ctx.Done():
			zlog.Info("Context done - stopping listen")
			if depreqWsConn != nil {
				depreqWsConn.Close()
				depreqWsConn = nil
			}
			listen = false
		default:
			_, p, err := depreqWsConn.ReadMessage()
			if err != nil {
				zlog.Sugar().Warnf("unable to read from websocket - stopping listen: %v", err)
				listen = false
			}
			zlog.Sugar().Debugf("Received message from websocket client: %s", string(p))
			handleWebsocketAction(ctx, p, depreqWsConn)
		}
	}
	zlog.Info("Exiting incomingDepReqWebsock")
}

// outgoingDepReqWebsock is for sending messages to websocket client.
// it listens on several queues that receive updates from CP DMS
func outgoingDepReqWebsock() {
	zlog.Info("Started listening for incoming messages from depreq stream to pass to websocket client")
	listen := true
	for listen {
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

				var wsResponse = wsMessage{
					Action:  "deployment-response",
					Message: json.RawMessage(msg),
				}

				resp, _ := json.Marshal(wsResponse)
				zlog.Sugar().Debugf("[websocket] deployment response to websocket: %s", string(resp))
				err := depreqWsConn.WriteMessage(websocket.TextMessage, resp)
				if err != nil {
					zlog.Sugar().Errorf("unable to write to websocket: %v", err)
				}

			} else {
				zlog.Info("Channel closed!")
				listen = false
			}
		case msg, ok := <-libp2p.JobLogStdoutQueue:
			if ok {
				stdoutLog := struct {
					Action string `json:"action"`
					Stdout string `json:"stdout"`
				}{
					Action: "log-stream-response",
					Stdout: msg,
				}

				resp, _ := json.Marshal(stdoutLog)
				zlog.Debug("[websocket] stdout update to websocket")
				err := depreqWsConn.WriteMessage(websocket.TextMessage, resp)
				if err != nil {
					zlog.Sugar().Errorf("unable to write to websocket: %v", err)
				}
			} else {
				zlog.Info("Channel closed!")
				listen = false
			}
		case msg, ok := <-libp2p.JobLogStderrQueue:
			if ok {
				stderrLog := struct {
					Action string `json:"action"`
					Stderr string `json:"stderr"`
				}{
					Action: "log-stream-response",
					Stderr: msg,
				}

				resp, _ := json.Marshal(stderrLog)
				zlog.Debug("[websocket] stderr update to websocket")
				err := depreqWsConn.WriteMessage(websocket.TextMessage, resp)
				if err != nil {
					zlog.Sugar().Errorf("unable to write to websocket: %v", err)
				}
			} else {
				zlog.Info("Channel closed!")
				listen = false
			}
		case _, ok := <-libp2p.JobCompletedQueue:
			if ok {
				wsResponse := wsMessage{
					Action: "job-completed",
				}
				resp, _ := json.Marshal(wsResponse)
				zlog.Debug("[websocket] job-completed to websocket")
				err := depreqWsConn.WriteMessage(websocket.TextMessage, resp)
				if err != nil {
					zlog.Sugar().Errorf("unable to write to websocket: %v", err)
				}

			} else {
				zlog.Info("Channel closed!")
				listen = false
			}
			zlog.Sugar().Infof("Job completed. - ok: %v - Returning", ok)
			listen = false
		case msg, ok := <-libp2p.JobFailedQueue:
			if ok {
				zlog.Debug("[websocket] job-failed to websocket")
				wsResponse := struct {
					Action  string `json:"action"`
					Message string `json:"message"`
				}{
					Action:  "job-failed",
					Message: msg,
				}
				resp, err := json.Marshal(wsResponse)
				if err != nil {
					zlog.Error("error in job-failed websocket response")
				}
				err = depreqWsConn.WriteMessage(websocket.TextMessage, resp)
				if err != nil {
					zlog.Sugar().Errorf("unable to write to websocket: %v", err)
				}
			} else {
				zlog.Info("Channel closed!")
				listen = false
			}
			zlog.Sugar().Infof("Job failed. - ok: %v - Returning", ok)
			listen = false
		}
	}
	zlog.Info("Exiting outgoingDepReqWebsock")
}

func handleWebsocketAction(ctx *gin.Context, payload []byte, conn *internal.WebSocketConnection) {
	// convert json to go values
	var m wsMessage
	err := json.Unmarshal(payload, &m)
	if err != nil {
		zlog.Sugar().Errorf("wrong message payload: %v", err)
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
	kLogger.Info("send deployment request", span)

	defer span.End()

	// load depReq from the database
	var depReqFlat models.DeploymentRequestFlat
	var depReq models.DeploymentRequest

	// SELECTs the first record; first record which is not marked as delete
	if err := db.DB.First(&depReqFlat).Error; err != nil {
		zlog.Error(err.Error())
	}

	err := json.Unmarshal([]byte(depReqFlat.DeploymentRequest), &depReq)
	if err != nil {
		zlog.Sugar().Errorf(err.Error())
	}

	//extract a span context parameters , so that a child span constructed on the reciever side
	depReq.TraceInfo.SpanID = span.SpanContext().TraceID().String()
	depReq.TraceInfo.TraceID = span.SpanContext().SpanID().String()
	depReq.TraceInfo.TraceFlags = span.SpanContext().TraceFlags().String()
	depReq.TraceInfo.TraceStates = span.SpanContext().TraceState().String()

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
//
//	@Summary		Sends blockchain status of contract creation.
//	@Description	HandleSendStatus is used by webapps to send status of blockchain activities. Such as if tokens have been put in escrow account and account creation.
//	@Tags			run
//	@Param			body	body		BlockchainTxStatus	true	"Blockchain Transaction Status Body"
//	@Success		200		{string}	string
//	@Router			/run/send-status [post]
func HandleSendStatus(c *gin.Context) {
	// TODO: This is a stub function. Replace the logic to talk with Oracle.
	rand.Seed(time.Now().Unix())

	body := BlockchainTxStatus{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cannot read payload body"})
		return
	}

	var requestTracker models.RequestTracker
	res := db.DB.Where("id = ?", 1).Find(&requestTracker)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
	}
	if body.TransactionType == "withdraw" && body.TransactionStatus == "success" {
		// Delete the entry
		if err := db.DB.Where("deleted_at IS NULL").Delete(&models.Services{}).Error; err != nil {
			zlog.Sugar().Errorln(err)
		}
	}

	serviceStatus := body.TransactionType + " with " + body.TransactionStatus

	requestTracker.Status = serviceStatus
	db.DB.Save(&requestTracker)

	c.JSON(200, gin.H{"message": fmt.Sprintf("transaction status %s acknowledged", body.TransactionStatus)})
}
