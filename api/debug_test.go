package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanupPeerHandler(t *testing.T) {
	router := SetupMockRouter()

	tests := []struct {
		description  string
		peerID       string
		expectedCode int
	}{
		{
			description:  "valid peer ID",
			peerID:       "validPeerID",
			expectedCode: 200,
		},
		{
			description:  "Invalid peer ID",
			peerID:       "invalidPeerID",
			expectedCode: 400,
		},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/cleanup?peerID="+tc.peerID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
	}
}

func TestPingPeerHandler(t *testing.T) {
	router := SetupMockRouter()

	tests := []struct {
		description  string
		peerID       string
		expectedCode int
	}{
		{
			description:  "valid peer ID",
			peerID:       "validPeerID",
			expectedCode: 200,
		},
		{
			description:  "self peer ID",
			peerID:       "selfPeerID",
			expectedCode: 400,
		},
		{
			description:  "invalid peer ID",
			peerID:       "invalidPeerID",
			expectedCode: 400,
		},
		{
			description:  "missing peer ID",
			peerID:       "",
			expectedCode: 400,
		},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/ping?peerID="+tc.peerID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
	}
}

func TestOldPingPeerHandler(t *testing.T) {
	router := SetupMockRouter()

	tests := []struct {
		description  string
		peerID       string
		expectedCode int
	}{
		{
			description:  "valid peer ID",
			peerID:       "validPeerID",
			expectedCode: 200,
		},
		{
			description:  "self peer ID",
			peerID:       "selfPeerID",
			expectedCode: 400,
		},
		{
			description:  "invalid peer ID",
			peerID:       "invalidPeerID",
			expectedCode: 400,
		},
		{
			description:  "missing peer ID",
			peerID:       "",
			expectedCode: 400,
		},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/old-ping?peerID="+tc.peerID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
	}
}

func TestDumpKademliaDHTHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/dump-dht", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
