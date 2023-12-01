package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	library "gitlab.com/nunet/device-management-service/lib"
	"gitlab.com/nunet/device-management-service/utils"
)

const (
	containerPath    = "maint-scripts/install_container_runtime"
	amdDriverPath    = "maint-scripts/install_amd_drivers"
	nvidiaDriverPath = "maint-scripts/install_nvidia_drivers"
	osFile           = "etc/os-release"
)

func NewGPUOnboardCmd(util Utility, librarier Librarier, executer Executer, fs FileSystem) *cobra.Command {
	return &cobra.Command{
		Use:    "onboard",
		Short:  "Install GPU drivers and Container Runtime",
		PreRun: isDMSRunning(),
		RunE: func(cmd *cobra.Command, args []string) error {
			wsl, err := util.CheckWSL()
			if err != nil {
				return fmt.Errorf("failed to check WSL: %w", err)
			}

			vendors, err := librarier.DetectGPUVendors()
			if err != nil {
				return fmt.Errorf("unable to detect GPU vendors: %w", err)
			}

			hasAMD := containsVendor(vendors, library.AMD)
			hasNVIDIA := containsVendor(vendors, library.NVIDIA)

			if !hasAMD && !hasNVIDIA {
				return fmt.Errorf("no AMD or NVIDIA GPU(s) detected...")
			}

			if wsl {
				fmt.Fprintf(cmd.OutOrStdout(), "You are running on Windows Subsystem for Linux (WSL)\n\nWARNING: AMD GPUs are not supported!\n")

				if hasNVIDIA {
					err := promptContainer(executer, containerPath, cmd.OutOrStdout())
					if err != nil {
						return fmt.Errorf("container runtime installation failed: %w", err)
					}
				} else {
					return fmt.Errorf("no NVIDIA GPU(s) detected...")
				}
			} else {
				mining, err := checkMiningOS(fs)
				if err != nil {
					return fmt.Errorf("could not check Mining OS: %w", err)
				}

				if mining {
					fmt.Fprintln(cmd.OutOrStdout(), "You are likely running a Mining OS. Skipping driver installation...")

					err := promptContainer(executer, containerPath, cmd.OutOrStdout())
					if err != nil {
						return fmt.Errorf("container runtime installation failed: %w", err)
					}

					return nil
				}

				if hasNVIDIA {
					NVIDIAGPUs, err := librarier.GetNVIDIAGPUInfo()
					if err != nil {
						return fmt.Errorf("could not fetch NVIDIA info: %w", err)
					}

					printGPUs(cmd.OutOrStdout(), NVIDIAGPUs)

					err = promptContainer(executer, containerPath, cmd.OutOrStdout())
					if err != nil {
						return fmt.Errorf("container runtime installation failed: %w", err)
					}

					err = promptDriverInstallation(executer, cmd.OutOrStdout(), library.NVIDIA, nvidiaDriverPath)
					if err != nil {
						return fmt.Errorf("NVIDIA drivers installation failed: %w", err)
					}
				}

				if hasAMD {
					AMDGPUs, err := librarier.GetAMDGPUInfo()
					if err != nil {
						return fmt.Errorf("failed to fetch AMD info: %w", err)
					}

					printGPUs(cmd.OutOrStdout(), AMDGPUs)

					err = promptDriverInstallation(executer, cmd.OutOrStdout(), library.AMD, amdDriverPath)
					if err != nil {
						return fmt.Errorf("AMD drivers installation failed: %w", err)
					}
				}
			}

			return nil
		},
	}
}

// containsVendor takes a slice of GPUVendor structs that were detected in the system
// and look for a specific vendor, returning true if it is found.
func containsVendor(vendors []library.GPUVendor, target library.GPUVendor) bool {
	for _, v := range vendors {
		if v == target {
			return true
		}
	}

	return false
}

// runScript executes a bash script from a given path.
// It takes the script's path as input and tries to run it, if successfull it prints the output.
func runScript(executer Executer, scriptPath string, w io.Writer) error {
	script := executer.Execute("/bin/bash", scriptPath)

	output, err := script.CombinedOutput()
	if err != nil {
		return fmt.Errorf("script failed with error: %w", err)
	}

	fmt.Fprintf(w, "%s\n", output)

	return nil
}

// promptContainer takes container runtime script path as input and prompts the user for confirmation.
// If confirmed, it runs the script.
func promptContainer(executer Executer, containerPath string, w io.Writer) error {
	proceed, err := utils.PromptYesNo("Do you want to proceed with Container Runtime installation? (y/N)")
	if err != nil {
		return err
	}
	if proceed {
		err := runScript(executer, containerPath, w)
		if err != nil {
			return fmt.Errorf("cannot run container runtime installation script: %v", err)
		}
	}

	return nil
}

// promptDriverInstallation takes GPUVendor (for printing) and the installation script as inputs.
// It prompts the user for confirmation and if confirmed it runs the script.
func promptDriverInstallation(executer Executer, w io.Writer, vendor library.GPUVendor, scriptPath string) error {
	prompt := fmt.Sprintf("Do you want to proceed with %s driver installation? (y/N)", vendor.String())

	proceed, err := utils.PromptYesNo(prompt)
	if err != nil {
		return err
	}

	if proceed {
		err := runScript(executer, scriptPath, w)
		if err != nil {
			return fmt.Errorf("cannot run driver installation script: %v", err)
		}
	}

	return nil
}

// printGPUs display a list of detected GPUs in the machine.
// It takes a slice of GPUInfo structs as input, get the vendor from the first element
// and then iterate over each element to display the GPU card series.
func printGPUs(w io.Writer, gpus []library.GPUInfo) {
	var vendor string

	if len(gpus) == 0 {
		return
	}

	vendor = gpus[0].Vendor.String()
	fmt.Fprintf(w, "Available %s GPU(s):", vendor)
	for _, gpu := range gpus {
		fmt.Fprintf(w, "- %s\n", gpu.GPUName)
	}
}

// checkMiningOS detects if host is running a mining OS.
// It reads from /etc/os-release file and look for common distros inside of it, if any is found it returns true.
func checkMiningOS(fs FileSystem) (bool, error) {
	miningOSes := []string{"Hive", "Rave", "PiMP", "Minerstat", "SimpleMining", "NH", "Miner", "SM", "MMP"}

	info, err := fs.ReadFile(osFile)
	if err != nil {
		return false, fmt.Errorf("cannot read file %s: %w", osFile, err)
	}

	infoStr := string(info)
	for _, os := range miningOSes {
		if strings.Contains(infoStr, os) {
			return true, nil
		}
	}

	return false, nil
}
