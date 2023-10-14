//go:build darwin

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {

}

var logCmd = &cobra.Command{
	Use:    "log",
	Short:  "Gather all logs into a tarball",
	Long:   "",
	PreRun: isDMSRunning(),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Log collection on MacOS is not yet supported.")
	},
}
