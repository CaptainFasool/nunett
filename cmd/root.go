package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/docs"
)

func init() {
	rootCmd.AddCommand(gpuCmd)
	rootCmd.AddCommand(offboardCmd)
	rootCmd.AddCommand(onboardMLCmd)
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(peerCmd)
	rootCmd.AddCommand(onboardCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(deviceCmd)
	rootCmd.AddCommand(capacityCmd)
	rootCmd.AddCommand(resourceConfigCmd)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(walletCmd)
}

var rootCmd = &cobra.Command{
	Use:     "nunet",
	Short:   "NuNet Device Management Service",
	Long:    `The Device Management Service (DMS) Command Line Interface (CLI)`,
	Version: docs.SwaggerInfo.Version,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: false,
		HiddenDefaultCmd:  true,
	},
	SilenceErrors: true,
	SilenceUsage:  true,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	// CheckErr prints formatted error message, if there is any, and exits
	cobra.CheckErr(rootCmd.Execute())
}
