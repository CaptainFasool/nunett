package firecracker

import (
	"context"
	"fmt"

	"bacalhau_firecracker/pkg/firecracker"
	"github.com/bacalhau-project/bacalhau/pkg/executor"
	"github.com/bacalhau-project/bacalhau/pkg/models"
	fc "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/atomic"
)

// executionHandler is a struct that holds the necessary information to manage the execution of a firecracker VM.
type executionHandler struct {
	//
	// provided by the executor
	client *firecracker.Client
	logger zerolog.Logger
	// meta data about the executor
	ID string

	//
	// meta data about the task
	executionID string
	machine     *fc.Machine
	resultsDir  string
	limits      executor.OutputLimits

	// synchronization
	// blocks until the container starts
	activeCh chan bool
	// blocks until the run method returns
	waitCh chan bool
	// true until the run method returns
	running *atomic.Bool

	// results
	result *models.RunCommandResult
}

// run starts the firecracker VM and waits for it to finish.
func (h *executionHandler) run(ctx context.Context) {
	ActiveExecutions.Inc(ctx, attribute.String("executor_id", h.ID))
	h.running.Store(true)

	defer func() {
		h.running.Store(false)
		close(h.waitCh)
		ActiveExecutions.Dec(ctx, attribute.String("executor_id", h.ID))
	}()

	// start the VM
	h.logger.Info().Msg("starting firecracker execution")
	if err := h.client.VMStart(ctx, h.machine); err != nil {
		startError := errors.Wrap(err, "failed to start container")
		h.logger.Warn().Err(startError).Msg(startError.Error())
		h.result = executor.NewFailedResult(startError.Error())
		return
	}

	// The VM is now active
	close(h.activeCh)

	err := h.machine.Wait(ctx)
	if err != nil {
		h.result = executor.NewFailedResult(err.Error())
		if ctx.Err() != nil {
			reason := fmt.Errorf(
				"context canceled while waiting on container status: %w",
				ctx.Err(),
			)
			h.logger.Err(reason).Msg("cancel waiting on container status")
			h.result = executor.NewFailedResult(reason.Error())
			return
		}
		// the docker client failed to begin the wait request or failed to get a response. We are aborting this execution.
		reason := fmt.Errorf(
			"received error response from docker client while waiting on container: %w",
			err,
		)
		h.logger.Warn().Err(reason).Msg("failed while waiting on container status")
		h.result = executor.NewFailedResult(reason.Error())
		return
	}

	h.logger.Info().Msg("container execution ended")
	h.result = &models.RunCommandResult{
		ExitCode: 0,
	}
}

// kill stops the firecracker VM.
func (h *executionHandler) kill(ctx context.Context) error {
	h.logger.Info().Msg("killing the VM")
	return h.client.VMShutdown(ctx, h.machine)
}

// active returns true if the firecracker VM is running.
func (h *executionHandler) active() bool {
	return h.running.Load()
}
