package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/models"
)

func TestRequestServiceHandler(t *testing.T) {
	router := SetupMockRouter()

	// fill depReq details
	depReq := models.DeploymentRequest{}
	bodyBytes, _ := json.Marshal(depReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/run/request-service", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestDeploymentRequestHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/run/deploy", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestListCheckpointHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/run/checkpoints", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
