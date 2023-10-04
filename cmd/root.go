package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/docs"
	"gitlab.com/nunet/device-management-service/utils"
)

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(peerCmd)
	rootCmd.AddCommand(onboardCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(capacityCmd)
	rootCmd.AddCommand(resourceConfigCmd)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(walletCmd)
}

var rootCmd = &cobra.Command{
	Use:     "nunet",
	Short:   "NuNet Device Management Service",
	Version: docs.SwaggerInfo.Version,

	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: false,
		HiddenDefaultCmd:  true,
	},
	Long: `The Device Management Service (DMS) Command Line Interface (CLI)`,

	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// pre-run hook to be used by every subcommand (ensures DMS is running before the command logic)
func isDMSRunning() func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		open, err := utils.ListenDMSPort()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		if !open {
			fmt.Printf("Looks like NuNet DMS is not running...\n\nPlease check:\n\tsystemctl status nunet-dms.service\n")
			os.Exit(1)
		}
	}
}
