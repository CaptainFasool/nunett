package firecracker

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/firecracker-microvm/firecracker-go-sdk"

	"gitlab.com/nunet/device-management-service/models"
)

// executionHandler is a struct that holds the necessary information to manage the execution of a firecracker VM.
type executionHandler struct {
	//
	// provided by the executor
	ID     string
	client *Client

	// meta data about the task
	JobID       string
	executionID string
	machine     *firecracker.Machine
	resultsDir  string

	// synchronization
	// synchronization
	activeCh chan bool    // Blocks until the container starts running.
	waitCh   chan bool    // BLocks until execution completes or fails.
	running  *atomic.Bool // Indicates if the container is currently running.

	// result of the execution
	result *models.ExecutionResult
}

// active returns true if the firecracker VM is running.
func (h *executionHandler) active() bool {
	return h.running.Load()
}

// run starts the firecracker VM and waits for it to finish.
func (h *executionHandler) run(ctx context.Context) {
	h.running.Store(true)

	defer func() {
		destroyTimeout := time.Second * 10
		if err := h.destroy(destroyTimeout); err != nil {
			zlog.Sugar().Warnf("failed to destroy container: %v\n", err)
		}
		h.running.Store(false)
		close(h.waitCh)
	}()

	// start the VM
	zlog.Sugar().Info("starting firecracker execution")
	if err := h.client.StartVM(ctx, h.machine); err != nil {
		h.result = models.NewFailedExecutionResult(fmt.Errorf("failed to start VM: %v", err))
		return
	}

	close(h.activeCh) // Indicate that the VM has started.

	err := h.machine.Wait(ctx)
	if err != nil {
		if ctx.Err() != nil {
			h.result = models.NewFailedExecutionResult(
				fmt.Errorf("context closed while waiting on VM: %v", err),
			)
			return
		}
		h.result = models.NewFailedExecutionResult(fmt.Errorf("failed to wait on VM: %v", err))
		return
	}

	h.result = models.NewExecutionResult(models.ExecutionStatusCodeSuccess)
}

// kill stops the firecracker VM.
func (h *executionHandler) kill(ctx context.Context) error {
	return h.client.ShutdownVM(ctx, h.machine)
}

// destroy stops the firecracker VM and removes its resources.
func (h *executionHandler) destroy(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return h.client.DestroyVM(ctx, h.machine, timeout)
}
