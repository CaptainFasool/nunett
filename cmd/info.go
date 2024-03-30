package cmd

import (
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/cmd/backend"
	"gitlab.com/nunet/device-management-service/models"
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
				return fmt.Errorf("cannot fetch machine metadata: %w", err)
			}

			displayMetadataInTable(cmd.OutOrStdout(), metadata)

			return nil
		},
	}
}

func displayMetadataInTable(w io.Writer, metadata *models.Metadata) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Info", "Value"})

	table.Append([]string{"Name", metadata.Name})
	table.Append([]string{"Update Timestamp", fmt.Sprintf("%d", metadata.UpdateTimestamp)})
	table.Append([]string{"Memory Max", fmt.Sprintf("%d", metadata.Resource.MemoryMax)})
	table.Append([]string{"Total Core", fmt.Sprintf("%d", metadata.Resource.TotalCore)})
	table.Append([]string{"CPU Max", fmt.Sprintf("%d", metadata.Resource.CPUMax)})
	table.Append([]string{"Available CPU", fmt.Sprintf("%d", metadata.Available.CPU)})
	table.Append([]string{"Available Memory", fmt.Sprintf("%d", metadata.Available.Memory)})
	table.Append([]string{"Reserved CPU", fmt.Sprintf("%d", metadata.Reserved.CPU)})
	table.Append([]string{"Reserved Memory", fmt.Sprintf("%d", metadata.Reserved.Memory)})
	table.Append([]string{"Network", metadata.Network})
	table.Append([]string{"Public Key", metadata.PublicKey})
	table.Append([]string{"Node ID", metadata.NodeID})
	table.Append([]string{"Allow Cardano", fmt.Sprintf("%t", metadata.AllowCardano)})
	table.Append([]string{"Dashboard", metadata.Dashboard})
	table.Append([]string{"NTX Price Per Minute", fmt.Sprintf("%f", metadata.NTXPricePerMinute)})

	table.Render()
}
