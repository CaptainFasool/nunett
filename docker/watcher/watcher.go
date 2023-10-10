/*
Package watcher provides functionality to monitor the DMS (Device Management Service) through
heartbeats. It starts both a server to receive heartbeats and a watcher client that checks
the health of the DMS. If the DMS doesn't send a heartbeat within a given interval, the watcher
client will initiate cleanup procedures.
*/
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
		conn.Close() // Simply close the connection after accepting it, since we're only receiving heartbeats.
	}
}

func startWatcherClient() {
	cmd := exec.Command("./docker/watcher/watcher_client")
	err := cmd.Start()
	if err != nil {
		log.Fatalf("Error starting the watcher: %s", err)
	}
	log.Println("Watcher process started with PID:", cmd.Process.Pid)
}

func SendHeartbeat() {
	// Connect to the watcher and send a heartbeat
	conn, err := net.Dial("tcp", "localhost"+port)
	if err != nil {
		log.Println("Error connecting to watcher:", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte("heartbeat"))
	if err != nil {
		log.Println("Error sending heartbeat:", err)
	}
}
