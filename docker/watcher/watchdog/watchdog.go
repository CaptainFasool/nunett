package main

import (
	"context"
	"log"
	"net"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	heartbeatTimeout = 20 * time.Second
	port             = ":9898"
)

func main() {
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error starting watchdog: %s", err)
	}
	defer ln.Close()

	heartbeatChan := make(chan bool)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println("Error accepting connection:", err)
				continue
			}
			go handleConnection(conn, heartbeatChan)
		}
	}()

	for {
		select {
		case <-heartbeatChan:
			// Reset the timer whenever a heartbeat is received
			continue
		case <-time.After(heartbeatTimeout):
			killDMSContainers()
			log.Println("Exiting watchdog after taking action.")
			return
		}
	}
}

func handleConnection(c net.Conn, heartbeatChan chan bool) {
	buffer := make([]byte, 1024)
	for {
		_, err := c.Read(buffer)
		if err != nil {
			c.Close()
			return
		}
		heartbeatChan <- true
	}
}

func killDMSContainers() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %s", err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Fatalf("Error listing containers: %s", err)
	}

	for _, container := range containers {
		if len(container.Names) > 0 && strings.HasPrefix(container.Names[0], "/DMS_") {
			err := cli.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{Force: true})
			if err != nil {
				log.Printf("Error killing container %s: %s", container.Names[0], err)
			}
		}
	}
}
