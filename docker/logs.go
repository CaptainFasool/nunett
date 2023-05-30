package docker

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"gitlab.com/nunet/device-management-service/libp2p"
)

// cleanFlushInfo takes in bytes.Buffer from docker logs output and for each line
// if it has a \r in the lines, takes the last one and compose another string
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

// GetLogs return logs from the container io.ReadCloser. It's the caller duty
// duty to do a stdcopy.StdCopy. Any other method might render unknown
// unicode character as log output has both stdout and stderr. That starting
// has info if that line is stderr or stdout.
func GetLogs(contName string) (io.ReadCloser, error) {
	options := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true}

	out, err := dc.ContainerLogs(ctx, contName, options)
	if err != nil {
		return nil, fmt.Errorf("unable to get logs: %v", err)
	}

	return out, nil
}

// sendLogsToSPD is a facade which handles fetching and sending of chunked
// logs to service provider.
func sendLogsToSPD(containerID string, since string) {
	// Fetch delta of logs from last log fetch.
	stdout, stderr := fetchLogsFromContainer(containerID, since)
	if stdout.Len() == 0 && stderr.Len() == 0 {
		return
	}

	// Send logs to service provider
	if stdout.String() != "" {
		libp2p.DeploymentUpdate(libp2p.MsgLogStdout, stdout.String(), false)
	}
	if stderr.String() != "" {
		libp2p.DeploymentUpdate(libp2p.MsgLogStderr, stderr.String(), false)
	}
}

func fetchLogsFromContainer(containerID string, since string) (stdout, stderr bytes.Buffer) {
	// use go docker api to fetch logs from given containerID
	options := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Since: since}

	out, err := dc.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return bytes.Buffer{}, bytes.Buffer{}
	}

	stdcopy.StdCopy(&stdout, &stderr, out)

	return stdout, stderr
}
