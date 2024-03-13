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

type fundingRespToSPD struct {
	ComputeProviderAddr string  `json:"compute_provider_addr"`
	EstimatedPrice      float64 `json:"estimated_price"`
	MetadataHash        string  `json:"metadata_hash"`
	WithdrawHash        string  `json:"withdraw_hash"`
	RefundHash          string  `json:"refund_hash"`
	Distribute_50Hash   string  `json:"distribute_50_hash"`
	Distribute_75Hash   string  `json:"distribute_75_hash"`
}

type checkpoint struct {
	CheckpointDir string `json:"checkpoint_dir"`
	FilenamePath  string `json:"filename_path"`
	LastModified  int64  `json:"last_modified"`
}

func TestRequestServiceHandler(t *testing.T) {
	router := SetupTestRouter()

	// fill depReq details
	depReq := models.DeploymentRequest{}
	bodyBytes, _ := json.Marshal(depReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/run/request-service", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestDeploymentRequestHandler(t *testing.T) {
	router := SetupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/run/deploy", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestListCheckpointHandler(t *testing.T) {
	router := SetupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/run/checkpoints", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var checks *[]checkpoint
	err := json.Unmarshal(w.Body.Bytes(), &checks)
	assert.NoError(t, err)
}
