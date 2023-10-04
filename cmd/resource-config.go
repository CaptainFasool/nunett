package cmd

import (
	"fmt"
	"os"

	"github.com/buger/jsonparser"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/utils"
)

func init() {
	resourceConfigCmd.Flags().Int64VarP(&flagMemory, "memory", "m", 0, "set amount of memory")
	resourceConfigCmd.Flags().Int64VarP(&flagCpu, "cpu", "c", 0, "set amount of CPU")
}

var resourceConfigCmd = &cobra.Command{
	Use:    "resource-config",
	Short:  "Update configuration of onboarded device",
	Long:   "",
	PreRun: isDMSRunning(),
	Run: func(cmd *cobra.Command, args []string) {
		memory, _ := cmd.Flags().GetInt64("memory")
		cpu, _ := cmd.Flags().GetInt64("cpu")

		// check for both flags values
		if memory == 0 || cpu == 0 {
			fmt.Println(`Error: all flag values must be specified

For updating resources, check:
    nunet resource-config --help`)
			os.Exit(1)
		}

		onboarded, err := utils.IsOnboarded()
		if err != nil {
			fmt.Println("Error checking onboard status:", err)
			os.Exit(1)
		}

		if !onboarded {
			fmt.Println(`Looks like your machine is not onboarded...

For onboarding, check:
    nunet onboard --help`)
			os.Exit(1)
		}

		// set data for body request
		resourceBody, err := setOnboardData(memory, cpu, "", "", false, false)
		if err != nil {
			fmt.Println("Error setting onboard data:", err)
			os.Exit(1)
		}

		resp, err := utils.ResponseBody(nil, "POST", "/api/v1/onboarding/resource-config", resourceBody)
		if err != nil {
			fmt.Println("Error getting response body:", err)
			os.Exit(1)
		}

		msg, err := jsonparser.GetString(resp, "error")
		if err == nil { // if error message IS found
			fmt.Println("Error:", msg)
			os.Exit(1)
		} else if err == jsonparser.KeyPathNotFoundError { // if NO error message is found
			fmt.Println("Resources updated successfully!")
			fmt.Println(string(resp))
			os.Exit(0)
		} else { // another error
			fmt.Println("Error getting error message from response body:", err)
			os.Exit(1)
		}
	},
}
