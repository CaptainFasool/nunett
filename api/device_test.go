package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var deviceStatus bool

type deviceAvailable struct {
	IsAvailable bool `json:"is_available"`
}

func (h *MockHandler) DeviceStatusHandler(c *gin.Context) {
	c.JSON(200, gin.H{"online": deviceStatus})
}

func (h *MockHandler) ChangeDeviceStatusHandler(c *gin.Context) {
	var status *deviceAvailable
	err := c.BindJSON(&status)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid payload data"})
		return
	}
	var msg string
	if status.IsAvailable {
		msg = "device status set to online"
	} else {
		msg = "device status set to offline"
	}
	c.JSON(200, gin.H{"message": msg})
}

func TestDeviceStatusHandler(t *testing.T) {
	router := SetupMockRouter()
	tests := []struct {
		description  string
		status       string
		expectedCode int
		expectedMsg  string
	}{
		{
			description:  "device online",
			status:       "true",
			expectedCode: 200,
		},
		{
			description:  "device offline",
			status:       "false",
			expectedCode: 200,
		},
	}
	for _, tc := range tests {
		status, err := strconv.ParseBool(tc.status)
		assert.NoError(t, err)

		deviceStatus = status

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/device/status", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
		assert.Contains(t, w.Body.String(), tc.status, tc.description)
	}
}

func TestChangeDeviceStatusHandler(t *testing.T) {
	router := SetupMockRouter()

	tests := []struct {
		description  string
		payload      []byte
		expectedCode int
		expectedMsg  string
	}{
		{
			description:  "change status to online",
			payload:      []byte(`{"is_available": true}`),
			expectedCode: 200,
			expectedMsg:  "device status set to online",
		},
		{
			description:  "change status to offline",
			payload:      []byte(`{"is_available": false}`),
			expectedCode: 200,
			expectedMsg:  "device status set to offline",
		},
		{
			description:  "invalid payload",
			payload:      []byte(`{"is_available": 350}`),
			expectedCode: 400,
		},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/device/status", bytes.NewBuffer(tc.payload))
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description, w.Body.String())
		assert.Contains(t, w.Body.String(), tc.expectedMsg, tc.description)
	}
}
