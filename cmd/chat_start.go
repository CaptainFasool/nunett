package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"

	"gitlab.com/nunet/device-management-service/utils"
)

func init() {

}

var chatStartCmd = &cobra.Command{
	Use:     "start",
	Short:   "Start chat with a peer",
	Example: "nunet chat start <peerID>",
	Long:    "",
	Run: func(cmd *cobra.Command, args []string) {
		err := validateStartChatInput(args)
		if err != nil {
			fmt.Println("Error:", err)
			fmt.Printf("\nFor starting chats, check:\n\tnunet chat start --help\n")

			os.Exit(1)
		}

		query := "peerID=" + args[0]

		startURL, err := utils.InternalAPIURL("ws", "/api/v1/peers/chat/start", query)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		client := &Client{}
		err = client.Initialize(startURL)
		if err != nil {
			fmt.Println("Failed to initialize client:", err)
			os.Exit(1)
		}
		defer client.Conn.Close()

		var wg sync.WaitGroup
		wg.Add(3)

		go func() {
			client.ReadMessages()
			wg.Done()
			client.stop <- true
		}()
		go func() {
			client.WriteMessages()
			wg.Done()
			client.stop <- true
		}()

		go func() {
			client.HandleInterruptsAndPings()
			wg.Done()
			client.stop <- true
		}()

		wg.Wait()
	},
}

func validateStartChatInput(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no peer ID specified")
	} else if len(args) > 1 {
		return fmt.Errorf("unable to start multiple chats")
	} else {
		_, err := peer.Decode(args[0])
		if err != nil {
			return fmt.Errorf("argument is not a valid peer ID")
		}
	}

	return nil
}
