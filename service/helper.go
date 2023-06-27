package service

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/client"
)

func createDockerClient() (*client.Client, error) {
	XDG_RUNTIME_DIR := os.Getenv("XDG_RUNTIME_DIR")
	sockPath := fmt.Sprintf("unix://%s/docker.sock", XDG_RUNTIME_DIR)

	// Create a new Docker client with default configuration
	dockerClient, err := client.NewClientWithOpts(
		client.WithHost(sockPath),
	)
	if err != nil {
		return nil, err
	}

	// Set the API version to match the Docker daemon version
	dockerClient.NegotiateAPIVersion(context.Background())

	// Ping Docker to verify the connection
	_, err = dockerClient.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return dockerClient, nil
}
