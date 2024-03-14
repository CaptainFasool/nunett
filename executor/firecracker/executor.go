package firecracker

import (
	"context"
	"fmt"

	"github.com/bacalhau-project/bacalhau/pkg/executor"
	"github.com/bacalhau-project/bacalhau/pkg/models"
	"github.com/bacalhau-project/bacalhau/pkg/storage"
	"github.com/bacalhau-project/bacalhau/pkg/util/generic"
	fc "github.com/firecracker-microvm/firecracker-go-sdk"
	fcclientmodels "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/rs/zerolog/log"
	"go.uber.org/atomic"

	fcmodels "bacalhau_firecracker/pkg/executor/firecracker/models"
	"bacalhau_firecracker/pkg/firecracker"
)

const (
	labelExecutorName = "bacalhau-executor"
	labelJobName      = "bacalhau-jobID"
	labelExecutionID  = "bacalhau-executionID"

	// DefaultCpuCount is the default number of CPUs for a Firecracker VM.
	DefaultCpuCount int64 = 1
	// DefaultMemSize is the default memory size in MiB for a Firecracker VM.
	DefaultMemSize  int64 = 50
)

// Executor is a Firecracker executor that implements the executor.Executor interface.
type Executor struct {
	ID string

	// handlers is a map of executionID to its handler.
	handlers generic.SyncMap[string, *executionHandler]

	activeFlags map[string]chan struct{}
	complete    map[string]chan struct{}
	client      *firecracker.Client
}

// NewExecutor creates a new Firecracker executor.
func NewExecutor(
	_ context.Context,
	id string,
) (*Executor, error) {
	firecrackerClient, err := firecracker.NewFirecrackerClient()
	if err != nil {
		return nil, err
	}
	fe := &Executor{
		ID:          id,
		client:      firecrackerClient,
		activeFlags: make(map[string]chan struct{}),
		complete:    make(map[string]chan struct{}),
	}

	return fe, nil
}

// IsInstalled checks if firecracker is installed on the system.
func (e *Executor) IsInstalled(ctx context.Context) (bool, error) {
	return e.client.IsInstalled(ctx), nil
}


// Start initiates an execution based on the provided RunCommandRequest.
func (e *Executor) Start(ctx context.Context, request *executor.RunCommandRequest) error {
	log.Ctx(ctx).Info().
		Str("executionID", request.ExecutionID).
		Str("jobID", request.JobID).
		Msg("starting execution")

	params := &FirecrackerVMParams{
		ExecutionID:   request.ExecutionID,
		JobID:         request.JobID,
		EngineSpec:    request.EngineParams,
		NetworkConfig: request.Network,
		Resources:     request.Resources,
		Inputs:        request.Inputs,
		Outputs:       request.Outputs,
		ResultsDir:    request.ResultsDir,
	}

	// It's possible that this is being called due to a restart. Whilst we check the handlers to see
	// if we already have a running execution, this map will be empty on a compute node restart. As
	// a result we need to explicitly ask docker if there is a running VM with the relevant
	// bacalhau execution id socket path _before_ we do anything else. If we are able to find one then we
	// will use that VM in the executionHandler that we create.
	machine, err := e.FindRunningVM(ctx, params)
	if err != nil {
		// Unable to find a running container for this execution, we will instead check for a handler, and
		// failing that will create a new container.
		if handler, found := e.handlers.Get(request.ExecutionID); found {
			if handler.active() {
				return fmt.Errorf(
					"starting execution (%s): %w",
					request.ExecutionID,
					executor.ErrAlreadyStarted,
				)
			} else {
				return fmt.Errorf(
					"starting execution (%s): %w",
					request.ExecutionID,
					executor.ErrAlreadyComplete,
				)
			}
		}

		machine, err = e.newFirecrackerVM(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to create firecracker job VM: %w", err)
		}
	}

	handler := &executionHandler{
		client: e.client,
		logger: log.With().
			Str("socket path", machine.Cfg.SocketPath).
			Str("execution", request.ExecutionID).
			Str("job", request.JobID).
			Logger(),
		ID:          e.ID,
		executionID: request.ExecutionID,
		machine:     machine,
		resultsDir:  request.ResultsDir,
		limits:      request.OutputLimits,
		waitCh:      make(chan bool),
		activeCh:    make(chan bool),
		running:     atomic.NewBool(false),
	}

	// register the handler for this executionID
	e.handlers.Put(request.ExecutionID, handler)
	// run the VM.
	go handler.run(ctx)
	return nil
}

