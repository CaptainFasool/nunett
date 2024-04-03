package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"gitlab.com/nunet/device-management-service/firecracker"
	"gitlab.com/nunet/device-management-service/integrations/tokenomics"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
)

var (
	debug       bool
	defaultPeer string
	mockHostID  = "Qm01testabcdefghjiklgfoobar123"
	// DumpKademliaDHTHandler
	dumpKadDHTPeers int
	// ListDHTPeersHandler
	dhtPeers int
	// ListKadDHTPeersHandler
	kadDHTPeers      int
	mockInboundChats int
)

type MockHandler struct {
	buf *bytes.Buffer
}

func newMockHandler() *MockHandler {
	return &MockHandler{
		buf: new(bytes.Buffer),
	}
}

// type MockOracle struct {
// 	mock.Mock
// }
//
// func (o *MockOracle) WithdrawTokenRequest(req *oracle.RewardRequest) (*oracle.RewardResponse, error) {
// 	args := o.Called(req)
// 	return args.Get(0).(*oracle.RewardResponse), args.Error(1)
// }

func SetupMockRouter() *gin.Engine {
	m := newMockHandler()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	onboard := v1.Group("/onboarding")
	{
		onboard.GET("/metadata", m.GetMetadataHandler)
		onboard.GET("/status", m.OnboardStatusHandler)
		onboard.GET("/provisioned", m.ProvisionedCapacityHandler)
		onboard.GET("/address/new", m.CreatePaymentAddressHandler)
		onboard.POST("/onboard", m.OnboardHandler)
		onboard.POST("/resource-config", m.ResourceConfigHandler)
		onboard.DELETE("/offboard", m.OffboardHandler)
	}
	device := v1.Group("/device")
	{
		device.GET("/status", m.DeviceStatusHandler)
		device.POST("/status", m.ChangeDeviceStatusHandler)
	}
	vm := v1.Group("/vm")
	{
		vm.POST("/start-default", m.StartDefaultHandler)
		vm.POST("/start-custom", m.StartCustomHandler)
	}
	run := v1.Group("/run")
	{
		run.GET("/checkpoints", m.ListCheckpointHandler)
		run.GET("/deploy", m.DeploymentRequestHandler)
		run.POST("/request-service", m.RequestServiceHandler)
	}
	tx := v1.Group("/transactions")
	{
		tx.GET("", m.GetJobTxHashesHandler)
		tx.POST("/request-reward", m.RequestRewardHandler)
		tx.POST("/send-status", m.SendTxStatusHandler)
		tx.POST("/update-status", m.UpdateTxStatusHandler)
	}
	tele := v1.Group("/telemetry")
	{
		tele.GET("/free", m.GetFreeResourcesHandler)
	}
	if debug == true {
		dht := v1.Group("/dht")
		{
			dht.GET("", m.DumpDHTHandler)
			dht.GET("/update", m.ManualDHTUpdateHandler)
		}
		kadDHT := v1.Group("/kad-dht")
		{
			kadDHT.GET("", m.DumpKademliaDHTHandler)
		}
		v1.GET("/ping", m.PingPeerHandler)
		v1.GET("/oldping", m.OldPingPeerHandler)
		v1.GET("/cleanup", m.CleanupPeerHandler)
	}
	p2p := v1.Group("/peers")
	{
		p2p.GET("", m.ListPeersHandler)
		p2p.GET("/dht", m.ListDHTPeersHandler)
		p2p.GET("/kad-dht", m.ListKadDHTPeersHandler)
		p2p.GET("/self", m.SelfPeerInfoHandler)
		p2p.GET("/chat", m.ListChatHandler)
		p2p.GET("/depreq", m.DefaultDepReqPeerHandler)
		p2p.GET("/chat/start", m.StartChatHandler)
		p2p.GET("/chat/join", m.JoinChatHandler)
		p2p.GET("/chat/clear", m.ClearChatHandler)
		p2p.GET("/file", m.ListFileTransferRequestsHandler)
		p2p.GET("/file/send", m.SendFileTransferHandler)
		p2p.GET("/file/accept", m.AcceptFileTransferHandler)
		p2p.GET("/file/clear", m.ClearFileTransferRequestsHandler)
		// peer.GET("/shell", internal.HandleWebSocket)
		// peer.GET("/log", internal.HandleWebSocket)
	}
	return router
}

