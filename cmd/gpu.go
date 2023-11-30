package cmd

import "github.com/spf13/cobra"

func init() {
	var (
		gpuControllerMap map[string]GPUController
		gpuMap           map[string]GPU

		gpuCapacityCmd *cobra.Command
		gpuStatusCmd   *cobra.Command
		gpuOnboardCmd  *cobra.Command
	)

	gpuControllerMap = map[string]GPUController{
		"amd":    &AMDController{},
		"nvidia": &NvidiaController{},
	}
	gpuMap = map[string]GPU{
		"amd":    &amdGPU{},
		"nvidia": &nvidiaGPU{},
	}

	gpuCapacityCmd = NewGPUCapacityCmd(librarier, docker)
	gpuStatusCmd = NewGPUStatusCmd(librarier, executer, nvmlManager, gpuControllerMap, gpuMap)
	gpuOnboardCmd = NewGPUOnboardCmd(utility, librarier, executer)

	gpuCmd.AddCommand(gpuCapacityCmd)
	gpuCmd.AddCommand(gpuStatusCmd)
	gpuCmd.AddCommand(gpuOnboardCmd)
}

var gpuCmd = &cobra.Command{
	Use:   "gpu",
	Short: "GPU-related operations",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
