package cmd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/buger/jsonparser"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/utils"
)

var flagForce bool

var offboardCmd = &cobra.Command{
	Use:     "offboard",
	Short:   "Offboard the device from NuNet",
	Long:    ``,
	PreRunE: isDMSRunning(networkService),
	Run: func(cmd *cobra.Command, args []string) {
		err := checkOnboarded(utilsService)
		if err != nil {
			fmt.Println("Machine isn't onboarded:", err)
			return
		}

		fmt.Println("Warning: Offboarding will remove all your data and you will not be able to onboard again with the same identity")
		answer, err := utils.PromptYesNo(cmd.InOrStdin(), cmd.OutOrStdout(), "Are you sure you want to offboard?")
		if err != nil {
			fmt.Println("Error reading answer for onboard prompt:", err)
			os.Exit(1)
		}

		if !answer {
			fmt.Println("Exiting...")
			os.Exit(1)
		} else {
			force, _ := cmd.Flags().GetBool("force")
			query := bytes.NewBufferString(fmt.Sprintf(`{"force": %t}`, force))

			body, err := utils.ResponseBody(nil, "POST", "/api/v1/onboarding/offboard", "", query.Bytes())
			if err != nil {
				fmt.Println("Error getting response body:", err)
				os.Exit(1)
			}

			if errMsg, err := jsonparser.GetString(body, "error"); err == nil { // if field "error" IS found
				fmt.Println("Error:", errMsg)
				os.Exit(1)
			} else if err == jsonparser.KeyPathNotFoundError { // if field "error" is NOT found
				msg, _ := jsonparser.GetString(body, "message")
				fmt.Println(msg)
			} else { // if another error occurred
				fmt.Println("Error parsing response:", err)
				os.Exit(1)
			}

			os.Exit(0)

		}
	},
}
