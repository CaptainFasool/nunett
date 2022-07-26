package firecracker

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
)

func NewClient(sockFile string) *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sockFile)
			},
		},
	}

	return client
}

func MakeRequest(c *gin.Context, client *http.Client, uri string, body []byte, errMsg string) {
	// set the HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(body))

	if err != nil {
		c.JSON(400, gin.H{
			"message":   errMsg,
			"timestamp": time.Now(),
		})
		return
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	_, err = client.Do(req)
	if err != nil {
		c.JSON(400, gin.H{
			"message":   errMsg,
			"timestamp": time.Now(),
		})
		return
	}

}

// InitVM starts the firecracker server for the specific VM. This endpoint requires a socket file.
// This socket file is further required
// Further requests are required for configuring the VM.
func InitVM(c *gin.Context) {

	// Check if socket file already exists
	if _, err := os.Stat("/tmp/firecracker.socket"); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Socket file already exists"})
		return
	}

	cmd := exec.Command("firecracker", "--api-sock", "/tmp/firecracker.socket")
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
		"message":   "VM initiated. Add boot-source, add filesystem, invoke start",
		"timestamp": time.Now(),
	})

}

func BootSource(c *gin.Context) {
	var jsonStr = []byte(`{"kernel_image_path":"/home/santosh/firecracker/vmlinux.bin", "boot_args": "console=ttyS0 reboot=k panic=1 pci=off"}`)

	// initialize http client
	client := NewClient("/tmp/firecracker.socket")

	errMsg := "Error in making PUT request to /boot-source with give body"

	MakeRequest(c, client, "http://localhost/boot-source", jsonStr, errMsg)
}

func Drives(c *gin.Context) {
	var jsonStr = []byte(`{"drive_id": "rootfs", "path_on_host":"/home/santosh/firecracker/bionic.rootfs.ext4", "is_root_device": true, "is_read_only": false}`)

	client := NewClient("/tmp/firecracker.socket")

	errMsg := "Error in making PUT request to /drives with give body"

	MakeRequest(c, client, "http://localhost/drives/rootfs", jsonStr, errMsg)

}

func MachineConfig(c *gin.Context) {
	var jsonStr = []byte(`{"vcpu_count": 2,"mem_size_mib": 512}`)

	// initialize http client
	client := NewClient("/tmp/firecracker.socket")

	errMsg := "Error in making PUT request to /machine-config with give body"

	MakeRequest(c, client, "http://localhost/machine-config", jsonStr, errMsg)

}

func NetworkInterfaces(c *gin.Context) {
	var jsonStr = []byte(`{ "iface_id": "eth0", "guest_mac": "AA:FC:00:00:00:01", "host_dev_name": "tap1" }`)

	// initialize http client
	client := NewClient("/tmp/firecracker.socket")

	errMsg := "Error in making PUT request to /network-interfaces with give body"

	MakeRequest(c, client, "http://localhost/network-interfaces/eth0", jsonStr, errMsg)
}

func Actions(c *gin.Context) {
	var jsonStr = []byte(`{"action_type": "InstanceStart"}`)

	// initialize http client
	client := NewClient("/tmp/firecracker.socket")

	errMsg := "Error in making PUT request to /actions with give body"

	MakeRequest(c, client, "http://localhost/actions", jsonStr, errMsg)

}
