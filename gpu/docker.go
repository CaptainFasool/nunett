package gpu

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/cpu"
	"gitlab.com/nunet/device-management-service/models"
)

const (
	vcpuToMicroseconds float64 = 100000
)

func PullImage(ctx context.Context, cli *client.Client, imageName string) {
	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer out.Close()
	// io.Copy(os.Stdout, out)
}

func mhzPerCore() float64 {
	cpus, err := cpu.Info()
	if err != nil {
		panic(err)
	}
	return cpus[0].Mhz
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func mhzToVCPU(cpuInMhz int) float64 {
	vcpu := float64(cpuInMhz) / mhzPerCore()
	return toFixed(vcpu, 2)
}

func RunContainer(ctx context.Context, cli *client.Client, depReq models.DeploymentRequest) (contID string) {
	gpuOpts := opts.GpuOpts{}
	gpuOpts.Set("all")

	modelUrl := depReq.Params.ModelURL
	packages := strings.Join(depReq.Params.Packages, " ")
	containerConfig := &container.Config{
		Image: depReq.Params.ImageID,
		Cmd:   []string{modelUrl, packages},
		// Tty:          true,
	}

	memoryMbToBytes := int64(depReq.Constraints.RAM * 1024 * 1024)
	VCPU := mhzToVCPU(depReq.Constraints.CPU)
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			DeviceRequests: gpuOpts.Value(),
			Memory:         memoryMbToBytes,
			CPUQuota:       int64(VCPU * vcpuToMicroseconds),
		},
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")

	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	return resp.ID
}

func GetLogs(ctx context.Context, cli *client.Client, contName string) (logOutput string) {
	options := types.ContainerLogsOptions{ShowStdout: true}

	out, err := cli.ContainerLogs(ctx, contName, options)
	if err != nil {
		panic(err)
	}

	// stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	bytes, _ := io.ReadAll(out)
	return string(bytes)
}

func DeleteContainer(ctx context.Context, cli *client.Client, contName string) {
	options := types.ContainerRemoveOptions{}

	err := cli.ContainerRemove(ctx, contName, options)
	if err != nil {
		panic(err)
	}
}

// HandleDockerDeployment does following docker based actions in the sequence:
// Pull image, run container, get logs, delete container, send log to the requester
func HandleDockerDeployment(depReq models.DeploymentRequest) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Pull the image
	imageName := depReq.Params.ImageID

	PullImage(ctx, cli, imageName)

	// Run the container.
	contID := RunContainer(ctx, cli, depReq)

	// // Get the logs.
	logOutput := GetLogs(ctx, cli, contID)
	fmt.Println(logOutput)

	// Delete the container.
	DeleteContainer(ctx, cli, contID)

	// ToBeImplemented: Send back the logs.
}
