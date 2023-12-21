package cmd

import "github.com/spf13/cobra"

var flagNode string

func init() {
	shellCmd.Flags().StringVar(&flagNode, "node-id", "", "set nodeID value")
}

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Send commands to DMS instance",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
