package machines

import (
	"context"
	"encoding/json"
	"errors"
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
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// add node_id and public_key in deployment request
	pKey, err := libp2p.GetPublicKey()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("unable to obtain public key: %v", err))
		return
	}
	selfNodeID := libp2p.GetP2P().Host.ID().String()

	depReq.Params.NodeID = selfNodeID
	depReq.Params.PublicKey = pKey.Type().String()

	// check if the pricing matched
	estimatedNtx := CalculateStaticNtxGpu(depReq)
	zlog.Sugar().Infof("estimated ntx price %v", estimatedNtx)
	if estimatedNtx > float64(depReq.MaxNtx) {
		c.AbortWithError(http.StatusBadRequest, errors.New("nunet estimation price is greater than client price"))
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
		c.AbortWithError(http.StatusBadRequest, errors.New("no peers found with matched specs"))
		return
	}
	computeProvider := onlinePeer
	if len(filteredPeers) < 1 {
		c.AbortWithError(http.StatusBadRequest, errors.New("no peers found with matched specs"))
		return
	}
	zlog.Sugar().Debugf("compute provider: ", computeProvider)

	depReq.Params.NodeID = computeProvider.PeerID

	// oracle inputs: service provider user address, max tokens amount, type of blockchain (cardano or ethereum)
	zlog.Sugar().Infof("sending fund contract request to oracle")
	fcr, err := oracle.FundContractRequest()
	if err != nil {
		zlog.Sugar().Infof("sending fund contract request to oracle failed")
		c.AbortWithError(http.StatusBadRequest, errors.New("cannot connect to oracle"))
		return
	}

	// Marshal struct to JSON
	depReqBytes, err := json.Marshal(depReq)
	if err != nil {
		zlog.Sugar().Errorf("failed to marshal struct to json: %v", err)
		return
	}

	// Convert JSON bytes to string
	depReqStr := string(depReqBytes)
	depReqFlat.DeploymentRequest = depReqStr

	if err := db.DB.Create(&depReqFlat).Error; err != nil {
		panic(err)
	}

	var requestTracker models.RequestTracker
	res := db.DB.Where("id = ?", 1).Find(&requestTracker)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
	}

	// sending ntx_payment info to stats database via grpc Call
	NtxPaymentParams := models.NtxPayment{
		CallID:      requestTracker.CallID,
		ServiceID:   requestTracker.ServiceType,
		AmountOfNtx: int32(estimatedNtx),
		PeerID:      requestTracker.NodeID,
		Timestamp:   float32(statsdb.GetTimestamp()),
	}
	statsdb.NtxPayment(NtxPaymentParams)

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

	go listenForDeploymentStatus(c, &conn)
	go listenForDeploymentResponse(c, &conn)
}

func listenForDeploymentStatus(ctx *gin.Context, conn *internal.WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			zlog.Error(fmt.Sprintf("%v", r))
		}
	}()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			zlog.Error(err.Error())
			conn.Close()
			return
		}

		handleWebsocketAction(ctx, p)
	}
}

func listenForDeploymentResponse(ctx *gin.Context, conn *internal.WebSocketConnection) {
	for {
		// 1. check if DepResQueue has anything
		select {
		case msg, ok := <-libp2p.DepResQueue:
			if ok {
				zlog.Info("Deployment response received. Sending it to connected websocket client")
				var depRes models.DeploymentResponse

				jsonDataMsg, _ := json.Marshal(msg)
				json.Unmarshal(jsonDataMsg, &depRes)
				msg, _ := json.Marshal(depRes)

				// 2. Send the content to the client connected
				wsResponse := wsMessage{
					Action:  "deployment-response",
					Message: json.RawMessage(msg),
				}

				msg, _ = json.Marshal(wsResponse)

				zlog.Sugar().Debugf("deployment response to websock: %s", string(msg))

				conn.WriteMessage(websocket.TextMessage, msg)
			} else {
				zlog.Info("Channel closed!")
			}
		}
	}
}

func handleWebsocketAction(ctx *gin.Context, payload []byte) {
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

		err = sendDeploymentRequest(ctx)
		if err != nil {
			zlog.Error(err.Error())
		}
	}
}

func sendDeploymentRequest(ctx *gin.Context) error {
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

	// delete temporary record
	// XXX: Needs to be modified to take multiple deployment requests from same service provider
	// deletes all the record in table; deletes == mark as delete
	if err := db.DB.Where("deleted_at IS NULL").Delete(&models.DeploymentRequestFlat{}).Error; err != nil {
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

	depReqStream, err := libp2p.SendDeploymentRequest(ctx, depReq)
	if err != nil {
		return err
	}

	//TODO: Context handling and cancellation on all messaging related code
	//      most importantly, depreq/depres messaging
	//XXX: needs a lot of testing.
	go libp2p.DeploymentResponseListener(depReqStream)

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
		c.AbortWithError(http.StatusBadRequest, err)
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
