package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/models"
)

// TODO: Implement one single function with t.Run() for cases

func TestManualDHTUpdateHandler(t *testing.T) {
	debug = true
	router := SetupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/dht/update", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCleanupPeerHandler(t *testing.T) {
	debug = true
	router := SetupTestRouter()

	tests := []struct {
		description  string
		peerID       string
		expectedCode int
	}{
		{
			description:  "valid peer ID",
			peerID:       "Qmx0abcdefhjgiklbazbaz23",
			expectedCode: 200,
		},
		{
			description:  "empty peer ID",
			peerID:       "",
			expectedCode: 400,
		},
		{
			description:  "invalid peer ID",
			peerID:       "foobar",
			expectedCode: 400,
		},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/cleanup?peerID="+tc.peerID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
		if tc.expectedCode == 200 {
			assert.Contains(t, w.Body.String(), tc.peerID, tc.description)
		}
	}
}

func TestPingPeerHandler(t *testing.T) {
	debug = true
	router := SetupTestRouter()

	tests := []struct {
		description  string
		peerID       string
		expectedCode int
	}{
		{
			description:  "valid peer ID",
			peerID:       "Qmx0abcdefhjgiklbazbaz23",
			expectedCode: 200,
		},
		{
			description:  "self peer ID",
			peerID:       mockHostID,
			expectedCode: 400,
		},
		{
			description:  "invalid peer ID",
			peerID:       "foobar",
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
		if tc.expectedCode == 200 {
			assert.Contains(t, w.Body.String(), tc.peerID, tc.description)
		}
	}
}

func TestOldPingPeerHandler(t *testing.T) {
	debug = true
	router := SetupTestRouter()

	tests := []struct {
		description  string
		peerID       string
		expectedCode int
	}{
		{
			description:  "valid peer ID",
			peerID:       "Qmx0abcdefhjgiklbazbaz23",
			expectedCode: 200,
		},
		{
			description:  "self peer ID",
			peerID:       mockHostID,
			expectedCode: 400,
		},
		{
			description:  "invalid peer ID",
			peerID:       "foobar",
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
		req, _ := http.NewRequest("GET", "/api/v1/oldping?peerID="+tc.peerID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
		if tc.expectedCode == 200 {
			assert.Contains(t, w.Body.String(), tc.peerID, tc.description)
		}
	}
}

func TestDumpKademliaDHTHandler(t *testing.T) {
	debug = true
	router := SetupTestRouter()

	tests := []struct {
		description  string
		peers        int
		expectedCode int
		expectedMsg  string
	}{
		{
			description:  "dht with peers",
			peers:        3,
			expectedCode: 200,
		},
		{
			description:  "empty dht",
			peers:        0,
			expectedCode: 200,
			expectedMsg:  "no peers found",
		},
	}
	for _, tc := range tests {
		dumpKadDHTPeers = tc.peers

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/kad-dht", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code)
		if tc.peers > 0 {
			var peerData *[]models.PeerData
			err := json.Unmarshal(w.Body.Bytes(), &peerData)
			assert.NoError(t, err)
		} else {
			assert.Contains(t, w.Body.String(), tc.expectedMsg)
		}
	}
}

func validateMockID(id string) bool {
	if strings.HasPrefix(id, "Qm") {
		return true
	}
	return false
}

func mockDumpList() []models.PeerData {
	if dumpKadDHTPeers == 0 {
		return []models.PeerData{}
	}
	return []models.PeerData{
		{
			PeerID:      "foobarfoobar123",
			IsAvailable: true,
		},
		{
			PeerID:      "bazbazbaz123",
			IsAvailable: false,
		},
		{
			PeerID:      "foobarfoobaz567",
			IsAvailable: true,
		},
	}
}
