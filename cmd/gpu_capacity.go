package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	library "gitlab.com/nunet/device-management-service/lib"
)

// ContainerOptions set parameters for running a Docker container (NVIDIA/AMD)
type ContainerOptions struct {
	UseGPUs    bool
	Devices    []string
	Groups     []string
	Image      string
	Command    []string
	Entrypoint []string
}

var flagCudaTensor, flagRocmHip bool

func init() {
	gpuCapacityCmd.Flags().BoolVarP(&flagCudaTensor, "cuda-tensor", "c", false, "check CUDA Tensor")
	gpuCapacityCmd.Flags().BoolVarP(&flagRocmHip, "rocm-hip", "r", false, "check ROCM-HIP")
}

var gpuCapacityCmd = &cobra.Command{
	Use:    "capacity",
	Short:  "Check availability of NVIDIA/AMD GPUs",
	PreRun: isDMSRunning(),
	RunE: func(cmd *cobra.Command, args []string) error {
		cuda, _ := cmd.Flags().GetBool("cuda-tensor")
		rocm, _ := cmd.Flags().GetBool("rocm-hip")

		if !cuda && !rocm {
			return fmt.Errorf("no flags specified")
		}

		vendors, err := library.DetectGPUVendors()
		if err != nil {
			return fmt.Errorf("could not detect GPU vendors: %w", err)
		}

		hasAMD := containsVendor(vendors, library.AMD)
		hasNVIDIA := containsVendor(vendors, library.NVIDIA)

		if !hasAMD && !hasNVIDIA {
			return fmt.Errorf("no AMD or NVIDIA GPU(s) detected...")
		}

		ctx := context.Background()

		if cuda {
			if !hasNVIDIA {
				return fmt.Errorf("no NVIDIA GPU(s) detected...")
			}

			cudaOpts := ContainerOptions{
				UseGPUs:    true,
				Image:      "registry.gitlab.com/nunet/ml-on-gpu/ml-on-gpu-service/develop/pytorch",
				Command:    []string{"python", "check-cuda-and-tensor-cores-availability.py"},
				Entrypoint: []string{""},
			}

			// initialize client outside of command
			// pass as a field for the struct
			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				return fmt.Errorf("unable to create Docker client: %w", err)
			}

			images, err := cli.ImageList(ctx, types.ImageListOptions{})
			if err != nil {
				return fmt.Errorf("unable to list Docker images: %w", err)
			}

			if !imageExists(images, cudaOpts.Image) {
				err := pullImage(ctx, cli, cudaOpts.Image, cmd.OutOrStdout())
				if err != nil {
					return fmt.Errorf("failed to pull CUDA image %s: %w", cudaOpts.Image, err)
				}
			}

			err = runDockerContainer(ctx, cli, cudaOpts, cmd.OutOrStderr())
			if err != nil {
				return fmt.Errorf("failed to run CUDA container: %w", err)
			}
		}

		if rocm {
			if !hasAMD {
				return fmt.Errorf("no AMD GPU(s) detected...")
			}

			rocmOpts := ContainerOptions{
				Devices:    []string{"/dev/kfd", "/dev/dri"},
				Groups:     []string{"video", "render"},
				Image:      "registry.gitlab.com/nunet/ml-on-gpu/ml-on-gpu-service/develop/pytorch-amd",
				Command:    []string{"python", "check-rocm-and-hip-availability.py"},
				Entrypoint: []string{""},
			}

			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				return fmt.Errorf("failed to create Docker client: %w", err)
			}

			images, err := cli.ImageList(ctx, types.ImageListOptions{})
			if err != nil {
				return fmt.Errorf("failed to list Docker images: %w", err)
			}

			if !imageExists(images, rocmOpts.Image) {
				err := pullImage(ctx, cli, rocmOpts.Image, cmd.OutOrStdout())
				if err != nil {
					return fmt.Errorf("could not pull ROCm-HIP image %s: %w", rocmOpts.Image, err)
				}
			}

			err = runDockerContainer(ctx, cli, rocmOpts, cmd.OutOrStderr())
			if err != nil {
				return fmt.Errorf("failed to run ROCm-HIP container: %w", err)
			}
		}

		return nil
	},
}

