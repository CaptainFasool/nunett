package firecracker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/models"
)

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// InitVM starts the firecracker server for the specific VM. This endpoint requires a socket file.
// This socket file is further required
// Further requests are required for configuring the VM.
func InitVM(c *gin.Context) {
	var body models.InitVMRequest
	c.BindJSON(&body)

	// Check if socket file already exists
	if _, err := os.Stat(body.SocketFile); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Socket file already exists"})
		return
	}

	cmd := exec.Command("firecracker", "--api-sock", body.SocketFile)
	// output, _ := cmd.CombinedOutput() // for debugging purpose

	cmd.Stdout = os.Stdout // for debugging purpose
	// cmd.Stderr = os.Stderr // for debugging purpose
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   fmt.Sprintf("Failed to start cmd: %v", stderr.String()),
			"timestamp": time.Now(),
		})
		return
	}

	// use below code for testing only, waiting will never end the HTTP request
	// if err := cmd.Wait(); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"message":   fmt.Sprintf("Cmd returned error: %v", err),
	// 		"timestamp": time.Now(),
	// 	})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{
		"message":   "Firecracker VM initiated. Configure boot source and invoke start command",
		"timestamp": time.Now(),
	})

}

func BootSource(c *gin.Context) {
	var body models.BootSource
	c.BindJSON(&body)
	bodyJson, _ := json.Marshal(body)

	// initialize http client
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", body.SocketFile)
			},
		},
	}

	// set the HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, "http://localhost/boot-source", bytes.NewBuffer(bodyJson))
	if err != nil {
		c.JSON(400, gin.H{
			"message":   "Error in making PUT request to /boot-source with give body",
			"timestamp": time.Now(),
		})
		return
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	_, err = client.Do(req)
	if err != nil {
		c.JSON(400, gin.H{
			"message":   "Error in making PUT request to /boot-source with give body",
			"timestamp": time.Now(),
		})
		return
	}
}

func Drives(c *gin.Context) {
	var body models.Drive
	c.ShouldBindJSON(&body)
	body.DriveID = c.Param("drive_id")

	bodyJson, _ := json.Marshal(body)
	log.Println(string(bodyJson))

	// initialize http client
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", body.SocketFile)
			},
		},
	}

	// set the HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, "http://localhost/drives/"+body.DriveID, bytes.NewBuffer(bodyJson))
	if err != nil {
		panic(err)
		// c.JSON(400, gin.H{
		// 	"message":   "Error in making PUT request to /drives with give body",
		// 	"timestamp": time.Now(),
		// })
		// return
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	_, err = client.Do(req)
	if err != nil {
		panic(err)
		// c.JSON(400, gin.H{
		// 	"message":   "Error in making PUT request to /drives with give body",
		// 	"timestamp": time.Now(),
		// })
		// return
	}

}

func MachineConfig(c *gin.Context) {
	var body models.MachineConfig
	c.BindJSON(&body)
	bodyJson, _ := json.Marshal(body)

	log.Println(string(bodyJson))

	// initialize http client
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", body.SocketFile)
			},
		},
	}

	// set the HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, "http://localhost/machine-config", bytes.NewBuffer(bodyJson))
	if err != nil {
		c.JSON(400, gin.H{
			"message":   "Error in making PUT request to /machine-config with give body",
			"timestamp": time.Now(),
		})
		return
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	_, err = client.Do(req)
	if err != nil {
		c.JSON(400, gin.H{
			"message":   "Error in making PUT request to /machine-config with give body",
			"timestamp": time.Now(),
		})
		return
	}
}

func Actions(c *gin.Context) {
	var body models.Action
	c.BindJSON(&body)

	bodyJson, _ := json.Marshal(body)

	// initialize http client
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", body.SocketFile)
			},
		},
	}

	// set the HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, "http://localhost/actions", bytes.NewBuffer(bodyJson))
	if err != nil {
		panic(err)
		// c.JSON(400, gin.H{
		// 	"message":   "Error in making PUT request to /actions with give body",
		// 	"timestamp": time.Now(),
		// })
		// return
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	_, err = client.Do(req)
	if err != nil {
		panic(err)
		// c.JSON(400, gin.H{
		// 	"message":   "Error in making PUT request to /actions with give body",
		// 	"timestamp": time.Now(),
		// })
		// return
	}
}
