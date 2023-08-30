package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	clearCmd.AddCommand(clearChatCmd)
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear open chat streams",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
