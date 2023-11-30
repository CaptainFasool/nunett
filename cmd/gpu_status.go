package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	library "gitlab.com/nunet/device-management-service/lib"
)

func NewGPUStatusCmd(librarier Librarier, executer Executer, nvman NVMLManager, gpucontrol map[string]GPUController, gpuinfo map[string]GPU) *cobra.Command {
	return &cobra.Command{
		Use:    "status",
		Short:  "Check GPU status in real time",
		PreRun: isDMSRunning(),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				ctx    context.Context
				cancel context.CancelFunc

				hasAMD, hasNVIDIA bool
				vendors           []library.GPUVendor

				countAMD, countNvidia     int
				amdDevices, nvidiaDevices []GPU

				nvidiaController, amdController GPUController
				ok                              bool

				err error
			)

			ctx, cancel = context.WithCancel(context.Background())

			vendors, err = librarier.DetectGPUVendors()
			if err != nil {
				return fmt.Errorf("unable to detect GPU vendors: %w", err)
			}

			hasAMD = containsVendor(vendors, library.AMD)
			hasNVIDIA = containsVendor(vendors, library.NVIDIA)
			if !hasAMD && !hasNVIDIA {
				return fmt.Errorf("no AMD or NVIDIA GPU(s) detected")
			}

			if hasNVIDIA {
				err = nvman.Init()
				if err != nil {
					return fmt.Errorf("failed to initialize nvml: %w", err)
				}
				defer func() {
					err = nvman.Shutdown()
					if err != nil {
						fmt.Fprintln(cmd.OutOrStderr(), "Error: failed to shutdown nvml:", err)
					}
				}()

				nvidiaController, ok = gpucontrol["nvidia"]
				if !ok {
					return fmt.Errorf("value for key %s not found in gpucontrol map", "nvidia")
				}
				countNvidia, err = nvidiaController.CountDevices()
				if err != nil {
					return fmt.Errorf("failed to count Nvidia devices: %w", err)
				}
			}
			if hasAMD {
				amdController, ok = gpucontrol["amd"]
				if !ok {
					return fmt.Errorf("value for key %s not found in gpucontrol map", "amd")
				}
				countAMD, err = amdController.CountDevices()
				if err != nil {
					return fmt.Errorf("failed to count AMD devices: %w", err)
				}
			}

			for i := 0; i <= countNvidia; i++ {
				device, err := nvidiaController.GetDeviceByIndex(i)
				if err != nil {
					continue
				}
				nvidiaDevices = append(nvidiaDevices, device)
			}
			for i := 0; i <= countAMD; i++ {
				device, err := amdController.GetDeviceByIndex(i)
				if err != nil {
					continue
				}
				amdDevices = append(amdDevices, device)
			}

			// listen for interrupt signal
			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-signalChan
				fmt.Fprintln(cmd.OutOrStderr(), "signal: interrupt")
				cancel()
			}()

			monitorGPUs(ctx, cmd.OutOrStdout(), amdDevices, nvidiaDevices)
			return nil
		},
	}
}

func monitorGPUs(ctx context.Context, w io.Writer, amdDevices, nvidiaDevices []GPU) {
	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(w, "closing...")
			return
		default:
			// clear screen
			fmt.Print(w, "\033[H\033[2J")
			fmt.Fprintln(w, "========== NuNet GPU Status ==========")
			fmt.Fprintln(w, "========== GPU Utilization ==========")
			if len(nvidiaDevices) > 0 {
				printUtilization(w, nvidiaDevices)
			}
			if len(amdDevices) > 0 {
				printUtilization(w, amdDevices)
			}
			fmt.Fprintln(w, "========== Memory Capacity ==========")
			if len(nvidiaDevices) > 0 {
				printMemoryCapacity(w, nvidiaDevices)
			}
			if len(amdDevices) > 0 {
				printMemoryCapacity(w, amdDevices)
			}
			fmt.Fprintln(w, "========== Memory Used ==========")
			if len(nvidiaDevices) > 0 {
				printMemoryUsed(w, nvidiaDevices)
			}
			if len(amdDevices) > 0 {
				printMemoryUsed(w, amdDevices)
			}
			fmt.Fprintln(w, "========== Memory Free ==========")
			if len(nvidiaDevices) > 0 {
				printMemoryFree(w, nvidiaDevices)
			}
			if len(amdDevices) > 0 {
				printMemoryFree(w, amdDevices)
			}
			fmt.Fprintln(w, "========== Temperature ==========")
			if len(nvidiaDevices) > 0 {
				printTemperature(w, nvidiaDevices)
			}
			if len(amdDevices) > 0 {
				printTemperature(w, amdDevices)
			}
			fmt.Fprintln(w, "========== Power Usage ==========")
			if len(nvidiaDevices) > 0 {
				printPowerUsage(w, nvidiaDevices)
			}
			if len(amdDevices) > 0 {
				printPowerUsage(w, amdDevices)
			}
			fmt.Fprintln(w, "")
			fmt.Fprintln(w, "Press CTRL+C to exit...")
			fmt.Fprintln(w, "Refreshing status in a few seconds...")

			time.Sleep(2 * time.Second)
		}
	}
}

func printUtilization(w io.Writer, gpus []GPU) {
	for _, gpu := range gpus {
		fmt.Fprintf(w, "%s: %d%%", gpu.Name(), gpu.UtilizationRate())
	}
}

func printMemoryCapacity(w io.Writer, gpus []GPU) {
	for _, gpu := range gpus {
		fmt.Fprintf(w, "%s: %s\n", gpu.Name(), humanize.IBytes(gpu.Memory().total))
	}
}

func printMemoryUsed(w io.Writer, gpus []GPU) {
	for _, gpu := range gpus {
		fmt.Fprintf(w, "%s: %s\n", gpu.Name(), humanize.IBytes(gpu.Memory().used))
	}
}

func printMemoryFree(w io.Writer, gpus []GPU) {
	for _, gpu := range gpus {
		fmt.Fprintf(w, "%s: %s\n", gpu.Name(), humanize.IBytes(gpu.Memory().free))
	}
}

func printTemperature(w io.Writer, gpus []GPU) {
	for _, gpu := range gpus {
		fmt.Fprintf(w, "%s: %.0f°C\n", gpu.Name(), gpu.Temperature())
	}
}

func printPowerUsage(w io.Writer, gpus []GPU) {
	for _, gpu := range gpus {
		fmt.Fprintf(w, "%s: %dW\n", gpu.Name(), gpu.PowerUsage())
	}
}
