package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/integrations/oracle"
	"gitlab.com/nunet/device-management-service/integrations/tokenomics"
	library "gitlab.com/nunet/device-management-service/lib"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding"
	"gitlab.com/nunet/device-management-service/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MockOracle struct {
	mock.Mock
}

// Mock implemention of the oracle's WithdrawTokenRequest method
func (m *MockOracle) WithdrawTokenRequest(req *oracle.RewardRequest) (*oracle.RewardResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*oracle.RewardResponse), args.Error(1)
}

func SetUpRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")

	onboardingRoute := v1.Group("/onboarding")
	{
		onboardingRoute.GET("/metadata", onboarding.GetMetadata)
		onboardingRoute.POST("/onboard", onboarding.Onboard)
		onboardingRoute.GET("/provisioned", onboarding.ProvisionedCapacity)
		onboardingRoute.GET("/address/new", onboarding.CreatePaymentAddress)
	}

	txRoute := v1.Group("/transactions")
	{
		txRoute.POST("/request-reward", tokenomics.HandleRequestReward)
		txRoute.POST("/send-status", tokenomics.HandleSendStatus)
		txRoute.GET("", tokenomics.GetJobTxHashes)
	}

	return router
}

func TestHandleRequestReward(t *testing.T) {
	// Create a new instance of MockOracle
	mockOracle := new(MockOracle)

	// Set up expected behavior for WithdrawTokenRequest
	mockOracle.On("WithdrawTokenRequest", mock.Anything).Return(&oracle.RewardResponse{
		RewardType:        "TestReward",
		MessageHashDatum:  "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		Datum:             "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		MessageHashAction: "0x7890abcdef1234567890abcdef1234567890abcdef1234567890abcdef123456",
		Action:            "ConfirmReward",
	}, nil)
	oracle.Oracle = mockOracle

	// Open an in-memory SQLite database for testing
	mockDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm database", err)
	}

	// Run the migrations to create the necessary tables in the in-memory database
	err = mockDB.AutoMigrate(&models.Services{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when migrating the database", err)
	}
	db.DB = mockDB

	router := SetUpRouter()

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := `{
			"compute_provider_address": "someAddress",
			"tx_hash": "someTxHash"
		}`

		req, err := http.NewRequest("POST", "/api/v1/transactions/request-reward", bytes.NewBufferString(payload))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		router.ServeHTTP(w, req)

		expectedResponse := `{
			"reward_type": "TestReward",
			"message_hash_datum": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"datum": "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"message_hash_action": "0x7890abcdef1234567890abcdef1234567890abcdef1234567890abcdef123456",
			"action": "ConfirmReward"
		}`

		assert.Equal(t, 200, w.Code)
		assert.JSONEq(t, expectedResponse, w.Body.String())
	})

	t.Run("InvalidJSONPayload", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Intentionally malformed JSON (missing closing brace)
		payload := `{
			"compute_provider_address": "someAddress",
			"tx_hash": "someTxHash"
		`
		req, _ := http.NewRequest("POST", "/api/v1/transactions/request-reward", bytes.NewBufferString(payload))
		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid JSON format")
	})

	t.Cleanup(func() {
		originalOracle := oracle.Oracle
		oracle.Oracle = originalOracle
	})
}

func TestHandleSendStatus(t *testing.T) {
	// Open an in-memory SQLite database for testing
	mockDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm database", err)
	}

	// Run the migrations to create the necessary tables in the in-memory database
	err = mockDB.AutoMigrate(&models.Services{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when migrating the database", err)
	}

	db.DB = mockDB

	router := SetUpRouter()

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := `{"status": "completed", "tx_hash": "0xabcdef1234567890"}`
		req, err := http.NewRequest("POST", "/api/v1/transactions/send-status", bytes.NewBufferString(payload))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		router.ServeHTTP(w, req)

		expectedResponse := `{"message": "transaction status  acknowledged"}`
		assert.Equal(t, 200, w.Code)
		assert.JSONEq(t, expectedResponse, w.Body.String())
	})

	t.Run("InvalidPayload", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/transactions/send-status", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
		assert.Contains(t, w.Body.String(), "cannot read payload body")
	})
}

func TestGetJobTxHashes(t *testing.T) {
	// Open an in-memory SQLite database for testing
	mockDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm database", err)
	}

	// Run the migrations to create the necessary tables in the in-memory database
	err = mockDB.AutoMigrate(&models.Services{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when migrating the database", err)
	}

	db.DB = mockDB

	router := SetUpRouter()

	// test when no records in services table
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/transactions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
	assert.Contains(t, w.Body.String(), "no job deployed to request reward for")

	// test when there are records in services table
	mockHash := utils.RandomString(64)
	mockDB.Create(&models.Services{TxHash: mockHash, LogURL: "log.nunet.io"})

	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/v1/transactions", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), mockHash)

	t.Run("NoData", func(t *testing.T) {
		// Mock the database to return an empty list

	})
}

