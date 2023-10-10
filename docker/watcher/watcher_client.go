/*
Package watcher contains the logic for a "watcher" client that monitors DMS.

The watcher client works as follows:

1. When started, the watcher client establishes a connection to the DMS over a local TCP socket on port 9898.
2. The DMS periodically sends heartbeats over this connection to indicate that it's still running.
3. The watcher client continuously monitors these heartbeats.
4. If the watcher client does not receive a heartbeat from the DMS within a predefined timeout (currently set at 20 seconds), it assumes the DMS has crashed or been terminated.
5. Upon detecting the absence of heartbeats, the watcher client initiates a cleanup process:
    a. It connects to the local Docker daemon.
    b. Lists all running Docker containers.
    c. Identifies containers associated with the DMS based on a naming convention (containers with names that start with "DMS_").
    d. Stops and removes these identified containers.
6. After cleanup, the watcher client terminates itself.

This mechanism ensures that if the DMS crashes or is terminated unexpectedly, any Docker containers it started are also stopped and removed, preventing orphaned containers from consuming system resources.

Note: The watcher client and DMS communicate over a local TCP socket for simplicity and efficiency. No complex data is exchanged, only simple heartbeats to signal the DMS is still running.
*/

package watcher

import (
	"context"
	"log"
	"net"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func init() {
	// Create a channel to receive OS signals
	sigs := make(chan os.Signal, 1)

	// Notify the channel for SIGINT
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		// Block until a signal is received.
		<-sigs
		// Do nothing (or you can log that SIGINT is received but ignored)
	}()
}

const (
	heartbeatTimeout = 20 * time.Second
)

var lastHeartbeatReceived time.Time
var heartbeatMutex sync.Mutex

func updateHeartbeatTime(t time.Time) {
	heartbeatMutex.Lock()
	defer heartbeatMutex.Unlock()
	lastHeartbeatReceived = t
}

func WatchForHeartbeats() {
	updateHeartbeatTime(time.Now())

	conn, err := net.Dial("tcp", "localhost"+port)
	if err != nil {
		log.Fatalf("Watcher client error: %s", err)
	}
	defer conn.Close()

	go monitorHeartbeats()

	buffer := make([]byte, 128)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Println("Error reading heartbeat:", err)
			return
		}

		if string(buffer[:n]) == "heartbeat" {
			log.Println("Heartbeat received from server")
			updateHeartbeatTime(time.Now())
		}
	}
}

func monitorHeartbeats() {
	for {
		time.Sleep(heartbeatTimeout / 3) // Check three times within the heartbeat timeout period
		if time.Since(lastHeartbeatReceived) > heartbeatTimeout {
			log.Println("No heartbeat detected. Initiating cleanup...")
			cleanup()
			break
		}
	}
}

func cleanup() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Cannot initialize Docker client: %v", err)
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		log.Fatalf("Cannot list containers: %v", err)
	}

	for _, container := range containers {
		log.Printf("Inspecting container with name: %s", container.Names[0]) // Debugging line
		if matchesNamingConvention(container.Names[0]) {
			log.Printf("Attempting to stop container: %s", container.Names[0]) // Debugging line
			// Stop the container
			if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
				log.Printf("Failed to stop container %s: %v", container.ID, err)
			} else {
				log.Printf("Successfully stopped container %s", container.ID) // Debugging line
			}

			log.Printf("Attempting to remove container: %s", container.Names[0]) // Debugging line
			// Remove the container
			if err := cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
				log.Printf("Failed to remove container %s: %v", container.ID, err)
			} else {
				log.Printf("Successfully removed container %s", container.ID) // Debugging line
			}
		}
	}

	log.Println("Cleanup completed. Exiting watcher client.")
	exit()
}

// This function checks if a given container name matches our naming convention.
func matchesNamingConvention(name string) bool {
	name = strings.TrimPrefix(name, "/")
	return strings.HasPrefix(name, "DMS_")
}

func exit() {
	// Exit the watcher client process
	log.Println("Exiting watcher client.")
	os.Exit(0)
}
