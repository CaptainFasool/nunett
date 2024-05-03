package cmd

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/utils"
)

// sendFileCmd represents the command to send a file to a peer over the p2p network, specifying the peer ID, file path, and transfer type.
var sendFileCmd = &cobra.Command{
	Use:   "send-file <peer-id> <file-path> <transfer-type>",
	Short: "Send a file to a peer over the p2p network",
	Long: `Send a file to a peer over the p2p network. 
	       The transfer type should be one of: 0 for FTDEPREQ, 1 for FTMISC.`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		peerID, err := peer.Decode(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid peer ID: %v\n", err)
			return err
		}

		filePath := args[1]
		transferTypeArg := args[2]
		var transferType libp2p.FileTransferType

		switch transferTypeArg {
		case "0":
			transferType = libp2p.FTDEPREQ
		case "1":
			transferType = libp2p.FTMISC
		default:
			fmt.Fprintf(os.Stderr, "Invalid transfer type. Use '0' for FTDEPREQ or '1' for FTMISC.\n")
			return fmt.Errorf("invalid transfer type: %s", transferTypeArg)
		}

		query := url.Values{
			"peerID":       {peerID.String()},
			"filePath":     {filePath},
			"transferType": {fmt.Sprintf("%d", transferType)},
		}.Encode()

		startURL, err := utils.InternalAPIURL("ws", "/api/v1/peers/file/send", query)
		if err != nil {
			return fmt.Errorf("could not compose WebSocket URL: %w", err)
		}

		wsClient, _, err := websocket.DefaultDialer.Dial(startURL, nil)
		if err != nil {
			return fmt.Errorf("failed to initialize WebSocket client: %w", err)
		}
		defer wsClient.Close()

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		go func() {
			for {
				select {
				case <-interrupt:
					cancel()
					wsClient.Close()
					return
				case <-ctx.Done():
					return
				}
			}
		}()

		go func() {
			defer func() {
				wsClient.Close()
			}()

			for {
				_, message, err := wsClient.ReadMessage()
				if err != nil {
					log.Printf("read error: %v\n", err)
					break
				}
				log.Printf("Received: %s\n", message)
			}
		}()

		fmt.Println("\nFile has been sent successfully.")
		return nil
	},
}

// acceptFileCmd represents the command to accept an incoming file transfer
var acceptFileCmd = &cobra.Command{
	Use:   "accept-file",
	Short: "Accept an incoming file transfer",
	Long:  "Accept the most recent incoming file transfer request.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if libp2p.CurrentFileTransfer.InboundFileStream == nil {
			fmt.Println("No incoming file transfer request available.")
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		fmt.Println("Accepting file transfer...")

		filePath, progressChan, err := libp2p.AcceptFileTransfer(ctx, libp2p.CurrentFileTransfer)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to accept file transfer: %v\n", err)
			return err
		}

		fmt.Printf("File transfer initiated, saving to %s\n", filePath)

		for progress := range progressChan {
			fmt.Printf("\rTransfer progress: %.2f%%, %s remaining", progress.Percent, progress.Remaining().Round(time.Second))
		}

		fmt.Println("\nFile transfer completed successfully.")

		libp2p.ClearIncomingFileRequests()

		return nil
	},
}
