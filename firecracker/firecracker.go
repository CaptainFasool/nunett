package firecracker

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/firecracker-microvm/firecracker-go-sdk"
)

type Client struct{}

func NewFirecrackerClient() (*Client, error) {
	return &Client{}, nil
}

// IsInstalled checks if firecracker is installed on the system
func (c *Client) IsInstalled(ctx context.Context) bool {
	_, err := exec.LookPath("firecracker")
	return err == nil
}

// NewMachine creates a new firecracker machine
func (c *Client) NewMachine(
	ctx context.Context,
	cfg firecracker.Config,
) (*firecracker.Machine, error) {
	// stdout will be directed to this file
	stdoutPath := "/tmp/stdout.log"
	stdout, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to create stdout file: %v", err))
	}

	// stderr will be directed to this file
	stderrPath := "/tmp/stderr.log"
	stderr, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to create stderr file: %v", err))
	}
	cmd := firecracker.VMCommandBuilder{}.
		WithSocketPath(cfg.SocketPath).
		WithStdout(stdout).
		WithStderr(stderr).
		Build(ctx)

	machineOpts := []firecracker.Opt{
		firecracker.WithProcessRunner(cmd),
	}

	m, err := firecracker.NewMachine(ctx, cfg, machineOpts...)
	return m, err
}

// VMStart starts the firecracker machine
func (c *Client) VMStart(ctx context.Context, machine *firecracker.Machine) error {
	err := machine.Start(ctx)
	return err
}

// VMPause pauses the firecracker machine
func (c *Client) VMPause(ctx context.Context, machine *firecracker.Machine) error {
	err := machine.PauseVM(ctx)
	return err
}

// VMResume resumes the firecracker machine
func (c *Client) VMResume(ctx context.Context, machine *firecracker.Machine) error {
	err := machine.ResumeVM(ctx)
	return err
}

// VMShutdown shuts down the firecracker machine
func (c *Client) VMShutdown(ctx context.Context, machine *firecracker.Machine) error {
	err := machine.Shutdown(ctx)
	return err
}

// VMPassMMDs sets the metadata of the firecracker machine
func (c *Client) VMPassMMDs(
	ctx context.Context,
	machine *firecracker.Machine,
	mmds interface{},
) error {
	err := machine.UpdateMetadata(ctx, mmds)
	return err
}

// VMGetMMDs gets the metadata of the firecracker machine
func (c *Client) VMGetMMDs(
	ctx context.Context,
	machine *firecracker.Machine,
) (map[string]string, error) {
	mmds := map[string]string{}
	err := machine.GetMetadata(ctx, mmds)
	return mmds, err
}

// FindRunningVM tries to find a running VM from firecracker config
func (e *Client) FindRunningVM(
	ctx context.Context,
	cfg firecracker.Config,
) (*firecracker.Machine, error) {
	if _, err := os.Stat(cfg.SocketPath); err != nil {
		return &firecracker.Machine{}, fmt.Errorf(
			"VM with socket path %v not found",
			cfg.SocketPath,
		)
	}

	cmd := firecracker.VMCommandBuilder{}.WithSocketPath(cfg.SocketPath).Build(ctx)
	machine, err := firecracker.NewMachine(ctx, cfg, firecracker.WithProcessRunner(cmd))
	if err != nil {
		return &firecracker.Machine{}, fmt.Errorf(
			"Failed to recreate machine with socket %s",
			cfg.SocketPath,
		)
	}

	info, err := machine.DescribeInstanceInfo(ctx)
	if err != nil {
		fmt.Println("Failed to get instance info", err)
	}
	fmt.Println(*info.ID, *info.State)

	return machine, nil
}
