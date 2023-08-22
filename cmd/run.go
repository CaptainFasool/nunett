package cmd

import (
    "github.com/spf13/cobra"
    "gitlab.com/nunet/device-management-service/dms"
)

func init() {
	rootCmd.AddCommand(dmsRunCmd)
}

var dmsRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the Device Management Service",
	Long:  `The Device Management Service (DMS) is a system application for computing and service providers. It handles networking and device management.`,

	Run: func(cmd *cobra.Command, args []string) {
		dms.Run()
	},
}
