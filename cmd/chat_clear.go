package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/libp2p"
)

func init() {

}

var chatClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear open chat streams",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		err := libp2p.ClearIncomingChatRequests()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		os.Exit(0)
	},
}
