package docker

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync/atomic"
	"time"

	"gitlab.com/nunet/device-management-service/models"
)

const DestroyTimeout = time.Second * 10

// executionHandler manages the lifecycle and execution of a Docker container for a specific job.
type executionHandler struct {
	// provided by the executor
	ID     string
	client *Client // Docker client for container management.

	// meta data about the task
	jobID       string
	executionID string
	containerID string
	resultsDir  string // Directory to store execution results.

	// synchronization
	activeCh chan bool    // Blocks until the container starts running.
	waitCh   chan bool    // BLocks until execution completes or fails.
	running  *atomic.Bool // Indicates if the container is currently running.

	// result of the execution
	result *models.ExecutionResult
}

// active checks if the execution handler's container is running.
func (h *executionHandler) active() bool {
	return h.running.Load()
}

// run starts the container and handles its execution lifecycle.
func (h *executionHandler) run(ctx context.Context) {
	h.running.Store(true)
	defer func() {
		if err := h.destroy(DestroyTimeout); err != nil {
			zlog.Sugar().Warnf("failed to destroy container: %v\n", err)
		}
		h.running.Store(false)
		close(h.waitCh)
	}()

	if err := h.client.StartContainer(ctx, h.containerID); err != nil {
		h.result = models.NewFailedExecutionResult(fmt.Errorf("failed to start container: %v", err))
		return
	}

	close(h.activeCh) // Indicate that the container has started.

	var containerError error
	var containerExitStatusCode int64

	// Wait for the container to finish or for an execution error.
	statusCh, errCh := h.client.WaitContainer(ctx, h.containerID)
	select {
	case status := <-ctx.Done():
		h.result = models.NewFailedExecutionResult(fmt.Errorf("execution cancelled: %v", status))
		return
	case err := <-errCh:
		zlog.Sugar().Errorf("error while waiting for container: %v\n", err)
		h.result = models.NewFailedExecutionResult(
			fmt.Errorf("failed to wait for container: %v", err),
		)
		return
	case exitStatus := <-statusCh:
		containerExitStatusCode = exitStatus.StatusCode
		containerJSON, err := h.client.InspectContainer(ctx, h.containerID)
		if err != nil {
			h.result = &models.ExecutionResult{
				ExitCode: int(containerExitStatusCode),
				ErrorMsg: err.Error(),
			}
			return
		}
		if containerJSON.ContainerJSONBase.State.OOMKilled {
			containerError = errors.New("container was killed due to OOM")
			h.result = &models.ExecutionResult{
				ExitCode: int(containerExitStatusCode),
				ErrorMsg: containerError.Error(),
			}
			return
		}
		if exitStatus.Error != nil {
			containerError = errors.New(exitStatus.Error.Message)
		}
	}

	// Follow container logs to capture stdout and stderr.
	stdoutPipe, stderrPipe, err := h.client.FollowLogs(ctx, h.containerID)
	if err != nil {
		followError := fmt.Errorf("failed to follow container logs: %w", err)
		if containerError != nil {
			h.result = &models.ExecutionResult{
				ExitCode: int(containerExitStatusCode),
				ErrorMsg: fmt.Sprintf(
					"container error: '%s'. logs error: '%s'",
					containerError,
					followError,
				),
			}
		} else {
			h.result = &models.ExecutionResult{
				ExitCode: int(containerExitStatusCode),
				ErrorMsg: followError.Error(),
			}
		}
		return
	}

	// Capture the logs from the stdout and stderr pipes.
	h.result = models.NewExecutionResult(int(containerExitStatusCode))
	h.result.STDOUT, _ = bufio.NewReader(stdoutPipe).ReadString('\x00') // EOF delimiter
	h.result.STDERR, _ = bufio.NewReader(stderrPipe).ReadString('\x00')
}

// kill sends a stop signal to the container.
func (h *executionHandler) kill(ctx context.Context) error {
	return h.client.StopContainer(ctx, h.containerID, DestroyTimeout)
}

// destroy cleans up the container and its associated resources.
func (h *executionHandler) destroy(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// stop the container
	if err := h.kill(ctx); err != nil {
		return fmt.Errorf("failed to kill container (%s): %w", h.containerID, err)
	}

	if err := h.client.RemoveContainer(ctx, h.containerID); err != nil {
		return err
	}

	// Remove related objects like networks or volumes created for this execution.
	return h.client.RemoveObjectsWithLabel(
		ctx,
		labelExecutionID,
		labelExecutionValue(h.ID, h.jobID, h.executionID),
	)
}

func (h *executionHandler) outputStream(
	ctx context.Context,
	request models.LogStreamRequest,
) (io.ReadCloser, error) {
	since := "1" // Default to the start of UNIX time to get all logs.
	if request.Tail {
		since = strconv.FormatInt(time.Now().Unix(), 10)
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-h.activeCh: // Ensure the container is active before attempting to stream logs.
	}
	// Gets the underlying reader, and provides data since the value of the `since` timestamp.
	return h.client.GetOutputStream(ctx, h.containerID, since, request.Follow)
}
