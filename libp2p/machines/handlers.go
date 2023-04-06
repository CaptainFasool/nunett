package machines

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/integrations/oracle"
	"gitlab.com/nunet/device-management-service/internal"
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
	zlog.Sugar().Info("estimated ntx price %v", estimatedNtx)
	if estimatedNtx > float64(depReq.MaxNtx) {
		c.AbortWithError(http.StatusBadRequest, errors.New("nunet estimation price is greater than client price"))
		return
	}

	filteredPeers := FilterPeers(depReq, libp2p.GetP2P().Host)
	if len(filteredPeers) < 1 {
		c.AbortWithError(http.StatusBadRequest, errors.New("no peers found with matched specs"))
		return
	}
	computeProvider := filteredPeers[0]
	if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
		zlog.Sugar().Debugf("compute provider: %v", computeProvider)
	}

	depReq.Params.NodeID = computeProvider.PeerID

	// oracle inputs: service provider user address, max tokens amount, type of blockchain (cardano or ethereum)
	zlog.Sugar().Info("sending fund contract request to oracle")
	fcr, err := oracle.FundContractRequest()
	if err != nil {
		zlog.Sugar().Info("sending fund contract request to oracle failed")
		c.AbortWithError(http.StatusBadRequest, errors.New("cannot connect to oracle"))
		return
	}

	// Marshal struct to JSON
	depReqBytes, err := json.Marshal(depReq)
	if err != nil {
		zlog.Sugar().Errorln("marshaling struct to json: %v", err)
		return
	}

	// Convert JSON bytes to string
	depReqStr := string(depReqBytes)
	depReqFlat.DeploymentRequest = depReqStr

	result := db.DB.Create(&depReqFlat)
	if result.Error != nil {
		panic(result.Error)
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
	result := db.DB.First(&depReqFlat) // SELECTs the first record; first record which is not marked as delete
	if result.Error != nil {
		zlog.Sugar().Errorf("%v", result.Error)
	}

	// delete temporary record
	// XXX: Needs to be modified to take multiple deployment requests from same service provider
	result = db.DB.Where("deleted_at IS NULL").Delete(&models.DeploymentRequestFlat{}) // deletes all the record in table; deletes == mark as delete
	if result.Error != nil {
		zlog.Sugar().Errorf("%v", result.Error)
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
	if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
		zlog.Sugar().Debugf("deployment request: %v", depReq)
	}

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

	status := []string{"success", "error"}
	randomStatus := status[rand.Intn(len(status))]

	c.JSON(200, gin.H{"message": randomStatus})
}
