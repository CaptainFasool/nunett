package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	library "gitlab.com/nunet/device-management-service/lib"
	"gitlab.com/nunet/device-management-service/utils"
)

var imagesNVIDIA = []string{
	"registry.gitlab.com/nunet/ml-on-gpu/ml-on-gpu-service/develop/tensorflow",
	"registry.gitlab.com/nunet/ml-on-gpu/ml-on-gpu-service/develop/pytorch",
}

var imagesAMD = []string{
	"registry.gitlab.com/nunet/ml-on-gpu/ml-on-gpu-service/develop/tensorflow-amd",
	"registry.gitlab.com/nunet/ml-on-gpu/ml-on-gpu-service/develop/pytorch-amd",
}

var onboardMLCmd = &cobra.Command{
	Use:    "onboard-ml",
	Short:  "Setup for Machine Learning with GPU",
	PreRun: isDMSRunning(),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		wsl, err := utils.CheckWSL()
		if err != nil {
			return fmt.Errorf("failed to check WSL: %w", err)
		}

		vendors, err := library.DetectGPUVendors()
		if err != nil {
			return fmt.Errorf("unable to detect GPU vendors: %w", err)
		}

		// check for GPU vendors
		hasAMD := containsVendor(vendors, library.AMD)
		hasNVIDIA := containsVendor(vendors, library.NVIDIA)

		if !hasAMD && !hasNVIDIA {
			fmt.Println("No AMD or NVIDIA GPU(s) detected...")
			os.Exit(1)
		}

		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return fmt.Errorf("unable to create Docker client: %w", err)
		}

		imageList, err := cli.ImageList(ctx, types.ImageListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list Docker images: %w", err)
		}

		if wsl {
			fmt.Printf("You are running on Windows Subsystem for Linux (WSL)\nMake sure that NVIDIA drivers are set up correctly\n\nWARNING: AMD GPUs are not supported on WSL!\n")
		}

		if hasNVIDIA {
			err = pullMultipleImages(cli, ctx, imageList, imagesNVIDIA)
			if err != nil {
				return fmt.Errorf("failed to pull NVIDIA images: %w", err)
			}
		}

		if hasAMD {
			err = pullMultipleImages(cli, ctx, imageList, imagesAMD)
			if err != nil {
				return fmt.Errorf("failed to pull AMD images: %w", err)
			}
		}

		return nil
	},
}

func pullMultipleImages(cli *client.Client, ctx context.Context, imageList []types.ImageSummary, images []string) error {
	for i := 0; i < len(images); i++ {
		if !imageExists(imageList, images[i]) {
			err := pullImage(cli, ctx, images[i])
			if err != nil {
				return fmt.Errorf("unable to pull image %s: %v", images[i], err)
			}
		} else {
			fmt.Println("Image already pulled: %s", images[i])
		}
	}

	return nil
}
