package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	startCmd.AddCommand(startChatCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Initiate chat with peer",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
