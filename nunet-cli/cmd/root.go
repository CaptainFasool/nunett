/*
Copyright © 2023 Gustavo Silva <gustavo.silva@nunet.io>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nunet",
	Short: "NuNet DMS CLI utility",
	Long: `A command line interface (CLI) utility for NuNet's Device Management Service (DMS).

	DMS is the main component of NuNet decentralized network which serves the function of connecting both service and compute providers in a peer-to-peer environment.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
