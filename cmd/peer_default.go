package cmd

import (
	"fmt"

	"github.com/buger/jsonparser"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/cmd/backend"
)

var peerDefaultCmd = NewPeerDefaultCmd(utilsService)

func NewPeerDefaultCmd(utilsService backend.Utility) *cobra.Command {
	return &cobra.Command{
		Use:    "default [PEER_ID]",
		Short:  "Retrieve or Set default peer for deployment requests",
		Hidden: true,
		Args:   cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var peerID string
			if len(args) > 0 {
				peerID = args[0]
			}

			query := ""
			if peerID != "" {
				query = "peerID=" + peerID
			}

			body, err := utilsService.ResponseBody(nil, "GET", "/api/v1/peers/depreq", query, nil)
			if err != nil {
				return fmt.Errorf("error making request: %w", err)
			}

			if errMsg, err := jsonparser.GetString(body, "error"); err == nil {
				return fmt.Errorf("error: %s", errMsg)
			}

			message, err := jsonparser.GetString(body, "message")
			if err != nil {
				return fmt.Errorf("error parsing response message: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), message)

			return nil
		},
	}
}
