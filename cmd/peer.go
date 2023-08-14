package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	peerCmd.AddCommand(selfPeerCmd)
}

var peerCmd = &cobra.Command{
	Use:   "peer",
	Short: "Display information about peers",
    Long: `Usage: nunet peer COMMAND [OPTIONS]
    Display connected peers and self peer information

    Commands:
    list    list visible peers
    self    show self peer info

    For more information, visit: https://gitlab.com/nunet/device-management-service/-/wikis/home`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var selfPeerCmd = &cobra.Command{
	Use:   "self",
	Short: "Show self peer",
	Long: `Show self peer`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Show self peer")
	},
}