func WriteMockMetadata(fs afero.Fs) (string, error) {
	path := utils.GetMetadataFilePath()
	metadata := &models.Metadata{
		Name:            "ExampleName",
		UpdateTimestamp: 1625072042,
		Network:         "192.168.1.1",
		PublicKey:       "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3",
		NodeID:          "node-123",
		AllowCardano:    true,
		Dashboard:       "http://localhost:3000",
	}
	content, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("could not marshal data into mock metadata: %w", err)
	}
	err = afero.WriteFile(fs, path, content, 0644)
	if err != nil {
		return "", fmt.Errorf("could not write content to mock metadata: %w", err)
	}
	buf := bytes.NewBuffer(content)
	return buf.String(), nil
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

func TestProblemDetails(t *testing.T) {
	router := SetupRouter()

	// for /device/status endpoint
	type statusBody struct {
		IsAvailable bool `json:"is_available" binding:"required,boolean"`
	}

	tests := []struct {
		endpoint string
		body     any
	}{
		{
			endpoint: "onboarding/onboard",
			body:     models.CapacityForNunet{},
		},
		{
			endpoint: "onboarding/resource-config",
			body:     models.ResourceConfig{},
		},
		{
			endpoint: "device/status",
			body:     statusBody{},
		},
		{
			endpoint: "vm/start-default",
			body:     firecracker.DefaultVM{},
		},
		{
			endpoint: "vm/start-custom",
			body:     firecracker.CustomVM{},
		},
		{
			endpoint: "transactions/request-reward",
			body:     tokenomics.ClaimCardanoTokenBody{},
		},
		{
			endpoint: "transactions/send-status",
			body:     models.BlockchainTxStatus{},
		},
		{
			endpoint: "transactions/update-status",
			body:     tokenomics.UpdateTxStatusBody{},
		},
	}
	for _, tt := range tests {
		t.Run("missing request body", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/"+tt.endpoint, bytes.NewBuffer([]byte("")))
			router.ServeHTTP(w, req)

			assert.Equal(t, 400, w.Code)

			var problem ProblemDetail
			err := json.Unmarshal(w.Body.Bytes(), &problem)
			assert.NoError(t, err)
			assert.Equal(t, problem, NewEmptyBodyProblem())
		})
		t.Run("invalid request body", func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.body)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/"+tt.endpoint, bytes.NewBuffer(bodyBytes))
			router.ServeHTTP(w, req)

			assert.Equal(t, 400, w.Code)

			var problem ProblemDetail
			err = json.Unmarshal(w.Body.Bytes(), &problem)
			assert.NoError(t, err)
		})
	}
}

