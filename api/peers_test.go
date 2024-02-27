package api

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListPeersHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/peers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestListDHTPeersHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/peers/dht", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestListKadDHTPeersHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/peers/kad-dht", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestSelfPeerInfoHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/peers/self", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestListChatHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/peers/chat", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestClearChatHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/peers/chat/clear", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestStartChatHandlerWithQueries(t *testing.T) {
	router := SetupMockRouter()

	tests := []struct {
		description  string
		query        string
		expectedCode int
	}{
		{
			description:  "valid peer ID",
			query:        "?peerID=somePeerID",
			expectedCode: 200,
		},
		{
			description:  "invalid peer ID",
			query:        "?peerID=invalidPeerID",
			expectedCode: 400,
		},
		{
			description:  "missing peer ID",
			query:        "",
			expectedCode: 400,
		},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/peers/chat/start"+tc.query, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
	}
}

func TestJoinChatHandlerWithQueries(t *testing.T) {
	router := SetupMockRouter()

	tests := []struct {
		description  string
		query        string
		expectedCode int
	}{
		{
			description:  "Valid stream ID",
			query:        "?streamID=123",
			expectedCode: 200,
		},
		{
			description:  "Invalid stream ID",
			query:        "?streamID=abc",
			expectedCode: 400,
		},
		{
			description:  "Missing stream ID",
			query:        "",
			expectedCode: 400,
		},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/peers/chat/join"+tc.query, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
	}
}

func TestDumpDHTHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/dht", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestDefaultDepReqPeerHandlerWithQueries(t *testing.T) {
	router := SetupMockRouter()

	tests := []struct {
		description  string
		query        string
		expectedCode int
	}{
		{
			description:  "valid peer ID",
			query:        "?peerID=somePeerID",
			expectedCode: 200,
		},
		{
			description:  "remove default peer ID",
			query:        "?peerID=0",
			expectedCode: 200,
		},
		{
			description:  "missing peer ID",
			query:        "",
			expectedCode: 200,
		},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/peers/depreq"+tc.query, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
	}
}

func TestClearFileTransferRequestsHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/peers/file/clear", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestListFileTransferRequestsHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/peers/file", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestSendFileTransferHandlerWithQueries(t *testing.T) {
	router := SetupMockRouter()

	tests := []struct {
		description  string
		query        string
		expectedCode int
	}{
		{
			description:  "valid peer ID and file path",
			query:        "?peerID=somePeerID&filePath=/path/to/file",
			expectedCode: 200,
		},
		{
			description:  "missing peer ID",
			query:        "?filePath=/path/to/file",
			expectedCode: 400,
		},
		{
			description:  "missing file path",
			query:        "?peerID=somePeerID",
			expectedCode: 400,
		},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/peers/file/send"+tc.query, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
	}
}

func TestAcceptFileTransferHandlerWithQueries(t *testing.T) {
	router := SetupMockRouter()

	tests := []struct {
		description  string
		query        string
		expectedCode int
	}{
		{
			description:  "valid stream ID",
			query:        "?streamID=123",
			expectedCode: 200,
		},
		{
			description:  "invalid stream ID",
			query:        "?streamID=abc",
			expectedCode: 400,
		},
		{
			description:  "missing stream ID",
			query:        "",
			expectedCode: 400,
		},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/peers/file/accept"+tc.query, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
	}
}
