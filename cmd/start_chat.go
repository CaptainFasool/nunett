package cmd

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/internal/config"
)

func init() {

}

var startChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start chat with a peer",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		// Define interrupt signal (Ctrl+C)
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)

		// peerID passed as argument
		peerID := args[0]
		port := config.GetConfig().Rest.Port

		serverURL := url.URL{
			Scheme:   "ws",
			Host:     "localhost:" + string(port),
			Path:     "/api/v1/peers/chat/start",
			RawQuery: "peerID=" + peerID,
		}

		conn, _, err := websocket.DefaultDialer.Dial(serverURL.String(), nil)
		if err != nil {
			log.Fatalf("Could not connect: %v", err)
		}
		defer conn.Close()

		// Goroutine for receiving messages
		go func() {
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Fatalf("Error trying to read message: %v", err)
				}
				fmt.Println(message)
			}
		}()

		// Goroutine for sending messages
		go func() {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				text := scanner.Text()
				if err := conn.WriteMessage(websocket.TextMessage, []byte(text)); err != nil {
					log.Println("Error while sending message:", err)
					return
				}
			}
			if scanner.Err() != nil {
				log.Println("Error while reading from stdin:", scanner.Err())
			}
		}()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		// Define ping-pong mechanism
		for {
			select {
			case <-ticker.C:
				conn.WriteMessage(websocket.PingMessage, []byte{})
			case <-interrupt:
				log.Println("Received interrupt signal, closing connection")
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
		}
	},
}
