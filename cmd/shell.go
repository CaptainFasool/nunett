package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/utils"
)

var flagNode string

func init() {
	shellCmd.Flags().StringVar(&flagNode, "node-id", "", "set nodeID value")
}

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Send commands to DMS instance",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		nodeID, _ := cmd.Flags().GetString("node-id")
		query := ""

		if nodeID != "" {
			query = "nodeID=" + nodeID
		} else {
			fmt.Println("Error: --node-id is not specified")
			fmt.Println("See 'nunet shell --help' for more information")
			os.Exit(0)
		}

		shellURL, err := utils.InternalAPIURL("ws", "/api/v1/peers/ws", query)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		client := &Client{}
		err = client.Initialize(shellURL)
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