// func TestHandleDeploymentRequest(t *testing.T) {
// 	// Setup the router
// 	router := SetupMockRouter()
//
// 	// Start the mock WebSocket server
// 	mockServer := startMockWebSocketServer()
// 	defer mockServer.Close()
// 	t.Run("SuccessfulWebsocketConnection", func(t *testing.T) {
// 		// Use a WebSocket client to connect to the endpoint
// 		c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/api/v1/run/deploy", nil)
// 		if err != nil {
// 			t.Fatalf("Failed to connect: %v", err)
// 		}
// 		defer c.Close()
//
// 		// Check if the connection is successful
// 		assert.NotNil(t, c)
// 	})
//
// 	t.Run("FailedWebsocketConnection", func(t *testing.T) {
// 		// Simulate a scenario where the WebSocket connection fails
// 		c, _, err := websocket.DefaultDialer.Dial("ws://invalid-endpoint", nil)
// 		assert.Error(t, err)
// 		assert.Nil(t, c)
// 	})
//
// 	t.Run("WebsocketDataTransfer", func(t *testing.T) {
// 		// Use a WebSocket client to connect to the endpoint
// 		c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/api/v1/run/deploy", nil)
// 		if err != nil {
// 			t.Fatalf("Failed to connect: %v", err)
// 		}
// 		defer c.Close()
//
// 		// Send some data and check the response
// 		err = c.WriteMessage(websocket.TextMessage, []byte("test message"))
// 		assert.NoError(t, err)
//
// 		_, p, err := c.ReadMessage()
// 		assert.NoError(t, err)
// 		assert.Equal(t, "test message", string(p))
// 	})
//
// 	t.Run("GinHandlerWebsocketUpgrade", func(t *testing.T) {
// 		// Create a new HTTP server
// 		server := httptest.NewServer(router)
// 		defer server.Close()
//
// 		// Create a new HTTP request
// 		req, err := http.NewRequest("GET", server.URL+"/api/v1/run/deploy", nil)
// 		if err != nil {
// 			t.Fatalf("Failed to create request: %v", err)
// 		}
//
// 		// Set headers to simulate a WebSocket upgrade request
// 		req.Header.Set("Connection", "upgrade")
// 		req.Header.Set("Upgrade", "websocket")
// 		req.Header.Set("Sec-WebSocket-Version", "13")
// 		req.Header.Set("Sec-WebSocket-Key", "some-random-key")
//
// 		// Send the request and get the response
// 		resp, err := http.DefaultClient.Do(req)
// 		if err != nil {
// 			t.Fatalf("Failed to send request: %v", err)
// 		}
// 		defer resp.Body.Close()
//
// 		// Check if the response header indicates a WebSocket upgrade
// 		assert.Equal(t, "websocket", resp.Header.Get("Upgrade"))
// 		assert.Equal(t, "101 Switching Protocols", resp.Status)
// 	})
// }
//
// func TestHandleRequestService(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
// 	// Connect to the in-memory database
// 	mockDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
// 	if err != nil {
// 		t.Errorf("error trying to initialize mock db: %v", err)
// 	}
// 	db.DB = mockDB
//
// 	err = mockDB.AutoMigrate(&models.Libp2pInfo{})
// 	if err != nil {
// 		t.Fatalf("unable to migrate Libp2pInfo table - %v", err)
// 	}
//
// 	// Setup the router
// 	router := SetupMockRouter()
//
// 	// Setup host
// 	privK, pubK, _ := dmslibp2p.GenerateKey(time.Now().Unix())
// 	err = dmslibp2p.SaveNodeInfo(privK, pubK, true, true)
// 	if err != nil {
// 		t.Fatal("Failed to save node info")
// 	}
// 	host, _, err := dmslibp2p.NewHost(ctx, privK, true)
// 	if err != nil {
// 		t.Fatal("Failed to create libp2p host")
// 	}
// 	dmslibp2p.DMSp2pSet(host, nil)
//
// 	t.Run("InvalidPOSTData", func(t *testing.T) {
// 		w := httptest.NewRecorder()
// 		payload := `{
// 			"RequesterWalletAddress": "someInvalidAddress"
// 		}`
// 		req, _ := http.NewRequest("POST", "/api/v1/run/request-service", bytes.NewBufferString(payload))
// 		router.ServeHTTP(w, req)
//
// 		assert.Equal(t, 404, w.Code)
// 		var response map[string]interface{}
// 		json.Unmarshal(w.Body.Bytes(), &response)
// 		assert.Equal(t, "no peers found with matched specs", response["error"])
// 	})
// }
//
// func TestHandleSendStatus(t *testing.T) {
// 	// Open an in-memory SQLite database for testing
// 	mockDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening gorm database", err)
// 	}
//
// 	// Run the migrations to create the necessary tables in the in-memory database
// 	err = mockDB.AutoMigrate(&models.Services{})
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when migrating the database", err)
// 	}
//
// 	db.DB = mockDB
//
// 	router := SetupMockRouter()
//
// 	t.Run("Success", func(t *testing.T) {
// 		w := httptest.NewRecorder()
// 		payload := `{"status": "completed", "tx_hash": "0xabcdef1234567890"}`
// 		req, err := http.NewRequest("POST", "/api/v1/transactions/send-status", bytes.NewBufferString(payload))
// 		if err != nil {
// 			t.Fatalf("Failed to create request: %v", err)
// 		}
// 		router.ServeHTTP(w, req)
//
// 		expectedResponse := `{"message": "transaction status  acknowledged"}`
// 		assert.Equal(t, 200, w.Code)
// 		assert.JSONEq(t, expectedResponse, w.Body.String())
// 	})
//
// 	t.Run("InvalidPayload", func(t *testing.T) {
// 		w := httptest.NewRecorder()
// 		req, _ := http.NewRequest("POST", "/api/v1/transactions/send-status", nil)
// 		router.ServeHTTP(w, req)
//
// 		assert.Equal(t, 400, w.Code)
// 		assert.Contains(t, w.Body.String(), "cannot read payload body")
// 	})
// }
//
// func TestGetJobTxHashes(t *testing.T) {
// 	// Open an in-memory SQLite database for testing
// 	mockDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening gorm database", err)
// 	}
//
// 	db.DB = mockDB
//
// 	// Run the migrations to create the necessary tables in the in-memory database
// 	err = db.DB.AutoMigrate(&models.Services{})
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when migrating the database", err)
// 	}
//
// 	router := SetupMockRouter()
//
// 	// test without required parameter
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/api/v1/transactions", nil)
// 	router.ServeHTTP(w, req)
//
// 	assert.Equal(t, 400, w.Code)
// 	assert.Contains(t, w.Body.String(), "invalid size_done parameter")
//
// 	// test when no records in services table
// 	w = httptest.NewRecorder()
// 	req, _ = http.NewRequest("GET", "/api/v1/transactions?size_done=1", nil)
// 	router.ServeHTTP(w, req)
//
// 	assert.Equal(t, 404, w.Code)
// 	assert.Contains(t, w.Body.String(), "no job deployed to request reward for")
//
// 	// test when there are records in services table
// 	mockHash := utils.RandomString(64)
// 	db.DB.Create(&models.Services{TxHash: mockHash, LogURL: "log.nunet.io", TransactionType: "done"})
//
// 	w = httptest.NewRecorder()
// 	req, err = http.NewRequest("GET", "/api/v1/transactions?size_done=1", nil)
// 	if err != nil {
// 		t.Fatalf("Failed to create request: %v", err)
// 	}
// 	router.ServeHTTP(w, req)
// 	assert.Equal(t, 200, w.Code)
// 	assert.Contains(t, w.Body.String(), mockHash)
// }
//
// func TestCardanoAddressRoute(t *testing.T) {
// 	router := SetupMockRouter()
//
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/api/v1/onboarding/address/new", nil)
// 	router.ServeHTTP(w, req)
//
// 	resp := w.Result()
// 	body, _ := io.ReadAll(resp.Body)
//
// 	assert.Equal(t, 200, resp.StatusCode)
// 	assert.Contains(t, string(body), "address")
// 	assert.Contains(t, string(body), "mnemonic")
//
// 	var jsonMap map[string]interface{}
// 	json.Unmarshal(w.Body.Bytes(), &jsonMap)
//
// 	assert.NotEmpty(t, jsonMap)
// }
//
// func TestEthereumAddressRoute(t *testing.T) {
// 	router := SetupMockRouter()
//
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/api/v1/onboarding/address/new?blockchain=ethereum", nil)
// 	router.ServeHTTP(w, req)
//
// 	resp := w.Result()
// 	body, _ := io.ReadAll(resp.Body)
//
// 	assert.Equal(t, 200, resp.StatusCode)
// 	assert.Contains(t, string(body), "address")
// 	assert.Contains(t, string(body), "private_key")
//
// 	var jsonMap map[string]interface{}
// 	json.Unmarshal(w.Body.Bytes(), &jsonMap)
//
// 	assert.NotEmpty(t, jsonMap)
// }
//
// func TestProvisionedRoute(t *testing.T) {
// 	router := SetupMockRouter()
//
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/api/v1/onboarding/provisioned", nil)
// 	router.ServeHTTP(w, req)
//
// 	assert.Equal(t, 200, w.Code)
// 	assert.Contains(t, w.Body.String(), "cpu")
// 	assert.Contains(t, w.Body.String(), "memory")
// }
//
// func TestOnboardEmptyRequest(t *testing.T) {
// 	expectedResponse := `{"error":"invalid request data"}`
// 	router := SetupMockRouter()
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("POST", "/api/v1/onboarding/onboard", nil)
// 	router.ServeHTTP(w, req)
//
// 	assert.Equal(t, 400, w.Code)
// 	assert.Equal(t, expectedResponse, w.Body.String())
// }
//
// func TestOnboard_CapacityTooLowTooHigh(t *testing.T) {
// 	onboarding.FS = afero.NewMemMapFs()
// 	onboarding.AFS = &afero.Afero{Fs: onboarding.FS}
// 	onboarding.AFS.Mkdir("/etc/nunet", 0777)
// 	expectedCPUResponse := "CPU should be between 10% and 90% of the available CPU"
// 	expectedRAMResponse := "memory should be between 10% and 90% of the available memory"
//
// 	router := SetupMockRouter()
// 	w := httptest.NewRecorder()
//
// 	type TestRequestPayload struct {
// 		Memory      int64  `json:"memory"`
// 		CPU         int64  `json:"cpu"`
// 		Channel     string `json:"channel"`
// 		PaymentAddr string `json:"payment_addr"`
// 		Cardano     bool   `json:"cardano"`
// 	}
// 	totalCpu := library.GetTotalProvisioned().CPU
// 	totalMem := library.GetTotalProvisioned().Memory
//
// 	// test too low CPU (less than 10% of machine resource)
// 	lowCPUTestPayload := TestRequestPayload{
// 		Memory:      int64(float64(totalMem) * 0.5),  // 50% acceptable
// 		CPU:         int64(float64(totalCpu) * 0.05), // 5% unacceptable
// 		Channel:     "nunet-test",
// 		PaymentAddr: "addr1q99z75su8d8w0jv6drfnr3tuyycflcg4pqpvpnvfzlmmdl7m4nxzjpxhvx477ruhswnrkuqju0kyhx4mvwr0geqyfass7rwta8",
// 		Cardano:     false,
// 	}
// 	jsonPayload, _ := json.Marshal(lowCPUTestPayload)
// 	req, _ := http.NewRequest("POST", "/api/v1/onboarding/onboard", bytes.NewBuffer((jsonPayload)))
// 	router.ServeHTTP(w, req)
// 	assert.Equal(t, 400, w.Code)
// 	assert.Contains(t, w.Body.String(), expectedCPUResponse)
//
// 	// test too high CPU (more than 90% of machine resource)
// 	highCPUTestPayload := TestRequestPayload{
// 		Memory:      int64(float64(totalMem) * 0.5),  // 50% acceptable
// 		CPU:         int64(float64(totalCpu) * 0.95), // 95% unacceptable
// 		Channel:     "nunet-test",
// 		PaymentAddr: "addr1q99z75su8d8w0jv6drfnr3tuyycflcg4pqpvpnvfzlmmdl7m4nxzjpxhvx477ruhswnrkuqju0kyhx4mvwr0geqyfass7rwta8",
// 		Cardano:     false,
// 	}
// 	jsonPayload, _ = json.Marshal(highCPUTestPayload)
// 	req, _ = http.NewRequest("POST", "/api/v1/onboarding/onboard", bytes.NewBuffer((jsonPayload)))
// 	router.ServeHTTP(w, req)
// 	assert.Equal(t, 400, w.Code)
// 	assert.Contains(t, w.Body.String(), expectedCPUResponse)
//
// 	// test too low memory (less than 10% of machine resource)
// 	lowRAMTestPayload := TestRequestPayload{
// 		Memory:      int64(float64(totalMem) * 0.05), // 5% unacceptable
// 		CPU:         int64(float64(totalCpu) * 0.5),  // 50% acceptable
// 		Channel:     "nunet-test",
// 		PaymentAddr: "addr1q99z75su8d8w0jv6drfnr3tuyycflcg4pqpvpnvfzlmmdl7m4nxzjpxhvx477ruhswnrkuqju0kyhx4mvwr0geqyfass7rwta8",
// 		Cardano:     false,
// 	}
// 	jsonPayload, _ = json.Marshal(lowRAMTestPayload)
// 	req, _ = http.NewRequest("POST", "/api/v1/onboarding/onboard", bytes.NewBuffer((jsonPayload)))
// 	router.ServeHTTP(w, req)
// 	assert.Equal(t, 400, w.Code)
// 	assert.Contains(t, w.Body.String(), expectedRAMResponse)
//
// 	// test too high VPU (more than 90% of machine resource)
// 	highRAMTestPayload := TestRequestPayload{
// 		Memory:      int64(float64(totalMem) * 0.95), // 95% unacceptable
// 		CPU:         int64(float64(totalCpu) * 0.5),  // 50% acceptable
// 		Channel:     "nunet-test",
// 		PaymentAddr: "addr1q99z75su8d8w0jv6drfnr3tuyycflcg4pqpvpnvfzlmmdl7m4nxzjpxhvx477ruhswnrkuqju0kyhx4mvwr0geqyfass7rwta8",
// 		Cardano:     false,
// 	}
// 	jsonPayload, _ = json.Marshal(highRAMTestPayload)
// 	req, _ = http.NewRequest("POST", "/api/v1/onboarding/onboard", bytes.NewBuffer((jsonPayload)))
// 	router.ServeHTTP(w, req)
// 	assert.Equal(t, 400, w.Code)
// 	assert.Contains(t, w.Body.String(), expectedRAMResponse)
//
// }
