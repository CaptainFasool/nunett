package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/cmd/backend"
)

var chatClearCmd = NewChatClearCmd(p2pService)

func NewChatClearCmd(p2pService backend.PeerManager) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear open chat streams",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := p2pService.ClearIncomingChatRequests()
			if err != nil {
				return fmt.Errorf("clear chat failed: %w", err)
			}

			return nil
		},
	}
}
