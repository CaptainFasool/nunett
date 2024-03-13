package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/integrations/tokenomics"
	"gitlab.com/nunet/device-management-service/models"
)

func TestGetJobTxHashesHandler(t *testing.T) {
	router := SetupTestRouter()

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
	router := SetupTestRouter()

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
	router := SetupTestRouter()

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
	router := SetupTestRouter()

	body := tokenomics.UpdateTxStatusBody{
		// Fill in required fields
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/transactions/update-status", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
