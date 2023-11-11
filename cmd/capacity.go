package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding"
	"gitlab.com/nunet/device-management-service/utils"
)

const (
	bitFullFlag      = 1 << iota // 1
	bitAvailableFlag             // 2
	bitOnboardedFlag             // 4
)

var flagFull, flagAvailable, flagOnboarded bool

func init() {
	capacityCmd.Flags().BoolVarP(&flagFull, "full", "f", false, "display device ")
	capacityCmd.Flags().BoolVarP(&flagAvailable, "available", "a", false, "display amount of resources still available for onboarding")
	capacityCmd.Flags().BoolVarP(&flagOnboarded, "onboarded", "o", false, "display amount of onboarded resources")
}

var capacityCmd = &cobra.Command{
	Use:    "capacity",
	Short:  "Display capacity of device resources",
	Long:   `Retrieve capacity of the machine, onboarded or available amount of resources. `,
	PreRun: isDMSRunning(),
	Run: func(cmd *cobra.Command, args []string) {
		onboarded, _ := cmd.Flags().GetBool("onboarded")
		full, _ := cmd.Flags().GetBool("full")
		available, _ := cmd.Flags().GetBool("available")

		// flagCombination stores the bitwise value of combined flags
		var flagCombination int
		if full {
			flagCombination |= bitFullFlag
		}
		if available {
			flagCombination |= bitAvailableFlag
		}
		if onboarded {
			flagCombination |= bitOnboardedFlag
		}

		if flagCombination == 0 {
			fmt.Println(`Error: no flags specified

For more help, check:
    nunet capacity --help`)
			os.Exit(1)
		}

		var table *tablewriter.Table
		// setups table in case of full, available or onboarded flags are passed
		if flagCombination&bitFullFlag != 0 || flagCombination&bitAvailableFlag != 0 || flagCombination&bitOnboardedFlag != 0 {
			table = setupTable()
		}

		if flagCombination&bitFullFlag != 0 {
			handleFull(table)
		}
		if flagCombination&bitAvailableFlag != 0 {
			handleAvailable(table)
		}
		if flagCombination&bitOnboardedFlag != 0 {
			handleOnboarded(table)
		}

		if table != nil {
			table.Render()
		}

	},
}

func setFullData(provisioned *models.Provisioned) []string {
	return []string{
		"Full",
		fmt.Sprintf("%d", provisioned.Memory),
		fmt.Sprintf("%.0f", provisioned.CPU),
		fmt.Sprintf("%d", provisioned.NumCores),
	}
}

func setAvailableData(metadata *models.MetadataV2) []string {
	return []string{
		"Available",
		fmt.Sprintf("%d", metadata.Available.Memory),
		fmt.Sprintf("%d", metadata.Available.CPU),
		"",
	}
}

func setOnboardedData(metadata *models.MetadataV2) []string {
	return []string{
		"Onboarded",
		fmt.Sprintf("%d", metadata.Reserved.Memory),
		fmt.Sprintf("%d", metadata.Reserved.CPU),
		"",
	}
}

func setupTable() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	headers := []string{"Resources", "Memory", "CPU", "Cores"}
	table.SetHeader(headers)
	table.SetAutoMergeCellsByColumnIndex([]int{0})
	table.SetAutoFormatHeaders(false)

	return table
}

func handleFull(table *tablewriter.Table) {
	totalProvisioned := onboarding.GetTotalProvisioned()

	fullData := setFullData(totalProvisioned)
	table.Append(fullData)

}

func handleAvailable(table *tablewriter.Table) {
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

	metadata, err := utils.ReadMetadataFile()
	if err != nil {
		fmt.Println("Error reading metadata file:", err)
		os.Exit(1)
	}

	availableData := setAvailableData(metadata)
	table.Append(availableData)
}

func handleOnboarded(table *tablewriter.Table) {
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

	metadata, err := utils.ReadMetadataFile()
	if err != nil {
		fmt.Println("Error reading metadata file:", err)
		os.Exit(1)
	}

	onboardedData := setOnboardedData(metadata)
	table.Append(onboardedData)
}
