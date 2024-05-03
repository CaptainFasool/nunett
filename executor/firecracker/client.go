package firecracker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/firecracker-microvm/firecracker-go-sdk"
)

const pidCheckTickTime = 100 * time.Millisecond

// Client wraps the Firecracker SDK to provide high-level operations on Firecracker VMs.
type Client struct{}

func NewFirecrackerClient() (*Client, error) {
	return &Client{}, nil
}

// IsInstalled checks if Firecracker is installed on the host.
func (c *Client) IsInstalled() bool {
	// LookPath searches for an executable named file in the directories named by the PATH environment variable.
	// There might be a better way to check if Firecracker is installed.
	_, err := exec.LookPath("firecracker")
	return err == nil
}

// CreateVM creates a new Firecracker VM with the specified configuration.
func (c *Client) CreateVM(
	ctx context.Context,
	cfg firecracker.Config,
) (*firecracker.Machine, error) {
	cmd := firecracker.VMCommandBuilder{}.
		WithSocketPath(cfg.SocketPath).
		Build(ctx)

	machineOpts := []firecracker.Opt{
		firecracker.WithProcessRunner(cmd),
	}

	m, err := firecracker.NewMachine(ctx, cfg, machineOpts...)
	return m, err
}

// StartVM starts the Firecracker VM.
func (c *Client) StartVM(ctx context.Context, m *firecracker.Machine) error {
	return m.Start(ctx)
}

// ShutdownVM shuts down the Firecracker VM.
func (c *Client) ShutdownVM(ctx context.Context, m *firecracker.Machine) error {
	return m.Shutdown(ctx)
}

// DestroyVM destroys the Firecracker VM.
func (c *Client) DestroyVM(
	ctx context.Context,
	m *firecracker.Machine,
	timeout time.Duration,
) error {
	// Remove the socket file.
	defer os.Remove(m.Cfg.SocketPath)

	// Get the PID of the Firecracker process and shut down the VM.
	// If the process is still running after the timeout, kill it.
	pid, _ := m.PID()
	c.ShutdownVM(ctx, m)

	// If the process is not running, return early.
	if pid <= 0 {
		return nil
	}

	// This checks if the process is still running every pidCheckTickTime.
	// If the process is still running after the timeout it will set done to false.
	done := make(chan bool, 1)
	go func() {
		ticker := time.NewTicker(pidCheckTickTime)
		defer ticker.Stop()
		to := time.NewTimer(timeout)
		defer to.Stop()
		for {
			select {
			case <-to.C:
				done <- false
				return
			case <-ticker.C:
				if pid, _ := m.PID(); pid <= 0 {
					done <- true
					return
				}
			}
		}
	}()

	// Wait for the check to finish.
	select {
	case killed := <-done:
		if !killed {
			// The shutdown request timed out, kill the process with SIGKILL.
			err := syscall.Kill(int(pid), syscall.SIGKILL)
			if err != nil {
				return fmt.Errorf("failed to kill process: %v", err)
			}
		}
	}
	return nil
}

// FindVM finds a Firecracker VM by its socket path.
// This implementation checks if the VM is running by sending a request to the Firecracker API.
func (c *Client) FindVM(ctx context.Context, socketPath string) (*firecracker.Machine, error) {
	// Check if the socket file exists.
	if _, err := os.Stat(socketPath); err != nil {
		return nil, fmt.Errorf("VM with socket path %v not found", socketPath)
	}

	// Create a new Firecracker machine instance.
	cmd := firecracker.VMCommandBuilder{}.WithSocketPath(socketPath).Build(ctx)
	machine, err := firecracker.NewMachine(
		ctx,
		firecracker.Config{SocketPath: socketPath},
		firecracker.WithProcessRunner(cmd),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create machine with socket %s: %v", socketPath, err)
	}

	// Check if the VM is running by getting its instance info.
	info, err := machine.DescribeInstanceInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance info for socket %s: %v", socketPath, err)
	}

	if *info.State != "Running" {
		return nil, fmt.Errorf(
			"VM with socket %s is not running, current state: %s",
			socketPath,
			*info.State,
		)
	}

	return machine, nil
}
