package cmd

import "github.com/spf13/cobra"

func init() {
	gpuCmd.AddCommand(gpuCapacityCmd)
	gpuCmd.AddCommand(gpuStatusCmd)
	gpuCmd.AddCommand(gpuOnboardCmd)
}

var gpuCmd = &cobra.Command{
	Use:   "gpu",
	Short: "GPU-related operations",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
