package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/utils"
)

func init() {

}

var startChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start chat with a peer",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
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
		}()
		go func() {
			client.WriteMessages()
			wg.Done()
		}()

		go func() {
			client.HandleInterruptsAndPings()
			wg.Done()
		}()

		wg.Wait()
	},
}
