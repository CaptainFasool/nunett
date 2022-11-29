// internal modules is a work in progress. It is planned to accomodate
// modules such as db and models.
package internal

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgradeConnection = websocket.Upgrader{
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
// @Summary		Sends a command to specific node and prints back response.
// @Description	Sends a command to specific node and prints back response.
// @Tags		peers
// @Success		200
// @Router		/peers/ws [get]
func HandleWebSocket(c *gin.Context) {
	nodeId := c.Query("nodeID")
	if nodeId == "" {
		c.AbortWithStatusJSON(400, gin.H{"message": "nodeID not provided"})
	}

	ws, err := upgradeConnection.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %+v\n", err)
		return
	}

	welcomeMessage := fmt.Sprintf("Enter the commands that you wish to send to %s and press return.", nodeId)

	err = ws.WriteMessage(websocket.TextMessage, []byte(welcomeMessage))
	if err != nil {
		log.Println(err)
	}

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = nodeId

	go ListenForWs(&conn)
}

// ListenForWs listens to the connected client for any message. It is assumed that
// every message that is coming is a command to be executed.
func ListenForWs(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error:", fmt.Sprintf("%v", r))
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
		log.Printf("%v", command)
		// TO BE IMPLEMENTED
		// send command

		// fetch result

		// send back result
		command.Conn.WriteMessage(websocket.TextMessage, []byte(command.Command))
	}
}
