package machines

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gitlab.com/nunet/device-management-service/internal"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"go.opentelemetry.io/otel"
)

type wsMessage struct {
	Action  string          `json:"action"`
	Message json.RawMessage `json:"message"`
}

// HandleDeploymentRequest  godoc
// @Summary      Search devices on DHT with appropriate machines and sends a deployment request.
// @Description  HandleDeploymentRequest searches the DHT for non-busy, available devices with appropriate metadata. Then sends a deployment request to the first machine
// @Success      200  {string}  string
// @Router       /run/deploy [get]
func HandleDeploymentRequest(c *gin.Context) {
	ws, err := internal.UpgradeConnection.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Error(fmt.Sprintf("Failed to set websocket upgrade: %+v\n", err))
		return
	}

	err = ws.WriteMessage(websocket.TextMessage, []byte("You are connected to DMS for GPU deployment."))
	if err != nil {
		zlog.Error(err.Error())
	}

	conn := internal.WebSocketConnection{Conn: ws}

	go listenForDeploymentRequest(c, &conn)
	go listenForDeploymentResponse(c, &conn)
}

func listenForDeploymentRequest(ctx context.Context, conn *internal.WebSocketConnection) {
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

func listenForDeploymentResponse(ctx context.Context, conn *internal.WebSocketConnection) {
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
				conn.WriteMessage(websocket.TextMessage, msg)
			} else {
				zlog.Info("Channel closed!")
			}
		}
	}
}

func handleWebsocketAction(ctx context.Context, payload []byte) {
	// convert json to go values
	var m wsMessage
	err := json.Unmarshal(payload, &m)
	if err != nil {
		zlog.Error(fmt.Sprintf("wrong message payload: %v", err))
	}

	switch m.Action {
	case "deployment-request":
		err := sendDeploymentRequest(ctx, m.Message)
		if err != nil {
			zlog.Error(err.Error())
		}
	}
}

func sendDeploymentRequest(ctx context.Context, requestParams json.RawMessage) error {
	// parse the body, get service type, and filter devices
	var depReq models.DeploymentRequest
	if err := json.Unmarshal([]byte(requestParams), &depReq); err != nil {
		return errors.New("invalid deployment request body")
	}
	// add node_id and public_key in deployment request
	pKey, err := libp2p.GetP2P().Host.ID().ExtractPublicKey()
	if err != nil {
		return fmt.Errorf("Unable to Extract Public Key: %v", err)
	}
	selfNodeID := libp2p.GetP2P().Host.ID().String()

	depReq.Params.NodeID = selfNodeID
	depReq.Params.PublicKey = pKey.Type().String()

	// check if the pricing matched
	if estimatedNtx := CalculateStaticNtxGpu(depReq); estimatedNtx > float64(depReq.MaxNtx) {
		return errors.New("nunet estimation price is greater than client price")
	}

	peers := FilterPeers(depReq, libp2p.GetP2P().Host)

	//create a new span
	_, span := otel.Tracer("SendDeploymentRequest").Start(context.Background(), "SendDeploymentRequest")
	defer span.End()

	//extract a span context parameters , so that a child span constructed on the reciever side
	depReq.TraceInfo.SpanID = span.SpanContext().TraceID().String()
	depReq.TraceInfo.TraceID = span.SpanContext().SpanID().String()
	depReq.TraceInfo.TraceFlags = span.SpanContext().TraceFlags().String()
	depReq.TraceInfo.TraceStates = span.SpanContext().TraceState().String()

	// pick a peer from the list and send a message to the nodeID of the peer.
	selectedNode := peers[0]
	depReq.Timestamp = time.Now()

	depReqStream, err := libp2p.SendDeploymentRequest(ctx, selectedNode, depReq)
	if err != nil {
		return err
	}

	//TODO: Context handling and cancellation on all messaging related code
	//      most importantly, depreq/depres messaging
	//XXX: needs a lot of testing.
	go libp2p.DeploymentResponseListener(depReqStream)

	return nil
}
