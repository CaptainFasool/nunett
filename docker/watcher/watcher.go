package watcher

import (
	"log"
	"net"
	"os/exec"
	"time"
)

const (
	heartbeatInterval = 5 * time.Second
	port              = ":9898"
)

func StartServerAndClient() {
	go startServer()
	startWatcherClient()
}

func startServer() {
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {
	log.Println("Connection established from:", c.RemoteAddr().String())

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, err := c.Write([]byte("heartbeat"))
			if err != nil {
				log.Println("Error sending heartbeat:", err)
				c.Close()
				return
			}
		}
	}
}

func startWatcherClient() {
	// This shell command runs the watcher_client in the background and redirects both its
	// standard output and standard error to watcher_client.log.
	cmdStr := "nohup go run ./docker/watcher/watcher_client.go > watcher_client.log 2>&1 &"
	cmd := exec.Command("sh", "-c", cmdStr)

	err := cmd.Start()
	if err != nil {
		log.Fatalf("Error starting the watcher: %s", err)
	}
	log.Println("Watcher process started.")
}