func validateOptions(options ContainerOptions) error {
	if options.Image == "" {
		return fmt.Errorf("image name cannot be empty")
	}

	return nil
}

func configureContainer(options ContainerOptions) (*container.Config, *container.HostConfig, error) {
	var (
		config     *container.Config
		hostConfig *container.HostConfig
	)

	config = &container.Config{
		Image:      options.Image,
		Entrypoint: options.Entrypoint,
		Cmd:        options.Command,
		Tty:        true,
	}

	hostConfig = &container.HostConfig{}

	if options.UseGPUs {
		gpuOpts := opts.GpuOpts{}

		err := gpuOpts.Set("all")
		if err != nil {
			return &container.Config{}, &container.HostConfig{}, fmt.Errorf("failed setting GPU opts: %w", err)
		}

		hostConfig.DeviceRequests = gpuOpts.Value()
	}

	for _, device := range options.Devices {
		hostConfig.Devices = append(hostConfig.Devices, container.DeviceMapping{
			PathOnHost:        device,
			PathInContainer:   device,
			CgroupPermissions: "rwm",
		})
	}

	hostConfig.GroupAdd = options.Groups

	return config, hostConfig, nil
}

func createAndStartContainer(ctx context.Context, cli *client.Client, config *container.Config, hostConfig *container.HostConfig, w io.Writer) (string, error) {
	var (
		resp container.ContainerCreateCreatedBody
		err  error
	)

	resp, err = cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("cannot create container: %v", err)
	}

	defer func() {
		err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
		if err != nil {
			fmt.Fprintf(w, "WARNING: could not remove container: %v\n", err)
		}
	}()

	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("cannot start container: %v", err)
	}

	return resp.ID, nil
}

func attachAndMonitorContainer(ctx context.Context, cli *client.Client, id string, w io.Writer) error {
	var (
		out types.HijackedResponse
		err error

		waitCh <-chan container.ContainerWaitOKBody
		errCh  <-chan error
	)

	out, err = cli.ContainerAttach(ctx, id, types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return fmt.Errorf("failed attaching container: %v", err)
	}

	io.Copy(w, out.Reader)

	waitCh, errCh = cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case waitResult := <-waitCh:
		if waitResult.Error != nil {
			return fmt.Errorf("container exit error: %s", waitResult.Error.Message)
		}
	case err := <-errCh:
		return fmt.Errorf("error waiting for container: %v", err)
	}

	return nil
}

func runDockerContainer(ctx context.Context, cli *client.Client, options ContainerOptions, w io.Writer) error {
	var (
		err        error
		id         string
		config     *container.Config
		hostConfig *container.HostConfig
	)

	err = validateOptions(options)
	if err != nil {
		return fmt.Errorf("invalid container option: %w", err)
	}

	config, hostConfig, err = configureContainer(options)
	if err != nil {
		return fmt.Errorf("failed to configure container: %w", err)
	}

	id, err = createAndStartContainer(ctx, cli, config, hostConfig, w)
	if err != nil {
		return fmt.Errorf("failed to create and start container %s: %w", id, err)
	}

	err = attachAndMonitorContainer(ctx, cli, id, w)
	if err != nil {
		return fmt.Errorf("failed to attach and monitor container %s: %w", id, err)
	}

	return nil
}

func imageExists(images []types.ImageSummary, imageName string) bool {
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageName {
				return true
			}
		}
	}
	return false
}

func pullImage(ctx context.Context, cli *client.Client, image string, w io.Writer) error {
	ctxCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	out, err := cli.ImagePull(ctxCancel, image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("unable to pull image %s: %v", image, err)
	}

	// define interrupt to stop image pull
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interrupt
		fmt.Fprintln(w, "signal: interrupt")
		cancel()
	}()

	fmt.Fprint(w, "Pulling image: %s\nThis may take some time...\n", image)
	defer out.Close()

	io.Copy(w, out)

	return nil
}
