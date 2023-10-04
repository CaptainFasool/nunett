package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	walletCmd.AddCommand(walletNewCmd)
}

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Wallet Management",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