func TestCardanoAddressRoute(t *testing.T) {
	router := SetUpRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/onboarding/address/new", nil)
	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(body), "address")
	assert.Contains(t, string(body), "mnemonic")

	var jsonMap map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &jsonMap)

	assert.NotEmpty(t, jsonMap)
}

func TestEthereumAddressRoute(t *testing.T) {
	router := SetUpRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/onboarding/address/new?blockchain=ethereum", nil)
	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(body), "address")
	assert.Contains(t, string(body), "private_key")

	var jsonMap map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &jsonMap)

	assert.NotEmpty(t, jsonMap)
}

func TestProvisionedRoute(t *testing.T) {
	router := SetUpRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/onboarding/provisioned", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "cpu")
	assert.Contains(t, w.Body.String(), "memory")
}

func TestOnboardEmptyRequest(t *testing.T) {
	expectedResponse := `{"error":"invalid request data"}`
	router := SetUpRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/onboarding/onboard", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, expectedResponse, w.Body.String())
}

func TestOnboard_CapacityTooLowTooHigh(t *testing.T) {
	onboarding.FS = afero.NewMemMapFs()
	onboarding.AFS = &afero.Afero{Fs: onboarding.FS}
	onboarding.AFS.Mkdir("/etc/nunet", 0777)
	expectedCPUResponse := "CPU should be between 10% and 90% of the available CPU"
	expectedRAMResponse := "memory should be between 10% and 90% of the available memory"

	router := SetUpRouter()
	w := httptest.NewRecorder()

	type TestRequestPayload struct {
		Memory      int64  `json:"memory"`
		CPU         int64  `json:"cpu"`
		Channel     string `json:"channel"`
		PaymentAddr string `json:"payment_addr"`
		Cardano     bool   `json:"cardano"`
	}
	totalCpu := library.GetTotalProvisioned().CPU
	totalMem := library.GetTotalProvisioned().Memory

	// test too low CPU (less than 10% of machine resource)
	lowCPUTestPayload := TestRequestPayload{
		Memory:      int64(float64(totalMem) * 0.5),  // 50% acceptable
		CPU:         int64(float64(totalCpu) * 0.05), // 5% unacceptable
		Channel:     "nunet-test",
		PaymentAddr: "addr1q99z75su8d8w0jv6drfnr3tuyycflcg4pqpvpnvfzlmmdl7m4nxzjpxhvx477ruhswnrkuqju0kyhx4mvwr0geqyfass7rwta8",
		Cardano:     false,
	}
	jsonPayload, _ := json.Marshal(lowCPUTestPayload)
	req, _ := http.NewRequest("POST", "/api/v1/onboarding/onboard", bytes.NewBuffer((jsonPayload)))
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), expectedCPUResponse)

	// test too high CPU (more than 90% of machine resource)
	highCPUTestPayload := TestRequestPayload{
		Memory:      int64(float64(totalMem) * 0.5),  // 50% acceptable
		CPU:         int64(float64(totalCpu) * 0.95), // 95% unacceptable
		Channel:     "nunet-test",
		PaymentAddr: "addr1q99z75su8d8w0jv6drfnr3tuyycflcg4pqpvpnvfzlmmdl7m4nxzjpxhvx477ruhswnrkuqju0kyhx4mvwr0geqyfass7rwta8",
		Cardano:     false,
	}
	jsonPayload, _ = json.Marshal(highCPUTestPayload)
	req, _ = http.NewRequest("POST", "/api/v1/onboarding/onboard", bytes.NewBuffer((jsonPayload)))
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), expectedCPUResponse)

	// test too low memory (less than 10% of machine resource)
	lowRAMTestPayload := TestRequestPayload{
		Memory:      int64(float64(totalMem) * 0.05), // 5% unacceptable
		CPU:         int64(float64(totalCpu) * 0.5),  // 50% acceptable
		Channel:     "nunet-test",
		PaymentAddr: "addr1q99z75su8d8w0jv6drfnr3tuyycflcg4pqpvpnvfzlmmdl7m4nxzjpxhvx477ruhswnrkuqju0kyhx4mvwr0geqyfass7rwta8",
		Cardano:     false,
	}
	jsonPayload, _ = json.Marshal(lowRAMTestPayload)
	req, _ = http.NewRequest("POST", "/api/v1/onboarding/onboard", bytes.NewBuffer((jsonPayload)))
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), expectedRAMResponse)

	// test too high VPU (more than 90% of machine resource)
	highRAMTestPayload := TestRequestPayload{
		Memory:      int64(float64(totalMem) * 0.95), // 95% unacceptable
		CPU:         int64(float64(totalCpu) * 0.5),  // 50% acceptable
		Channel:     "nunet-test",
		PaymentAddr: "addr1q99z75su8d8w0jv6drfnr3tuyycflcg4pqpvpnvfzlmmdl7m4nxzjpxhvx477ruhswnrkuqju0kyhx4mvwr0geqyfass7rwta8",
		Cardano:     false,
	}
	jsonPayload, _ = json.Marshal(highRAMTestPayload)
	req, _ = http.NewRequest("POST", "/api/v1/onboarding/onboard", bytes.NewBuffer((jsonPayload)))
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), expectedRAMResponse)

}
