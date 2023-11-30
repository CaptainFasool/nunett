package cmd

import (
	"fmt"

	"github.com/buger/jsonparser"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/utils"
)

var (
	flagForce bool

	offboardCmd = NewOffboardCmd(&Utils{})
)

func NewOffboardCmd(util Utility) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "offboard",
		Short:  "Offboard the device from NuNet",
		PreRun: isDMSRunning(),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				onboarded bool
				confirmed bool

				force  bool
				query  string
				body   []byte
				errMsg string
				msg    string

				err error
			)

			onboarded, err = util.IsOnboarded()
			if err != nil {
				return fmt.Errorf("could not check onboard status: %w", err)
			}
			if !onboarded {
				return fmt.Errorf("looks like your machine is not onboarded")
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Warning: Offboarding will remove all your data and you will not be able to onboard again with the same identity")

			confirmed, err = utils.PromptYesNo("Are you sure you want to offboard? (y/N)")
			if err != nil {
				return fmt.Errorf("could not prompt for confirmation: %w", err)
			}
			if !confirmed {
				return fmt.Errorf("offboard aborted by user")
			}

			force, _ = cmd.Flags().GetBool("force")
			if force {
				query = fmt.Sprintf("force=%t", force)
			}

			body, err = util.ResponseBody(nil, "DELETE", "/api/v1/onboarding/offboard", query, nil)
			if err != nil {
				return fmt.Errorf("unable to fetch response body: %w", err)
			}

			errMsg, err = jsonparser.GetString(body, "error")
			if err == nil {
				return fmt.Errorf("got error response from server: %w", errMsg)
			} else if err != jsonparser.KeyPathNotFoundError {
				return fmt.Errorf("could not parse string response: %w", err)
			}

			msg, err = jsonparser.GetString(body, "message")
			if err != nil {
				return fmt.Errorf("failed to get message string from response body: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), msg)

			return nil
		},
	}
	cmd.Flags().BoolVarP(&flagForce, "force", "f", false, "force offboarding")

	return cmd
}
