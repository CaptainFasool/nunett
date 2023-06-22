package utils

import (
	"bytes"
	"fmt"
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
func MakeInternalRequest(c *gin.Context, methodType, internalEndpoint string, body []byte) http.Response {
	endpoint, err := GetInternalBaseURL(internalEndpoint)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(
		methodType,
		endpoint,
		bytes.NewBuffer(body),
	)
	if err != nil {
		panic(err)
	}

	client := http.Client{}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
		// c.JSON(400, gin.H{
		// 	"message":   fmt.Sprintf("Error making %s request to %s", methodType, internalEndpoint),
		// 	"timestamp": time.Now(),
		// })
		// return
	}

	return *resp
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
