package gpu

import (
	"context"
	"io"

	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

func PullImage(ctx context.Context, cli *client.Client, imageName string) {
	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer out.Close()
	// io.Copy(os.Stdout, out)
}

func RunContainer(ctx context.Context, cli *client.Client, imgName string, cmd []string) (contID string) {
	gpuOpts := opts.GpuOpts{}
	gpuOpts.Set("all")

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "nvidia/cuda:10.0-base",
		Cmd:   cmd,
		// Tty:   false,
	}, &container.HostConfig{Resources: container.Resources{DeviceRequests: gpuOpts.Value()}}, nil, nil, "")

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

func DeleteImage(ctx context.Context, cli *client.Client, imagID string) {
	options := types.ImageRemoveOptions{}

	imgDeleteResp, err := cli.ImageRemove(ctx, imagID, options)
	if err != nil {
		panic(err)
	}

	_ = imgDeleteResp // contains hashes of all the images tags and their child
}

// HandleDockerDeployment does following docker based actions in the sequence:
// Pull image, run container, get logs, delete container, delete image, send log to the requester
func HandleDockerDeployment(c *gin.Context) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Pull the image
	// TODO: Where in the flow do we get the image name.
	imageName := "nvidia/cuda:10.0-base"

	PullImage(ctx, cli, imageName)

	// Run the container.
	// TODO: What command do we run inside container? Do does the Image comes pre-baked?
	contID := RunContainer(ctx, cli, imageName, []string{"nvidia-smi"})

	// Get the logs.
	logOutput := GetLogs(ctx, cli, contID)

	// Delete the container.
	DeleteContainer(ctx, cli, contID)

	// Remove the downloaded image.
	DeleteImage(ctx, cli, imageName)

	// Send back the logs.
	// TODO: Send message to the requesting peer instead of below stub.
	c.JSON(200, logOutput)
}
