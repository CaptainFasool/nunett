package cmd

import (
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"

	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/cmd/backend"
)

func NewSendFileCmd(utilsService backend.Utility) *cobra.Command {
	return &cobra.Command{
		Use:   "send-file <peer-id> <file-path>",
		Short: "Send a file to a peer over the p2p network via WebSocket",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Perform any necessary onboarding checks
			if err := checkOnboarded(utilsService); err != nil {
				log.Fatal("Onboarding check failed:", err)
			}

			peerID, filePath := args[0], args[1]
			sendFileViaWebSocket(peerID, filePath)
		},
	}
}

func sendFileViaWebSocket(peerID, filePath string) {
	wsURL := url.URL{Scheme: "ws", Host: "localhost:1236", Path: "/api/v1/peers/file/send",
		RawQuery: fmt.Sprintf("peerID=%s&filePath=%s", url.QueryEscape(peerID), url.QueryEscape(filePath))}

	c, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	log.Println("Connection established. Waiting for progress updates...")
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		log.Printf("Received: %s", message)
	}
}
