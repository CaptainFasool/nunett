package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var deviceStatus bool

func (h *MockHandler) DeviceStatusHandler(c *gin.Context) {
	c.JSON(200, gin.H{"online": deviceStatus})
}

func (h *MockHandler) ChangeDeviceStatusHandler(c *gin.Context) {
	var status struct {
		IsAvailable bool `json:"is_available"`
	}
	if c.Request.ContentLength == 0 {
		c.JSON(400, gin.H{"error": "empty payload data"})
		return
	}
	err := c.ShouldBindJSON(&status)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid payload data"})
		return
	}
	var msg string
	if status.IsAvailable {
		msg = "device set to online"
	} else {
		msg = "device status set to offline"
	}
	c.JSON(200, gin.H{"message": msg})
}

func TestDeviceStatusHandler(t *testing.T) {
	router := SetupMockRouter()
	tests := []struct {
		description  string
		status       bool
		expectedCode int
		expectedMsg  string
	}{
		{
			description:  "device online",
			status:       true,
			expectedCode: 200,
		},
		{
			description:  "device offline",
			status:       false,
			expectedCode: 200,
		},
	}
	for _, tc := range tests {
		deviceStatus = tc.status

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/device/status", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
		assert.Contains(t, tc.status, w.Body.String(), tc.description)
	}
}

func TestChangeDeviceStatusHandler(t *testing.T) {
	router := SetupMockRouter()

	tests := []struct {
		description  string
		payload      map[string]bool
		expectedCode int
		expectedMsg  string
	}{
		{
			description:  "change status to online",
			payload:      map[string]bool{"is_available": true},
			expectedCode: 200,
			expectedMsg:  "device set to online",
		},
		{
			description:  "change status to offline",
			payload:      map[string]bool{"is_available": false},
			expectedCode: 200,
			expectedMsg:  "device set to offline",
		},
		{
			description:  "invalid payload",
			payload:      map[string]bool{},
			expectedCode: 400,
		},
	}

	for _, tc := range tests {
		bodyBytes, _ := json.Marshal(tc.payload)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/device/status", bytes.NewBuffer(bodyBytes))
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
		assert.Equal(t, tc.expectedMsg, w.Body.String(), tc.description)
	}
}
