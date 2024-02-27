package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeviceStatusHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/device/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestChangeDeviceStatusHandler(t *testing.T) {
	router := SetupMockRouter()

	tests := []struct {
		description  string
		payload      map[string]bool
		expectedCode int
	}{
		{
			description:  "Change status to online",
			payload:      map[string]bool{"is_available": true},
			expectedCode: 200,
		},
		{
			description:  "Change status to offline",
			payload:      map[string]bool{"is_available": false},
			expectedCode: 200,
		},
		{
			description:  "Invalid payload",
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
	}
}
