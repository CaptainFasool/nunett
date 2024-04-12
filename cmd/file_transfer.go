package cmd

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/cmd/backend"
)

func NewSendFileCmd(utilsService backend.Utility, wsClient *backend.WebSocket) *cobra.Command {
	return &cobra.Command{
		Use:   "send-file <peer-id> <file-path>",
		Short: "Send a file to a peer over the p2p network via WebSocket",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			onboarded, err := utilsService.IsOnboarded()
			if err != nil || !onboarded {
				return fmt.Errorf("onboarding check failed: %w", err)
			}

			return sendFileViaWebSocket(wsClient, args[0], args[1], cmd)
		},
	}
}

// Handler to send files using WebSocket
func sendFileViaWebSocket(wsClient *backend.WebSocket, peerID, filePath string, cmd *cobra.Command) error {
	url := fmt.Sprintf("ws://localhost:1236/api/v1/peers/file/send?peerID=%s&filePath=%s", url.QueryEscape(peerID), url.QueryEscape(filePath))
	if err := wsClient.Initialize(url); err != nil {
		return fmt.Errorf("websocket initialization failed: %w", err)
	}
	defer wsClient.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandler(cancel)

	return handleWebSocketCommunication(wsClient, ctx, cmd)
}

func NewAcceptFileCmd(utilsService backend.Utility, wsClient *backend.WebSocket) *cobra.Command {
	return &cobra.Command{
		Use:   "accept-file <stream-id>",
		Short: "Accept an incoming file transfer via WebSocket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			onboarded, err := utilsService.IsOnboarded()
			if err != nil || !onboarded {
				return fmt.Errorf("onboarding check failed: %w", err)
			}

			return acceptFileViaWebSocket(wsClient, args[0], cmd)
		},
	}
}

// Handler to accept files using WebSocket
func acceptFileViaWebSocket(wsClient *backend.WebSocket, streamID string, cmd *cobra.Command) error {
	url := fmt.Sprintf("ws://localhost:1236/api/v1/peers/file/accept?streamID=%s", url.QueryEscape(streamID))
	if err := wsClient.Initialize(url); err != nil {
		return fmt.Errorf("websocket initialization failed: %w", err)
	}
	defer wsClient.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandler(cancel)

	return handleWebSocketCommunication(wsClient, ctx, cmd)
}

// Handle incoming and outgoing WebSocket messages
func handleWebSocketCommunication(wsClient *backend.WebSocket, ctx context.Context, cmd *cobra.Command) error {
	go func() {
		if err := wsClient.ReadMessage(ctx, cmd.OutOrStdout()); err != nil {
			log.Printf("Read error: %v\n", err)
		}
	}()

	go func() {
		if err := wsClient.WriteMessage(ctx, cmd.InOrStdin()); err != nil {
			log.Printf("Write error: %v\n", err)
		}
	}()

	return wsClient.Ping(ctx, cmd.OutOrStderr())
}

// Signal handler for graceful shutdown
func setupSignalHandler(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			cancel()
		}
	}()
}
