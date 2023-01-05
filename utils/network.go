package utils

import (
	"bytes"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// DMSBaseURL is base of all API endpoints on DMS
	DMSBaseURL = "http://localhost:9999/api/v1"
	// AdapterGrpcURL is DMS connects to nunet-adapter
	AdapterGrpcURL = "localhost:60777"
)

// MakeInternalRequest is a helper method to make call to DMS's own API
func MakeInternalRequest(c *gin.Context, methodType, internalEndpoint string, body []byte) http.Response {
	req, err := http.NewRequest(methodType, DMSBaseURL+internalEndpoint, bytes.NewBuffer(body))
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
			"timestamp": time.Now(),
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
