package machines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/peer"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/dms/config"
	"gitlab.com/nunet/device-management-service/integrations/oracle"
	"gitlab.com/nunet/device-management-service/internal"
	kLogger "gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type wsMessage struct {
	Action  string          `json:"action"`
	Message json.RawMessage `json:"message"`
}

type fundingRespToSPD struct {
	ComputeProviderAddr string  `json:"compute_provider_addr"`
	EstimatedPrice      float64 `json:"estimated_price"`
	MetadataHash        string  `json:"metadata_hash"`
	WithdrawHash        string  `json:"withdraw_hash"`
	RefundHash          string  `json:"refund_hash"`
	Distribute_50Hash   string  `json:"distribute_50_hash"`
	Distribute_75Hash   string  `json:"distribute_75_hash"`
}

var depreqWsConn *internal.WebSocketConnection

func RequestService(ctx context.Context, depReq models.DeploymentRequest) (*fundingRespToSPD, error) {
	_, debug := os.LookupEnv("NUNET_DEBUG_VERBOSE")

	// Check if there is already a running job
	if libp2p.IsDepReqStreamOpen() {
		return nil, fmt.Errorf("a service is already running: only 1 service is supported at the moment")
	}

	// since IsDepReqStreamOpen() is false, we can assume that there is no running job because we've lost connection to the peer
	var depReqFlat models.DeploymentRequestFlat
	result := db.DB.Where("deleted_at IS NULL").Find(&depReqFlat).RowsAffected
	if result > 0 {
		zlog.Sugar().Infof(
			"Deployed job unknown status (outbound depReq not open anymore). Deleting DepReqFlat record (i=%d, n=%d) from DB",
			depReqFlat.ID, result)
		// XXX: Needs to be modified to take multiple deployment requests from same service provider
		// deletes all the record in table; deletes == mark as delete
		err := db.DB.Where("deleted_at IS NULL").Delete(&depReqFlat).Error
		if err != nil {
			// TODO: Do not delete, update JobStatus
			zlog.Sugar().Errorf("%v", err)
		}
	}

	// add node_id and public_key in deployment request
	pubKey, err := libp2p.GetPublicKey()
	if err != nil {
		return nil, fmt.Errorf("unable to obtain public key: %w", err)
	}
	selfID := libp2p.GetP2P().Host.ID().String()

	depReq.Timestamp = time.Now().In(time.UTC)
	depReq.Params.LocalNodeID = selfID
	localPubKey, _ := pubKey.Raw()
	depReq.Params.LocalPublicKey = string(localPubKey)

	// check if the pricing matched
	estimatedNTX := CalculateStaticNtxGpu(depReq)

	zlog.Sugar().Infof("estimated ntx price %v", estimatedNTX)
	if estimatedNTX > float64(depReq.MaxNtx) {
		return nil, fmt.Errorf("NuNet estimation price is greater than client price")
	}

	var filteredPeers []models.PeerData

	// XXX setting target peer if specified in config
	if config.GetConfig().Job.TargetPeer != "" {
		zlog.Debug("Going for target peer specified in config")
		machines := libp2p.FetchMachines(libp2p.GetP2P().Host)
		filteredPeers = libp2p.PeersWithMatchingSpec([]models.PeerData{machines[config.GetConfig().Job.TargetPeer]}, depReq)
	} else if depReq.Params.RemoteNodeID != "" {
		zlog.Sugar().Debugf("Going for target peer specified in deployment request: %s", depReq.Params.RemoteNodeID)
		machines := libp2p.FetchMachines(libp2p.GetP2P().Host)

		selectedPeerInfo, ok := machines[depReq.Params.RemoteNodeID]
		if ok {
			filteredPeers = libp2p.PeersWithMatchingSpec([]models.PeerData{selectedPeerInfo}, depReq)
		} else {
			return nil, fmt.Errorf("targeted peer is not within host DHT")
		}
	} else {
		zlog.Debug("Filtering peers - no default target peer specified")
		filteredPeers = FilterPeers(depReq, libp2p.GetP2P().Host)
	}

	zlog.Sugar().Infof("FILTERED PEERS: %+v", filteredPeers)

	if len(filteredPeers) < 1 {
		return nil, fmt.Errorf("no peers found with matched specs")
	}

	var onlinePeer models.PeerData
	var rtt time.Duration = 1000000000000000000
	for _, node := range filteredPeers {
		// check if tokenomics address is valid, if not, skip
		err = utils.ValidateAddress(node.TokenomicsAddress)
		if err != nil {
			zlog.Sugar().Errorf("invalid tokenomics address: %v", err)
			zlog.Sugar().Error("skipping peer due to invalid tokenomics address")
			continue
		}

		targetPeer, err := peer.Decode(node.PeerID)
		if err != nil {
			zlog.Sugar().Errorf("Error decoding peer ID: %v", err)
			return nil, fmt.Errorf("could not decode peer ID: %w", err)
		}
		// CONTEXT: Request
		pingResult, pingCancel := libp2p.Ping(ctx, targetPeer)
		result := <-pingResult
		if result.Error == nil {
			if debug {
				zlog.Sugar().Info("Peer is online.", "RTT", result.RTT, "PeerID", node.PeerID)
			}
			if result.RTT < rtt {
				rtt = result.RTT
				onlinePeer = node
			}
		} else {
			if debug {
				zlog.Sugar().Infof("Peer - %s is offline. Error: %v", node.PeerID, result.Error)
			}
		}
		pingCancel()
	}

	if onlinePeer.PeerID == "" {
		return nil, fmt.Errorf("no peers found with matched specs")
	}
	computeProvider := onlinePeer

	zlog.Sugar().Debugf("compute provider: %v", computeProvider)

	depReq.Params.RemoteNodeID = computeProvider.PeerID
	computeProviderPeerID, err := peer.Decode(computeProvider.PeerID)
	if err != nil {
		zlog.Sugar().Errorf("Error decoding peer ID: %v", err)
		return nil, fmt.Errorf("could not decode compute provider ID: %w", err)
	}
	computeProviderPubKey, err := libp2p.GetP2P().Host.Peerstore().PubKey(computeProviderPeerID).Raw()
	if err != nil {
		zlog.Sugar().Errorf("unable to obtain compute provider public key: %v", err)
	}

	depReq.Params.RemotePublicKey = string(computeProviderPubKey)

	// oracle inputs: service provider user address, max tokens amount, type of blockchain (cardano or ethereum)
	zlog.Info("sending fund contract request to oracle")
	oracleResp, err := oracle.FundContractRequest(&oracle.FundingRequest{
		ServiceProviderAddr: depReq.RequesterWalletAddress,
		ComputeProviderAddr: computeProvider.TokenomicsAddress,
		EstimatedPrice:      int64(estimatedNTX),
	})
	if err != nil {
		zlog.Info("sending fund contract request to oracle failed")
		return nil, fmt.Errorf("sending fund contract request to oracle failed: %w", err)
	}

	// Seding hashes to CP's DMS in Deployment Request
	depReq.MetadataHash = oracleResp.MetadataHash
	depReq.WithdrawHash = oracleResp.WithdrawHash
	depReq.RefundHash = oracleResp.RefundHash
	depReq.Distribute_50Hash = oracleResp.Distribute_50Hash
	depReq.Distribute_75Hash = oracleResp.Distribute_75Hash

	// Marshal struct to JSON
	depReqBytes, _ := json.Marshal(depReq)
	// Convert JSON bytes to string
	depReqStr := string(depReqBytes)

	depReqFlat = models.DeploymentRequestFlat{} // reset
	depReqFlat.DeploymentRequest = depReqStr
	depReqFlat.JobStatus = "awaiting"

	err = db.DB.Create(&depReqFlat).Error
	if err != nil {
		zlog.Info("cannot write to database")
		return nil, fmt.Errorf("cannot write to database: %w", err)
	}

	// oracle outputs: estimated price, metadata hash, withdraw hash, refund hash, distribute hash
	resp := fundingRespToSPD{
		ComputeProviderAddr: computeProvider.TokenomicsAddress,
		EstimatedPrice:      estimatedNTX,
		MetadataHash:        oracleResp.MetadataHash,
		WithdrawHash:        oracleResp.WithdrawHash,
		RefundHash:          oracleResp.RefundHash,
		Distribute_50Hash:   oracleResp.Distribute_50Hash,
		Distribute_75Hash:   oracleResp.Distribute_75Hash,
	}
	zlog.Sugar().Debugf("%+v", resp)

	go outgoingDepReqWebsock()
	return &resp, nil
}

