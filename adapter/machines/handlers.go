package machines

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gitlab.com/nunet/device-management-service/adapter"
	"gitlab.com/nunet/device-management-service/internal"
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

	go listenForDeploymentRequest(&conn)
	go listenForDeploymentResponse(&conn)
}

func listenForDeploymentRequest(conn *internal.WebSocketConnection) {
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

		handleWebsocketAction(p)
	}
}

func listenForDeploymentResponse(conn *internal.WebSocketConnection) {
	for {
		// 1. check if DepResQueue has anything
		select {
		case adapterMsg, ok := <-adapter.DepResQueue:
			if ok {
				zlog.Info("Deployment response received. Sending it to connected websocket client")
				var depRes models.DeploymentResponse

				jsonDataMsg, _ := json.Marshal(adapterMsg.Data.Message)
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

func handleWebsocketAction(payload []byte) {
	// convert json to go values
	var m wsMessage
	err := json.Unmarshal(payload, &m)
	if err != nil {
		zlog.Error(fmt.Sprintf("wrong message payload: %v", err))
	}

	switch m.Action {
	case "deployment-request":
		err := sendDeploymentRequest(m.Message)
		if err != nil {
			zlog.Error(err.Error())
		}
	}
}

func sendDeploymentRequest(requestParams json.RawMessage) error {
	// parse the body, get service type, and filter devices
	var depReq models.DeploymentRequest
	if err := json.Unmarshal([]byte(requestParams), &depReq); err != nil {
		return errors.New("invalid deployment request body")
	}
	// add node_id and public_key in deployment request
	pKey, err := adapter.GetMasterPKey()
	if err != nil {
		return err
	}
	selfNodeID := adapter.GetPeerID()

	depReq.Params.NodeID = selfNodeID
	depReq.Params.PublicKey = pKey

	// check if the pricing matched
	if estimatedNtx := CalculateStaticNtxGpu(depReq); estimatedNtx > float64(depReq.MaxNtx) {
		return errors.New("nunet estimation price is greater than client price")
	}

	// filter peers based on passed criteria
	// TODO: Replace with the following
	// peers := libp2p.FilterPeers(depReq, libp2p.Node)
	peers, err := FilterPeers(depReq)
	if err != nil {
		return err
	}

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

	out, err := json.Marshal(depReq)
	if err != nil {
		return errors.New("error converting deployment request body to string")
	}
	// TODO: Replace with libp2p messaging
	response, err := adapter.SendMessage(selectedNode.PeerInfo.NodeID, string(out))
	if err != nil {
		return errors.New("cannot send message to the peer")
	}

	_ = response // do we send any response back?

	return nil
}

// ListPeers  godoc
// @Summary      Return list of peers currently connected to
// @Description  Gets a list of peers the adapter can see within the network and return a list of peer info
// @Tags         run
// @Produce      json
// @Success      200  {string}	string
// @Router       /peers [get]
func ListPeers(c *gin.Context) {
	response, err := adapter.FetchMachines()
	if err != nil {
		c.JSON(500, gin.H{"error": "can not fetch peers"})
		panic(err)
	}
	c.JSON(200, response)

}
