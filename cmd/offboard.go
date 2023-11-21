package cmd

import (
	"fmt"

	"github.com/buger/jsonparser"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/utils"
)

var flagForce bool

func init() {
	offboardCmd.Flags().BoolVarP(&flagForce, "force", "f", false, "force offboarding")
}

var offboardCmd = &cobra.Command{
	Use:    "offboard",
	Short:  "Offboard the device from NuNet",
	PreRun: isDMSRunning(),
	RunE: func(cmd *cobra.Command, args []string) error {
		onboarded, err := utils.IsOnboarded()
		if err != nil {
			return fmt.Errorf("could not check onboard status: %w", err)
		}

		if !onboarded {
			return fmt.Errorf("looks like your machine is not onboarded")
		}

		fmt.Println("Warning: Offboarding will remove all your data and you will not be able to onboard again with the same identity")
		answer := utils.PromptYesNo("Are you sure you want to offboard? (y/N)")
		if !answer {
			return fmt.Errorf("offboard aborted by user")
		} else {
			force, _ := cmd.Flags().GetBool("force")
			if !force {
				body, err := utils.ResponseBody(nil, "DELETE", "/api/v1/onboarding/offboard", "", nil)
				if err != nil {
					return fmt.Errorf("unable to fetch response body: %w", err)
				}

				if errMsg, err := jsonparser.GetString(body, "error"); err == nil { // if field "error" IS found
					return fmt.Errorf("got error response from server: %w", errMsg)
				} else if err == jsonparser.KeyPathNotFoundError { // if field "error" is NOT found
					msg, _ := jsonparser.GetString(body, "message")
					fmt.Println(msg)
				} else { // if another error occurred
					return fmt.Errorf("could not parse string response: %w", err)
				}
			} else {
				query := fmt.Sprintf("force=%t", force)

				body, err := utils.ResponseBody(nil, "DELETE", "/api/v1/onboarding/offboard", query, nil)
				if err != nil {
					return fmt.Errorf("unable to fetch response body: %w", err)
				}

				if errMsg, err := jsonparser.GetString(body, "error"); err == nil { // if field "error" IS found
					return fmt.Errorf("got error response from server: %w", errMsg)
				} else if err == jsonparser.KeyPathNotFoundError { // if field "error" is NOT found
					msg, _ := jsonparser.GetString(body, "message")
					fmt.Println(msg)
				} else { // if another error occurred
					return fmt.Errorf("could not parse string response: %w", err)
				}
			}
		}

		return nil
	},
}
