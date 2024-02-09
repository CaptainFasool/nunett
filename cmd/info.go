package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/cmd/backend"
)

var infoCmd = NewInfoCmd(networkService, utilsService)

func NewInfoCmd(net backend.NetworkManager, utilsService backend.Utility) *cobra.Command {
	return &cobra.Command{
		Use:     "info",
		Short:   "Display information about onboarded device",
		Long:    "Display metadata of onboarded device on Nunet Device Management Service",
		PreRunE: isDMSRunning(net),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := checkOnboarded(utilsService)
			if err != nil {
				return err
			}

			metadata, err := utilsService.ReadMetadataFile()
			if err != nil {
				return fmt.Errorf("cannot read metadata file: %w", err)
			}

			printMetadata(cmd.OutOrStdout(), metadata)
			return nil
		},
	}
}
