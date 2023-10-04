package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
)

var infoCmd = &cobra.Command{
	Use:    "info",
	Short:  "Display information about onboarded device",
	Long:   "Display metadata of onboarded device on Nunet Device Management Service",
	PreRun: isDMSRunning(),
	Run: func(cmd *cobra.Command, args []string) {
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

		printMetadata(metadata)
	},
}

// custom formatter for printing metadata YAML-like
func printMetadata(metadata models.MetadataV2) {
	fmt.Println("metadata:")

	if metadata.Name != "" {
		fmt.Printf("  name: %s\n", metadata.Name)
	}

	if metadata.UpdateTimestamp != 0 {
		fmt.Printf("  update_timestamp: %d\n", metadata.UpdateTimestamp)
	}

	if metadata.Resource.MemoryMax != 0 || metadata.Resource.TotalCore != 0 || metadata.Resource.CPUMax != 0 {
		fmt.Println("  resource:")

		if metadata.Resource.MemoryMax != 0 {
			fmt.Printf("    memory_max: %d\n", metadata.Resource.MemoryMax)
		}

		if metadata.Resource.TotalCore != 0 {
			fmt.Printf("    total_core: %d\n", metadata.Resource.TotalCore)
		}

		if metadata.Resource.CPUMax != 0 {
			fmt.Printf("    cpu_max: %d\n", metadata.Resource.CPUMax)
		}
	}

	if metadata.Available.CPU != 0 || metadata.Available.Memory != 0 {
		fmt.Println("  available:")

		if metadata.Available.CPU != 0 {
			fmt.Printf("    cpu: %d\n", metadata.Available.CPU)
		}

		if metadata.Available.Memory != 0 {
			fmt.Printf("    memory: %d\n", metadata.Available.Memory)
		}
	}

	if metadata.Reserved.CPU != 0 || metadata.Reserved.Memory != 0 {
		fmt.Println("  reserved:")

		if metadata.Reserved.CPU != 0 {
			fmt.Printf("    cpu: %d\n", metadata.Reserved.CPU)
		}

		if metadata.Reserved.Memory != 0 {
			fmt.Printf("    memory: %d\n", metadata.Reserved.Memory)
		}
	}

	if metadata.Network != "" {
		fmt.Printf("  network: %s\n", metadata.Network)
	}

	if metadata.PublicKey != "" {
		fmt.Printf("  public_key: %s\n", metadata.PublicKey)
	}

	if metadata.NodeID != "" {
		fmt.Printf("  node_id: %s\n", metadata.NodeID)
	}

	if metadata.AllowCardano {
		fmt.Printf("  allow_cardano: %v\n", metadata.AllowCardano)
	}

	if len(metadata.GpuInfo) > 0 {
		fmt.Println("  gpu_info:")
		for i, gpu := range metadata.GpuInfo {
			fmt.Printf("    - gpu %d:\n", i+1)
			if gpu.Name != "" {
				fmt.Printf("      name: %s\n", gpu.Name)
			}
			if gpu.TotVram != 0 {
				fmt.Printf("      tot_vram: %d\n", gpu.TotVram)
			}
			if gpu.FreeVram != 0 {
				fmt.Printf("      free_vram: %d\n", gpu.FreeVram)
			}
		}
	}

	if metadata.Dashboard != "" {
		fmt.Printf("  dashboard: %s\n", metadata.Dashboard)
	}
}
