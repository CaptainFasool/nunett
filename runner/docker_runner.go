package runner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
    "github.com/docker/go-connections/nat"
)

// DockerRunner implements the Runner interface for Docker containers.
type DockerRunner struct {
	dockerClient *client.Client
}

// NewDockerRunner creates a new DockerRunner.
func NewDockerRunner() (*DockerRunner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerRunner{dockerClient: cli}, nil
}

func (d *DockerRunner) Capabilities() map[string]interface{} {
    return map[string]interface{}{
        "Networking":    true,
        "VolumeMount":   true,
        "GPU":           false,
    }
}

// Start creates and starts a Docker container based on the provided JobConfig.
func (d *DockerRunner) Start(job JobConfig) (JobStatus, error) {
	ctx := context.Background()

	containerConfig := &container.Config{
		Image: job.Source,
		Env:   convertMapToSlice(job.Environment),
		Cmd:   job.Command,
	}

	hostConfig := &container.HostConfig{}

	// Handle StorageConfig
    for _, binding := range job.Storage.Bindings {
        hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s:%s", binding.Source, binding.Destination, binding.Mode))
    }

    // Handle NetworkConfig
    portMap, err := createPortMap(job.Network.PortBindings)
    if err != nil {
        return JobStatus{}, err
    }
    hostConfig.PortBindings = portMap
    hostConfig.NetworkMode = container.NetworkMode(job.Network.NetworkMode)


	resp, err := d.dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, job.ID)
	if err != nil {
		return JobStatus{}, err
	}

	if err := d.dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return JobStatus{}, err
	}

	return JobStatus{
		ID:     job.ID,
		Status: JobStatusRunning,
	}, nil
}

// Stop stops a Docker container based on the provided job ID.
func (d *DockerRunner) Stop(jobID string) error {
	ctx := context.Background()
	timeout := 30 * time.Second
	return d.dockerClient.ContainerStop(ctx, jobID, &timeout)
}

// Pause pauses a Docker container based on the provided job ID.
func (d *DockerRunner) Pause(jobID string) error {
	ctx := context.Background()
	return d.dockerClient.ContainerPause(ctx, jobID)
}

// Resume unpauses a Docker container based on the provided job ID.
func (d *DockerRunner) Resume(jobID string) error {
	ctx := context.Background()
	return d.dockerClient.ContainerUnpause(ctx, jobID)
}

// HealthCheck checks the health of the Docker daemon.
func (d *DockerRunner) HealthCheck() error {
	ctx := context.Background()
	_, err := d.dockerClient.Ping(ctx)
	return err
}

// Status retrieves the status of a Docker container based on the provided job ID.
func (d *DockerRunner) Status(jobID string) (JobStatus, error) {
	ctx := context.Background()
	containerJSON, err := d.dockerClient.ContainerInspect(ctx, jobID)
	if err != nil {
		return JobStatus{}, err
	}

	// Parse the StartedAt and FinishedAt fields from string to time.Time
	var startTime, completionTime time.Time
	if containerJSON.State.StartedAt != "" {
		startTime, err = time.Parse(time.RFC3339Nano, containerJSON.State.StartedAt)
		if err != nil {
			return JobStatus{}, fmt.Errorf("failed to parse start time: %w", err)
		}
	}

	if containerJSON.State.FinishedAt != "" &&
		containerJSON.State.FinishedAt != "0001-01-01T00:00:00Z" {
		completionTime, err = time.Parse(time.RFC3339Nano, containerJSON.State.FinishedAt)
		if err != nil {
			return JobStatus{}, fmt.Errorf("failed to parse completion time: %w", err)
		}
	}
	return JobStatus{
		ID:             jobID,
		Status:         determineJobStatus(containerJSON.State),
		ExitCode:       containerJSON.State.ExitCode,
		StartTime:      startTime,
		CompletionTime: completionTime,
		ErrorMessage:   containerJSON.State.Error,
	}, nil
}

// StatusStream streams the status of a Docker container based on the provided job ID.
func (d *DockerRunner) StatusStream(jobID string) (JobStatusStream, error) {
	// Implementation for this method will require setting up a mechanism to
	// stream updates from Docker, possibly using Docker events or periodic polling.
	return nil, errors.New("StatusStream method not implemented")
}

// Helper functions

func createPortMap(bindings []PortBinding) (nat.PortMap, error) {
    portMap := nat.PortMap{}
    for _, binding := range bindings {
        containerPort, err := nat.NewPort(binding.Protocol, binding.RunnerPort)
        if err != nil {
            return nil, err
        }
        portMap[containerPort] = []nat.PortBinding{
            {
                HostIP:   "0.0.0.0",
                HostPort: binding.HostPort,
            },
        }
    }
    return portMap, nil
}

func convertMapToSlice(m map[string]string) []string {
	var s []string
	for k, v := range m {
		s = append(s, k+"="+v)
	}
	return s
}

func determineJobStatus(state *types.ContainerState) JobStatusCode {
	if state.Running {
		return JobStatusRunning
	} else if state.Paused {
		return JobStatusPending
	} else if state.Restarting {
		// Restarting could be interpreted differently based on your use case
		return JobStatusPending
	} else if state.Status == "exited" || state.Status == "dead" {
		if state.ExitCode == 0 {
			return JobStatusCompleted
		}
		return JobStatusFailed
	}
	return JobStatusFailed // Default to failed for unknown states
}
