package firecracker

import (
	"bytes"
	"fmt"
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

func LaunchVM(c *gin.Context) {
	var body models.LaunchVMRequest
	c.BindJSON(&body)

	// TODO: In future, need to automatically set socket file name based on VM/task name

	// Check if socket file already exists
	if _, err := os.Stat(body.SocketFile); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Socket file already exists"})
		return
	}

	cmd := exec.Command("firecracker", "--api-sock", body.SocketFile, "--config-file", body.ConfigFile)
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
		"message":   "Successfully created VM with specified config file",
		"timestamp": time.Now(),
	})

}