// Wait initiates a wait for the completion of a specific execution using its
// executionID. The function returns two channels: one for the result and another
// for any potential error. If the executionID is not found, an error is immediately
// sent to the error channel. Otherwise, an internal goroutine (doWait) is spawned
// to handle the asynchronous waiting. Callers should use the two returned channels
// to wait for the result of the execution or an error. This can be due to issues
// either beginning the wait or in getting the response. This approach allows the
// caller to synchronize Wait with calls to Start, waiting for the execution to complete.
func (e *Executor) Wait(
	ctx context.Context,
	executionID string,
) (<-chan *models.RunCommandResult, <-chan error) {
	handler, found := e.handlers.Get(executionID)
	resultCh := make(chan *models.RunCommandResult, 1)
	errCh := make(chan error, 1)

	if !found {
		errCh <- fmt.Errorf("waiting on execution (%s): %w", executionID, executor.ErrNotFound)
		return resultCh, errCh
	}

	go e.doWait(ctx, resultCh, errCh, handler)
	return resultCh, errCh
}

// doWait is a helper function that actively waits for an execution to finish. It
// listens on the executionHandler's wait channel for completion signals. Once the
// signal is received, the result is sent to the provided output channel. If there's
// a cancellation request (context is done) before completion, an error is relayed to
// the error channel. If the execution result is nil, an error suggests a potential
// flaw in the executor logic.
func (e *Executor) doWait(
	ctx context.Context,
	out chan *models.RunCommandResult,
	errCh chan error,
	handler *executionHandler,
) {
	log.Info().Str("executionID", handler.executionID).Msg("waiting on execution")
	defer close(out)
	defer close(errCh)

	select {
	case <-ctx.Done():
		errCh <- ctx.Err() // Send the cancellation error to the error channel
		return
	case <-handler.waitCh:
		if handler.result != nil {
			log.Info().Str("executionID", handler.executionID).Msg("received results from execution")
			out <- handler.result
		} else {
			errCh <- fmt.Errorf("execution (%s) result is nil", handler.executionID)
		}
	}
}

// Cancel tries to cancel a specific execution by its executionID.
// It returns an error if the execution is not found.
func (e *Executor) Cancel(ctx context.Context, executionID string) error {
	handler, found := e.handlers.Get(executionID)
	if !found {
		return fmt.Errorf("canceling execution (%s): %w", executionID, executor.ErrNotFound)
	}
	return handler.kill(ctx)
}

// Run initiates and waits for the completion of an execution in one call.
// This method serves as a higher-level convenience function that
// internally calls Start and Wait methods.
// It returns the result of the execution or an error if either starting
// or waiting fails, or if the context is canceled.
func (e *Executor) Run(
	ctx context.Context,
	request *executor.RunCommandRequest,
) (*models.RunCommandResult, error) {
	if err := e.Start(ctx, request); err != nil {
		return nil, err
	}
	resCh, errCh := e.Wait(ctx, request.ExecutionID)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case out := <-resCh:
		return out, nil
	case err := <-errCh:
		return nil, err
	}
}

// FirecrackerVMParams is a struct that holds the parameters required to create a Firecracker VM.
type FirecrackerVMParams struct {
	ExecutionID   string
	JobID         string
	EngineSpec    *models.SpecConfig
	NetworkConfig *models.NetworkConfig
	Resources     *models.Resources
	Inputs        []storage.PreparedStorage
	Outputs       []*models.ResultPath
	ResultsDir    string
}

