package docker

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"gitlab.com/nunet/device-management-service/libp2p"
)

// cleanFlushInfo takes in bytes.Buffer from docker logs output and for each line
// if it has a \r in the lines, takes the last one and composes another string
// out of that.
func cleanFlushInfo(bytesBuffer *bytes.Buffer) string {
	scanner := bufio.NewScanner(bytesBuffer)
	finalString := ""

	for scanner.Scan() {
		line := scanner.Text()
		chunks := strings.Split(line, "\r")
		lastChunk := chunks[len(chunks)-1] // fetch the last update of the line
		finalString += lastChunk + "\n"
	}

	return finalString
}

// GetLogs returns logs from the container as io.ReadCloser.
func GetLogs(ctx context.Context, contName string) (io.ReadCloser, error) {
	options := types.ContainerLogsOptions{ShowStdout: true}

	out, err := dc.ContainerLogs(ctx, contName, options)
	if err != nil {
		return nil, fmt.Errorf("unable to get logs: %v", err)
	}

	return out, nil
}

// sendLogsToSPD is a facade which handles fetching and sending of chunked
// logs to the service provider.
func sendLogsToSPD(ctx context.Context, containerID string, since string) error {
	// Fetch delta of logs from the last log fetch.
	stdout, err := fetchLogsFromContainer(ctx, containerID, since)
	if err != nil {
		return fmt.Errorf("failed to fetch logs: %v", err)
	}

	if stdout.Len() == 0 {
		return fmt.Errorf("Logs not found to send to SPD: %v", err)
	}

	// Send logs to the service provider
	if stdout.String() != "" {
		libp2p.DeploymentUpdate(libp2p.MsgLogStdout, stdout.String(), false)
	}

	return nil
}

func fetchLogsFromContainer(ctx context.Context, containerID string, since string) (stdout bytes.Buffer, err error) {
	// Use Docker API to fetch logs from the given containerID
	options := types.ContainerLogsOptions{ShowStdout: true, Since: since}

	out, err := dc.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return bytes.Buffer{}, err
	}

	// Assuming TTY mode is enabled: Combined output
	io.Copy(&stdout, out)

	return stdout, nil
}
