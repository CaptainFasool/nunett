package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/buger/jsonparser"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
)

var flagCpu int64
var flagMemory int64
var flagChan, flagAddr, flagPlugin string
var flagCardano, flagLocal bool

func init() {
	onboardCmd.Flags().Int64VarP(&flagMemory, "memory", "m", 0, "set value for memory usage")
	onboardCmd.Flags().Int64VarP(&flagCpu, "cpu", "c", 0, "set value for CPU usage")
	onboardCmd.Flags().StringVarP(&flagChan, "nunet-channel", "n", "", "set channel")
	onboardCmd.Flags().StringVarP(&flagAddr, "address", "a", "", "set wallet address")
	onboardCmd.Flags().StringVarP(&flagPlugin, "plugin", "p", "", "set plugin")
	onboardCmd.Flags().BoolVarP(&flagLocal, "local-enable", "l", true, "set server mode (enable for local)")
	onboardCmd.Flags().BoolVarP(&flagCardano, "cardano", "C", false, "set Cardano wallet")
}

var onboardCmd = &cobra.Command{
	Use:    "onboard",
	Short:  "Onboard current machine to NuNet",
	Long:   "",
	PreRun: isDMSRunning(),
	Run: func(cmd *cobra.Command, args []string) {
		memory, _ := cmd.Flags().GetInt64("memory")
		cpu, _ := cmd.Flags().GetInt64("cpu")
		channel, _ := cmd.Flags().GetString("nunet-channel")
		address, _ := cmd.Flags().GetString("address")
		local, _ := cmd.Flags().GetBool("local-enable")
		cardano, _ := cmd.Flags().GetBool("cardano")

		if memory == 0 || cpu == 0 || channel == "" || address == "" {
			fmt.Println(`Error: missing at least one required flag

For onboarding, check:
    nunet onboard --help`)
			os.Exit(1)
		}

		fmt.Println("Checking onboard status...")

		onboarded, err := utils.IsOnboarded()
		if err != nil {
			fmt.Println("Error checking onboard status:", err)
			os.Exit(1)
		}

		err = promptOnboard(onboarded)
		if err != nil {
			fmt.Println("Exiting:", err)
			os.Exit(0)
		}

		onboardJson, err := setOnboardData(memory, cpu, channel, address, cardano, local)
		if err != nil {
			fmt.Println("Error setting onboard data:", err)
			os.Exit(1)
		}

		body, err := utils.ResponseBody(nil, "POST", "/api/v1/onboarding/onboard", onboardJson)
		if err != nil {
			fmt.Println("Error getting response body:", err)
			os.Exit(1)
		}

		errMsg, err := jsonparser.GetString(body, "error")
		if err == nil { // if error message IS found
			fmt.Println("Error:", errMsg)
			os.Exit(1)
		} else if err == jsonparser.KeyPathNotFoundError { // if NO error message is found
			fmt.Println("Sucessfully onboarded!")
		} else { // another error
			fmt.Println("Error parsing response body:", err)
			os.Exit(1)
		}

		os.Exit(0)
	},
}

func promptOnboard(onboarded bool) error {
	if onboarded {
		promptResp := utils.PromptYesNo("Looks like your machine is already onboarded. Do you want to reonboard it? (y/N)")
		if promptResp {
			fmt.Println("Proceeding with reonboard...")
			return nil
		} else {
			return fmt.Errorf("reonboard process aborted by the user")
		}
	} else {
		fmt.Println("Proceeding with onboard...")
	}

	return nil
}

func setOnboardData(memory int64, cpu int64, channel, address string, cardano, serverMode bool) (data []byte, err error) {
	reserved := models.CapacityForNunet{
		Memory:         memory,
		CPU:            cpu,
		Channel:        channel,
		PaymentAddress: address,
		Cardano:        cardano,
		ServerMode:     serverMode,
	}

	data, err = json.Marshal(reserved)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal JSON data: %v", err)
	}

	return data, nil
}
