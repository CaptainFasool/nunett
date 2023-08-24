package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	joinCmd.AddCommand(joinChatCmd)
}

var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join a chat stream",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