func DeploymentRequest(ctx, reqCtx context.Context, w http.ResponseWriter, r *http.Request) error {
	ws, err := internal.UpgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		zlog.Sugar().Errorf("failed to set websocket upgrade: %v\n", err)
		return fmt.Errorf("failed to set websocket upgrade: %w", err)
	}

	err = ws.WriteMessage(websocket.TextMessage, []byte("{\"action\": \"connected\", \"message\": \"You are connected to DMS for docker (GPU/CPU) deployment.\"}"))
	if err != nil {
		zlog.Error(err.Error())
		return fmt.Errorf("could not write message to websocket: %w", err)
	}

	depreqWsConn = &internal.WebSocketConnection{Conn: ws}

	go incomingDepReqWebsock(ctx, reqCtx)
	return nil
}

// incomingDepReqWebsock listens for messages from websocket client
func incomingDepReqWebsock(ctx, reqCtx context.Context) {
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
			err = handleWebsocketAction(ctx, reqCtx, p, depreqWsConn)
			if err != nil {
				zlog.Sugar().Errorf("%v", err)
			}
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

func handleWebsocketAction(ctx, reqCtx context.Context, payload []byte, conn *internal.WebSocketConnection) error {
	// convert json to go values
	var m wsMessage
	err := json.Unmarshal(payload, &m)
	if err != nil {
		return fmt.Errorf("wrong message payload: %v", err)
	}

	switch m.Action {
	case "send-status":
		var txStatus models.BlockchainTxStatus
		err := json.Unmarshal(m.Message, &txStatus)
		if err != nil || txStatus.TransactionStatus != "success" {
			// Send event to Signoz
			span := trace.SpanFromContext(reqCtx)
			span.SetAttributes(attribute.String("URL", "/run/deploy"))
			span.SetAttributes(attribute.String("TransactionStatus", "error"))
			return err
		}

		err = sendDeploymentRequest(ctx, reqCtx, conn, txStatus.TxHash)
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}
	return nil
}

func sendDeploymentRequest(ctx, reqCtx context.Context, conn *internal.WebSocketConnection, txHash string) error {
	span := trace.SpanFromContext(reqCtx)
	span.SetAttributes(attribute.String("URL", "/run/deploy"))
	span.SetAttributes(attribute.String("TransactionStatus", "success"))
	kLogger.Info("send deployment request", span)

	defer span.End()

	// load depReq from the database
	var depReqFlat models.DeploymentRequestFlat
	var depReq models.DeploymentRequest
	var service models.Services

	// SELECTs the first record; first record which is not marked as delete
	err := db.DB.First(&depReqFlat).Error
	if err != nil {
		return fmt.Errorf("could not select first record on database: %w", err)
	}

	err = json.Unmarshal([]byte(depReqFlat.DeploymentRequest), &depReq)
	if err != nil {
		return fmt.Errorf("unable to unmarshal deployment request: %w", err)
	}

	depReq.TxHash = txHash

	//extract a span context parameters , so that a child span constructed on the reciever side
	depReq.TraceInfo.SpanID = span.SpanContext().TraceID().String()
	depReq.TraceInfo.TraceID = span.SpanContext().SpanID().String()
	depReq.TraceInfo.TraceFlags = span.SpanContext().TraceFlags().String()
	depReq.TraceInfo.TraceStates = span.SpanContext().TraceState().String()

	zlog.Sugar().Debugf("deployment request: %+v", depReq)

	// Saving service info in SP side
	service.TxHash = txHash
	service.MetadataHash = depReq.MetadataHash
	service.WithdrawHash = depReq.WithdrawHash
	service.RefundHash = depReq.RefundHash
	service.Distribute_50Hash = depReq.Distribute_50Hash
	service.Distribute_75Hash = depReq.Distribute_75Hash
	service.TransactionType = "refund"

	err = db.DB.Create(&service).Error
	if err != nil {
		zlog.Sugar().Errorf("couldn't save service on SP side: %v", err)
	}

	// notify websocket client about the
	submitted := map[string]string{"action": libp2p.JobSubmitted}
	err = conn.WriteJSON(submitted)
	if err != nil {
		return fmt.Errorf("unable to write submit message to websocket: %w", err)
	}

	depReqStream, err := libp2p.SendDeploymentRequest(ctx, depReq)
	if err != nil {
		return fmt.Errorf("failed to send deployment request: %w", err)
	}

	// if this is a resume request, send the file as well
	if depReq.Params.ResumeJob.Resume {
		zlog.Sugar().Infof("sending progress file: %s", depReq.Params.ResumeJob.ProgressFile)
		remotePeerID, err := peer.Decode(depReq.Params.RemoteNodeID)
		if err != nil {
			zlog.Sugar().Errorf("could not decode string ID to peerID: %v", err)
			return fmt.Errorf("could not decode string ID to peerID: %v", err)
		}
		transferChan, err := libp2p.SendFileToPeer(ctx, remotePeerID, depReq.Params.ResumeJob.ProgressFile, libp2p.FTDEPREQ)
		if err != nil {
			zlog.Sugar().Errorf("error: couldn't send file to peer - %v", err)
			return fmt.Errorf("could not send file to peer %s: %w", remotePeerID, err)
		}

		// wait until transferChan is closed, which happens when file is transferred 100%
		zlog.Info("waiting for checkpoint file transfer to complete")
		<-transferChan
	}

	//TODO: Context handling and cancellation on all messaging related code
	//      most importantly, depreq/depres messaging
	//XXX: needs a lot of testing.
	go libp2p.DeploymentUpdateListener(depReqStream)
	return nil
}
