package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/utils"
)

func init() {
}

var joinChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Join open chat stream",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		query := "streamID=" + args[0]

		joinURL, err := utils.InternalAPIURL("ws", "/api/v1/peers/chat/join", query)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		client := &Client{}
		err = client.Initialize(joinURL)
		if err != nil {
			fmt.Println("Failed to initialize client:", err)
			os.Exit(1)
		}
		defer client.Conn.Close()

		go client.ReadMessages()
		go client.WriteMessages()
		client.HandleInterruptsAndPings()
	},
}
