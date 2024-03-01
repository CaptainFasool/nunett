package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

func (h *MockHandler) RequestServiceHandler(c *gin.Context) {
	var depReq models.DeploymentRequest
	err := c.BindJSON(&depReq)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	resp := fundingRespToSPD{
		ComputeProviderAddr: "foobar123",
		EstimatedPrice:      12.34,
		MetadataHash:        depReq.MetadataHash,
		WithdrawHash:        depReq.WithdrawHash,
		RefundHash:          depReq.RefundHash,
		Distribute_50Hash:   depReq.Distribute_50Hash,
		Distribute_75Hash:   depReq.Distribute_75Hash,
	}
	c.JSON(200, resp)
}

func (h *MockHandler) DeploymentRequestHandler(c *gin.Context) {
	// original func do not return anything
}

func (h *MockHandler) ListCheckpointHandler(c *gin.Context) {
	checkpoints := []checkpoint{
		{
			CheckpointDir: "/foo/bar/1",
			FilenamePath:  "/foo/bar/baz.tar.gz",
			LastModified:  1609459200,
		},
		{
			CheckpointDir: "/foo/bar/2",
			FilenamePath:  "/foo/bar/baz2.tar.gz",
			LastModified:  1612137600,
		},
		{
			CheckpointDir: "/foo/bar/3",
			FilenamePath:  "/foo/bar/baz3.tar.gz",
			LastModified:  1614556800,
		},
	}
	c.JSON(200, checkpoints)
}

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

	var checks *[]checkpoint
	err := json.Unmarshal(w.Body.Bytes(), &checks)
	assert.NoError(t, err)
}
