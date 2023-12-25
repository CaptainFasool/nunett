package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/libp2p/machines"
	"gitlab.com/nunet/device-management-service/onboarding"
	"gitlab.com/nunet/device-management-service/utils"

	"github.com/stretchr/testify/mock"
	"gitlab.com/nunet/device-management-service/integrations/oracle"
	"gitlab.com/nunet/device-management-service/integrations/tokenomics"
	library "gitlab.com/nunet/device-management-service/lib"
	dmslibp2p "gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"

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

	run := v1.Group("/run")
	{
		run.POST("/request-service", machines.RequestServiceHandler)
		run.GET("/deploy", machines.DeploymentRequestHandler)
	}

	txRoute := v1.Group("/transactions")
	{
		txRoute.POST("/request-reward", tokenomics.HandleRequestReward)
		txRoute.POST("/send-status", tokenomics.HandleSendStatus)
		txRoute.GET("", tokenomics.GetJobTxHashes)
	}

	return router
}

func startMockWebSocketServer() *http.Server {
	upgrader := websocket.Upgrader{}
	server := &http.Server{
		Addr: "localhost:8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()
			for {
				messageType, p, err := conn.ReadMessage()
				if err != nil {
					return
				}
				if err := conn.WriteMessage(messageType, p); err != nil {
					return
				}
			}
		}),
	}

	go server.ListenAndServe()
	return server
}

func TestHandleDeploymentRequest(t *testing.T) {
	// Setup the router
	router := SetUpRouter()

	// Start the mock WebSocket server
	mockServer := startMockWebSocketServer()
	defer mockServer.Close()
	t.Run("SuccessfulWebsocketConnection", func(t *testing.T) {
		// Use a WebSocket client to connect to the endpoint
		c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/api/v1/run/deploy", nil)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer c.Close()

		// Check if the connection is successful
		assert.NotNil(t, c)
	})

	t.Run("FailedWebsocketConnection", func(t *testing.T) {
		// Simulate a scenario where the WebSocket connection fails
		c, _, err := websocket.DefaultDialer.Dial("ws://invalid-endpoint", nil)
		assert.Error(t, err)
		assert.Nil(t, c)
	})

	t.Run("WebsocketDataTransfer", func(t *testing.T) {
		// Use a WebSocket client to connect to the endpoint
		c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/api/v1/run/deploy", nil)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer c.Close()

		// Send some data and check the response
		err = c.WriteMessage(websocket.TextMessage, []byte("test message"))
		assert.NoError(t, err)

		_, p, err := c.ReadMessage()
		assert.NoError(t, err)
		assert.Equal(t, "test message", string(p))
	})

	t.Run("GinHandlerWebsocketUpgrade", func(t *testing.T) {
		// Create a new HTTP server
		server := httptest.NewServer(router)
		defer server.Close()

		// Create a new HTTP request
		req, err := http.NewRequest("GET", server.URL+"/api/v1/run/deploy", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Set headers to simulate a WebSocket upgrade request
		req.Header.Set("Connection", "upgrade")
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Sec-WebSocket-Version", "13")
		req.Header.Set("Sec-WebSocket-Key", "some-random-key")

		// Send the request and get the response
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		// Check if the response header indicates a WebSocket upgrade
		assert.Equal(t, "websocket", resp.Header.Get("Upgrade"))
		assert.Equal(t, "101 Switching Protocols", resp.Status)
	})
}

func TestHandleRequestService(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Connect to the in-memory database
	mockDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Errorf("error trying to initialize mock db: %v", err)
	}
	db.DB = mockDB

	err = mockDB.AutoMigrate(&models.Libp2pInfo{})
	if err != nil {
		t.Fatalf("unable to migrate Libp2pInfo table - %v", err)
	}

	// Setup the router
	router := SetUpRouter()

	// Setup host
	privK, pubK, _ := dmslibp2p.GenerateKey(time.Now().Unix())
	err = dmslibp2p.SaveNodeInfo(privK, pubK, true)
	if err != nil {
		t.Fatal("Failed to save node info")
	}
	host, _, err := dmslibp2p.NewHost(ctx, privK, true)
	if err != nil {
		t.Fatal("Failed to create libp2p host")
	}
	dmslibp2p.DMSp2pSet(host, nil)

	t.Run("InvalidPOSTData", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := `{
			"RequesterWalletAddress": "someInvalidAddress"
		}`
		req, _ := http.NewRequest("POST", "/api/v1/run/request-service", bytes.NewBufferString(payload))
		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "no peers found with matched specs", response["error"])
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

	db.DB = mockDB

	// Run the migrations to create the necessary tables in the in-memory database
	err = db.DB.AutoMigrate(&models.Services{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when migrating the database", err)
	}

	router := SetUpRouter()

	// test without required parameter
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/transactions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "invalid size_done parameter")

	// test when no records in services table
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/transactions?size_done=1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
	assert.Contains(t, w.Body.String(), "no job deployed to request reward for")

	// test when there are records in services table
	mockHash := utils.RandomString(64)
	db.DB.Create(&models.Services{TxHash: mockHash, LogURL: "log.nunet.io", TransactionType: "done"})

	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/v1/transactions?size_done=1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), mockHash)
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
