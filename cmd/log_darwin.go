//go:build darwin

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:     "log",
	Short:   "Gather all logs into a tarball",
	PreRunE: isDMSRunning(networkService),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "Log collection on MacOS is not yet supported.")
	},
}
