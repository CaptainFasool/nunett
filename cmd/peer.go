package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	peerCmd.AddCommand(peerListCmd)
	peerCmd.AddCommand(peerSelfCmd)
}

var peerCmd = &cobra.Command{
	Use:   "peer",
	Short: "Peer-related operations",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