// newFirecrackerVM is an internal method called by Start to set up a new firecracker VM
// for the job execution. It configures the VM based on the provided FirecrackerVMParams.
// This includes decoding engine specifications, setting up machine configuration.
// It then creates the VM but does not start it.
// The method returns a firecracker Machine and an error if any part of the setup fails.
func (e *Executor) newFirecrackerVM(
	ctx context.Context,
	params *FirecrackerVMParams,
) (*fc.Machine, error) {
	fcConfig, err := e.fcConfig(params)
	if err != nil {
		return &fc.Machine{}, fmt.Errorf("decoding engine spec: %w", err)
	}

	machine, err := e.client.NewMachine(ctx, fcConfig)
	if err == nil {
		fcArgs, err := fcmodels.DecodeSpec(params.EngineSpec)
		if err != nil {
			return machine, fmt.Errorf("decoding engine spec: %w", err)
		}
		e.client.VMPassMMDs(ctx, machine, fcArgs.MMDSMessage)
	}

	return machine, nil
}

// FindRunningVM is an internal method that checks if a VM is already running for a given
// executionID. It uses the provided FirecrackerVMParams to decode the engine specifications
// and then searches for a running VM with the same socket path. If found, it returns the
// running machine. If not found, it returns an empty machine and an error.
func (e *Executor) FindRunningVM(
	ctx context.Context,
	params *FirecrackerVMParams,
) (*fc.Machine, error) {
	cfg, err := e.fcConfig(params)
	if err != nil {
		return &fc.Machine{}, fmt.Errorf("decoding engine spec: %w", err)
	}
	return e.client.FindRunningVM(ctx, cfg)
}

// fcConfig is an internal method that decodes the engine specifications from the provided
// FirecrackerVMParams and returns a firecracker Config object. It also sets the machine
// configuration based on the provided resources. If the resources are not provided, it uses
// the default values for memory and CPU count.
func (e *Executor) fcConfig(params *FirecrackerVMParams) (fc.Config, error) {
	fcArgs, err := fcmodels.DecodeSpec(params.EngineSpec)
	if err != nil {
		return fc.Config{}, fmt.Errorf("decoding engine spec: %w", err)
	}

	memSize := DefaultMemSize
	cpuCount := DefaultCpuCount
	if params.Resources != nil {
		if params.Resources.Memory > 0 {
			memSize = int64(params.Resources.Memory)
		}
		if params.Resources.CPU > 0 {
			cpuCount = int64(params.Resources.CPU)
		}
	}

	fcConfig := fc.Config{
		LogFifo:         "./log.fifo",
		VMID:            params.ExecutionID,
		SocketPath:      e.generateSocketPath(params.ExecutionID),
		KernelImagePath: fcArgs.KernelImage,
		InitrdPath:      fcArgs.Initrd,
		KernelArgs:      fcArgs.KernelArgs,
		MachineCfg: fcclientmodels.MachineConfiguration{
			MemSizeMib: fc.Int64(memSize),
			VcpuCount:  fc.Int64(cpuCount),
		},
	}

	drives := fc.NewDrivesBuilder(fcArgs.RootFileSystem)

	for _, input := range params.Inputs {
		drives.AddDrive(input.Volume.Source, input.Volume.ReadOnly)
	}

	fcConfig.Drives = drives.Build()

	return fcConfig, nil
}

// generateSocketPath generates a unique socket path for a given executionID.
func (e *Executor) generateSocketPath(executionID string) string {
	return fmt.Sprintf("/tmp/%s_%s.sock", e.ID, executionID)
}

// labelExecutionValue generates a unique label for a given executionID and executorID.
func labelExecutionValue(executorID string, executionID string) string {
	return executorID + executionID
}

