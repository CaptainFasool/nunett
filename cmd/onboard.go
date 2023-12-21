package cmd

import (
	"fmt"

	"github.com/buger/jsonparser"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/cmd/backend"
)

var (
	onboardCmd                     = NewOnboardCmd(networkService, utilsService)
	flagCpu, flagMemory            int64
	flagChan, flagAddr, flagPlugin string
	flagCardano, flagLocal         bool
)

func NewOnboardCmd(net backend.NetworkManager, utilsService backend.Utility) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "onboard",
		Short:   "Onboard current machine to NuNet",
		PreRunE: isDMSRunning(net),
		RunE: func(cmd *cobra.Command, args []string) error {
			memory, _ := cmd.Flags().GetInt64("memory")
			cpu, _ := cmd.Flags().GetInt64("cpu")
			channel, _ := cmd.Flags().GetString("nunet-channel")
			address, _ := cmd.Flags().GetString("address")
			local, _ := cmd.Flags().GetBool("local-enable")
			cardano, _ := cmd.Flags().GetBool("cardano")

			if memory == 0 || cpu == 0 || channel == "" || address == "" {
				return fmt.Errorf("missing at least one required flag")
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Checking onboard status...")

			onboarded, err := utilsService.IsOnboarded()
			if err != nil {
				return fmt.Errorf("could not check onboard status: %w", err)
			}

			if onboarded {
				err := promptReonboard(cmd.InOrStdin(), cmd.OutOrStdout())
				if err != nil {
					return err
				}
			}

			onboardJson, err := setOnboardData(memory, cpu, channel, address, cardano, local)
			if err != nil {
				return fmt.Errorf("failed to set onboard data: %w", err)
			}

			body, err := utilsService.ResponseBody(nil, "POST", "/api/v1/onboarding/onboard", "", onboardJson)
			if err != nil {
				return fmt.Errorf("could not get response body: %w", err)
			}

			errMsg, err := jsonparser.GetString(body, "error")
			if err == nil { // if error message IS found
				return fmt.Errorf(errMsg)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Sucessfully onboarded!")
			return nil
		},
	}

	cmd.Flags().Int64VarP(&flagMemory, "memory", "m", 0, "set value for memory usage")
	cmd.Flags().Int64VarP(&flagCpu, "cpu", "c", 0, "set value for CPU usage")
	cmd.Flags().StringVarP(&flagChan, "nunet-channel", "n", "", "set channel")
	cmd.Flags().StringVarP(&flagAddr, "address", "a", "", "set wallet address")
	cmd.Flags().StringVarP(&flagPlugin, "plugin", "p", "", "set plugin")
	cmd.Flags().BoolVarP(&flagLocal, "local-enable", "l", true, "set server mode (enable for local)")
	cmd.Flags().BoolVarP(&flagCardano, "cardano", "C", false, "set Cardano wallet")
	return cmd
}
