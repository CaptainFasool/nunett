package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/integrations/tokenomics"
	"gitlab.com/nunet/device-management-service/models"
)

type rewardRespToCPD struct {
	ServiceProviderAddr string `json:"service_provider_addr"`
	ComputeProviderAddr string `json:"compute_provider_addr"`
	RewardType          string `json:"reward_type,omitempty"`
	SignatureDatum      string `json:"signature_datum,omitempty"`
	MessageHashDatum    string `json:"message_hash_datum,omitempty"`
	Datum               string `json:"datum,omitempty"`
	SignatureAction     string `json:"signature_action,omitempty"`
	MessageHashAction   string `json:"message_hash_action,omitempty"`
	Action              string `json:"action,omitempty"`
}

func (h *MockHandler) GetJobTxHashesHandler(c *gin.Context) {
	sizeStr := c.Query("size_done")
	clean := c.Query("clean_tx")
	_, err := strconv.Atoi(sizeStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid size_done parameter"})
		return
	}
	if clean != "done" && clean != "refund" && clean != "withdraw" && clean != "" {
		c.JSON(400, gin.H{"error": "invalid clean_tx parameter"})
		return
	}
	resp := []tokenomics.TxHashResp{
		{
			TxHash:          "foobar123",
			TransactionType: "baz",
			DateTime:        "122994589",
		},
		{
			TxHash:          "foobaz456",
			TransactionType: "bar",
			DateTime:        "98930900",
		},
	}
	c.JSON(200, resp)
}

func (h *MockHandler) RequestRewardHandler(c *gin.Context) {
	var payload tokenomics.ClaimCardanoTokenBody
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid payload data"})
		return
	}
	resp := rewardRespToCPD{
		ServiceProviderAddr: "foobar",
		ComputeProviderAddr: "foobaz",
		RewardType:          "NTX",
		SignatureDatum:      "foofoofoofoofoo123",
		MessageHashDatum:    "barbarbarbarbar123",
		Datum:               "bazbazbazbazbaz123",
		SignatureAction:     "foobarfoobarfoobar123",
		MessageHashAction:   "foobazfoobazfoobaz123",
		Action:              "reward",
	}
	c.JSON(200, resp)
}

func (h *MockHandler) SendTxStatusHandler(c *gin.Context) {
	var body models.BlockchainTxStatus
	err := c.BindJSON(&body)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "cannot read payload body"})
		return
	}
	status := body.TransactionStatus
	c.JSON(200, gin.H{"message": fmt.Sprintf("sent %s status", status)})
}

func (h *MockHandler) UpdateTxStatusHandler(c *gin.Context) {
	var body tokenomics.UpdateTxStatusBody
	err := c.BindJSON(&body)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid payload data"})
		return
	}
	c.JSON(200, gin.H{"message": "transaction statuses synchronized with blockchain successfully"})
}

func TestGetJobTxHashesHandler(t *testing.T) {
	router := SetupMockRouter()

	tests := []struct {
		description  string
		query        string
		expectedCode int
	}{
		{
			description:  "valid size and clean query",
			query:        "?size_done=10&clean_tx=refund",
			expectedCode: 200,
		},
		{
			description:  "invalid size query",
			query:        "?size_done=invalid&clean_tx=refund",
			expectedCode: 400,
		},
		{
			description:  "invalid clean query",
			query:        "?size_done=10&clean_tx=foobar",
			expectedCode: 400,
		},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/transactions"+tc.query, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
	}
}

func TestRequestRewardHandler(t *testing.T) {
	router := SetupMockRouter()

	payload := tokenomics.ClaimCardanoTokenBody{
		// Fill in required fields
	}
	bodyBytes, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/transactions/request-reward", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestSendTxStatusHandler(t *testing.T) {
	router := SetupMockRouter()

	body := models.BlockchainTxStatus{
		// Fill in required fields
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/transactions/send-status", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestUpdateTxStatusHandler(t *testing.T) {
	router := SetupMockRouter()

	body := tokenomics.UpdateTxStatusBody{
		// Fill in required fields
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/transactions/update-status", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
