package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/docs"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var rootCmd = &cobra.Command{
	Use:     "nunet",
	Short:   "NuNet Device Management Service",
	Version: docs.SwaggerInfo.Version,

	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
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
