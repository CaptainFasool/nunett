package cmd

import (
    "github.com/spf13/cobra"
)

func init() {
    chatCmd.AddCommand(chatClearCmd)
    chatCmd.AddCommand(chatJoinCmd)
    chatCmd.AddCommand(chatListCmd)
    chatCmd.AddCommand(chatStartCmd)
}

var chatCmd = &cobra.Command{
    Use:    "chat",
    Short:  "Chat-related operations",
    Long:   ``,
    Run: func(cmd *cobra.Command, args []string){
        cmd.Help()
    },
}
