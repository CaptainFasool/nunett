package gpu

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// ListContainers(ctx, cli)
	// ListImages(ctx, cli)
	// PullImage(ctx, cli)
	// PrintLogs(ctx, cli)
	RunContainer(ctx, cli)
}

func RunContainer(ctx context.Context, cli *client.Client) {
	reader, err := cli.ImagePull(ctx, "hello-world", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer reader.Close()
	io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   []string{"echo", "hello world"},
		Tty:   false,
	}, nil, nil, nil, "")

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

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

func PullImage(ctx context.Context, cli *client.Client) {
	out, err := cli.ImagePull(ctx, "mongo", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer out.Close()

	io.Copy(os.Stdout, out)
}

func PrintLogs(ctx context.Context, cli *client.Client) {
	options := types.ContainerLogsOptions{ShowStdout: true}

	out, err := cli.ContainerLogs(ctx, "nunet-adapter-kubuntu-20595d13-67cb-2af9-e5da-a4e2520af544", options)
	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, out)
}
