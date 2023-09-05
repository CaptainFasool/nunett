package utils

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/internal/config"
)

// GetInternalBaseURL is a helper method to allow calls to any resources
func GetInternalBaseURL(internalEndpoint string) (string, error) {
	if internalEndpoint == "" {
		return "", fmt.Errorf("internalEndpoint cannot be empty")
	}

	endpoint := fmt.Sprintf(
		"http://localhost:%d%s",
		config.GetConfig().Rest.Port,
		internalEndpoint,
	)

	return endpoint, nil
}

// MakeInternalRequest is a helper method to make call to DMS's own API
func MakeInternalRequest(c *gin.Context, methodType, internalEndpoint string, body []byte) (*http.Response, error) {
	endpoint, err := GetInternalBaseURL(internalEndpoint)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		methodType,
		endpoint,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	client := http.Client{}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
		// c.JSON(400, gin.H{
		// 	"message":   fmt.Sprintf("Error making %s request to %s", methodType, internalEndpoint),
		// 	"timestamp": time.Now(),
		// })
		// return
	}

	return resp, nil
}

func MakeRequest(c *gin.Context, client *http.Client, uri string, body []byte, errMsg string) {
	// set the HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(body))

	if err != nil {
		c.JSON(400, gin.H{
			"message":   errMsg,
			"timestamp": time.Now().In(time.UTC),
		})
		return
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	_, err = client.Do(req)
	if err != nil {
		// c.JSON(400, gin.H{
		// 	"message":   errMsg,
		// 	"timestamp": time.Now(),
		// })
		// return
		panic(err)
	}
}

// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
