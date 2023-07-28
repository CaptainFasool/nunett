// Package internal is a work in progress. It is planned to accomodate
// modules such as db and models.
package internal

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// UpgradeConnection is generic protocol upgrader for entire DMS.
var UpgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// WebSocketConnection is pointer to gorilla/websocket.Conn
type WebSocketConnection struct {
	*websocket.Conn
}

// Command represents a command to be executed
type Command struct {
	Command string
	NodeID  string // ID of the node where command will be executed
	Result  string
	Conn    *WebSocketConnection
}

var commandChan = make(chan Command)
var clients = make(map[WebSocketConnection]string)

// HandleWebSocket godoc
//	@Summary		Sends a command to specific node and prints back response.
//	@Description	Sends a command to specific node and prints back response.
//	@Tags			peers
//	@Success		200
//	@Router			/peers/ws [get]
func HandleWebSocket(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/peers/ws"))

	nodeID := c.Query("nodeID")
	if nodeID == "" {
		c.AbortWithStatusJSON(400, gin.H{"message": "nodeID not provided"})
	}

	ws, err := UpgradeConnection.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Sugar().Errorf("Failed to set websocket upgrade: %+v\n", err)
		return
	}

	welcomeMessage := fmt.Sprintf("Enter the commands that you wish to send to %s and press return.", nodeID)

	err = ws.WriteMessage(websocket.TextMessage, []byte(welcomeMessage))
	if err != nil {
		zlog.Error(err.Error())
	}

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = nodeID

	go ListenForWs(&conn)
}

// ListenForWs listens to the connected client for any message. It is assumed that
// every message that is coming is a command to be executed.
func ListenForWs(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			zlog.Sugar().Errorf("Error:", fmt.Sprintf("%v", r))
		}
	}()

	cmd := Command{NodeID: clients[*conn], Conn: conn}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			// do nothing
		} else {
			// logic to send command and fetch the output
			cmd.Command = string(msg)
			commandChan <- cmd
		}
	}
}

// SendCommandForExecution work is to send command for execution and fetch the result
// This function listens for new commands from commandChan
func SendCommandForExecution() {
	for {
		command := <-commandChan
		zlog.Sugar().Infof("%v", command)
		// TO BE IMPLEMENTED
		// send command

		// fetch result

		// send back result
		command.Conn.WriteMessage(websocket.TextMessage, []byte(command.Command))
	}
}
