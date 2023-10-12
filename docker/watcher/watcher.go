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
	watchdogBinary    = "./docker/watcher/watchdog/watchdog"
)

// StartWatcherAndInvokeWatchdog starts the watcher server and invokes the watchdog.
func StartWatcherAndInvokeWatchdog() {
	go startServer()
	invokeWatchdog()
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

func invokeWatchdog() {
	cmd := exec.Command(watchdogBinary)

	err := cmd.Start()
	if err != nil {
		log.Fatalf("Error starting the watchdog: %s", err)
	}
	log.Println("Watchdog process started.")
}
