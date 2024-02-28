package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/models"
)

var mockHostID = "Qm01testabcdefghjiklgfoobar123"

func (h *MockHandler) CleanupPeerHandler(c *gin.Context) {
	id := c.Query("peerID")
	if !validateMockID(id) {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid peerID query"})
		return
	}
	c.JSON(200, gin.H{"message": fmt.Sprintf("successfully cleaned up peer: %s", id)})
}

func (m *MockHandler) PingPeerHandler(c *gin.Context) {
	id := c.Query("peerID")
	if !validateMockID(id) {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid peerID query"})
		return
	} else if id == mockHostID {
		c.AbortWithStatusJSON(400, gin.H{"error": "peerID string cannot be self peer ID"})
		return
	}
	c.JSON(200, gin.H{"message": fmt.Sprintf("ping successful with peer %s", id), "peer_in_dht": true, "RTT": "28859000"})
}

func (m *MockHandler) OldPingPeerHandler(c *gin.Context) {
	id := c.Query("peerID")
	if !validateMockID(id) {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid peerID query"})
		return
	} else if id == mockHostID {
		c.AbortWithStatusJSON(400, gin.H{"error": "peerID string cannot be self peer ID"})
		return
	}
	c.JSON(200, gin.H{"message": fmt.Sprintf("ping successful with peer %s", id), "peer_in_dht": true, "RTT": "28859000"})

}

func (m *MockHandler) DumpKademliaDHTHandler(c *gin.Context) {
	// TODO: set this as a function with parameters
	dht := []models.PeerData{
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
	c.JSON(200, dht)
}

func TestCleanupPeerHandler(t *testing.T) {
	router := SetupMockRouter()

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
			assert.Contains(t, tc.peerID, w.Body.String(), tc.description)
		}
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
			assert.Contains(t, tc.peerID, w.Body.String, tc.description)
		}
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
		req, _ := http.NewRequest("GET", "/api/v1/old-ping?peerID="+tc.peerID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
		if tc.expectedCode == 200 {
			assert.Contains(t, tc.peerID, w.Body.String, tc.description)
		}
	}
}

func TestDumpKademliaDHTHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/dump-dht", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func validateMockID(id string) bool {
	if strings.HasPrefix(id, "Qm") {
		return true
	}
	return false
}
