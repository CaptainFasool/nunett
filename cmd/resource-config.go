package cmd

import (
	"fmt"

	"github.com/buger/jsonparser"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/cmd/backend"
)

var resourceConfigCmd = NewResourceConfigCmd(networkService, utilsService)

func NewResourceConfigCmd(net backend.NetworkManager, utilsService backend.Utility) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resource-config",
		Short:   "Update configuration of onboarded device",
		PreRunE: isDMSRunning(net),
		RunE: func(cmd *cobra.Command, args []string) error {
			memory, _ := cmd.Flags().GetInt64("memory")
			cpu, _ := cmd.Flags().GetInt64("cpu")

			// check for both flags values
			if memory == 0 || cpu == 0 {
				return fmt.Errorf("all flag values must be specified")
			}

			err := checkOnboarded(utilsService)
			if err != nil {
				return err
			}

			// set data for body request
			resourceBody, err := setOnboardData(memory, cpu, "", "", false, false, true)
			if err != nil {
				return fmt.Errorf("failed to set onboard data: %w", err)
			}

			resp, err := utilsService.ResponseBody(nil, "POST", "/api/v1/onboarding/resource-config", "", resourceBody)
			if err != nil {
				return fmt.Errorf("could not get response body: %w", err)
			}

			msg, err := jsonparser.GetString(resp, "error")
			if err == nil { // if error message IS found
				return fmt.Errorf(msg)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Resources updated successfully!")
			fmt.Fprintln(cmd.OutOrStdout(), string(resp))
			return nil
		},
	}

	cmd.Flags().Int64VarP(&flagMemory, "memory", "m", 0, "set amount of memory")
	cmd.Flags().Int64VarP(&flagCpu, "cpu", "c", 0, "set amount of CPU")
	return cmd
}
