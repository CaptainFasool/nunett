package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding"
)

func (h *MockHandler) GetMetadataHandler(c *gin.Context) {
	metadata := models.MetadataV2{
		Name:            "metadata",
		UpdateTimestamp: 1633036800,
		Resource: struct {
			MemoryMax int64 `json:"memory_max,omitempty"`
			TotalCore int64 `json:"total_core,omitempty"`
			CPUMax    int64 `json:"cpu_max,omitempty"`
		}{
			MemoryMax: 16000,
			TotalCore: 8,
			CPUMax:    8,
		},
		Available: struct {
			CPU    int64 `json:"cpu,omitempty"`
			Memory int64 `json:"memory,omitempty"`
		}{
			CPU:    4,
			Memory: 8000,
		},
		Reserved: struct {
			CPU    int64 `json:"cpu,omitempty"`
			Memory int64 `json:"memory,omitempty"`
		}{
			CPU:    4,
			Memory: 8000,
		},
		Network:           "mainnet",
		PublicKey:         "abc123xyz",
		NodeID:            "node-001",
		AllowCardano:      true,
		NTXPricePerMinute: 0.1,
	}
	c.JSON(200, metadata)
}

func (h *MockHandler) ProvisionedCapacityHandler(c *gin.Context) {
	prov := models.Provisioned{
		CPU:      3.5,
		Memory:   16384,
		NumCores: 4,
	}
	c.JSON(200, prov)
}

func (h *MockHandler) CreatePaymentAddressHandler(c *gin.Context) {
	wallet := c.DefaultQuery("blockchain", "cardano")
	if wallet != "cardano" && wallet != "ethereum" {
		c.JSON(400, gin.H{"error": "invalid query data"})
		return
	}
	var addr, phrase string
	if wallet == "cardano" {
		addr = "abc123xyz"
		phrase = "barbarbarbar"
	} else {
		addr = "foobar123baz"
		phrase = "bazbazbazbaz"
	}
	key := models.BlockchainAddressPrivKey{
		Address:  addr,
		Mnemonic: phrase,
	}
	c.JSON(200, key)
}

func (h *MockHandler) OnboardHandler(c *gin.Context) {
	capacity := models.CapacityForNunet{
		ServerMode: true,
	}
	err := c.BindJSON(&capacity)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid request data"})
		return
	}
	metadata := models.MetadataV2{
		Name:            "foobar",
		UpdateTimestamp: 1625097600,
		Reserved: struct {
			CPU    int64 `json:"cpu,omitempty"`
			Memory int64 `json:"memory,omitempty"`
		}{
			CPU:    capacity.CPU,
			Memory: capacity.Memory,
		},
		Network:           capacity.Channel,
		PublicKey:         "bazbazbaz",
		NodeID:            "foo123bar",
		AllowCardano:      capacity.Cardano,
		NTXPricePerMinute: capacity.NTXPricePerMinute,
	}
	c.JSON(200, metadata)
}

func (h *MockHandler) OnboardStatusHandler(c *gin.Context) {
	status := models.OnboardingStatus{
		Onboarded:    true,
		Error:        "",
		MachineUUID:  "foo",
		MetadataPath: "/.nunet/metadataV2.json",
		DatabasePath: "/.nunet/nunet.db",
	}
	c.JSON(200, status)
}

func (h *MockHandler) OffboardHandler(c *gin.Context) {
	query := c.DefaultQuery("force", "false")
	force, err := strconv.ParseBool(query)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid query data"})
		return
	}
	var msg string
	if force {
		msg = "forced offboard successfull"
	} else {
		msg = "offboard successfull"
	}
	c.JSON(200, gin.H{"message": msg})
}

func (h *MockHandler) ResourceConfigHandler(c *gin.Context) {
	if c.Request.ContentLength == 0 {
		c.JSON(400, gin.H{"error": "request body is empty"})
		return
	}

	var capacity models.CapacityForNunet
	err := c.BindJSON(&capacity)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid request data"})
		return
	}
	metadata := models.MetadataV2{
		Name:            "foobar",
		UpdateTimestamp: 1625097600,
		Reserved: struct {
			CPU    int64 `json:"cpu,omitempty"`
			Memory int64 `json:"memory,omitempty"`
		}{
			CPU:    capacity.CPU,
			Memory: capacity.Memory,
		},
		Network:           capacity.Channel,
		PublicKey:         "bazbazbaz",
		NodeID:            "foo123bar",
		AllowCardano:      capacity.Cardano,
		NTXPricePerMinute: capacity.NTXPricePerMinute,
	}
	c.JSON(200, metadata)
}

func TestGetMetadataHandler(t *testing.T) {
	var metadata models.MetadataV2
	router := SetupMockRouter()
	tests := []struct {
		description  string
		route        string
		expectedCode int
		expectedBody models.MetadataV2
	}{
		{
			description:  "GET /onboarding/metadata",
			route:        "/api/v1/onboarding/metadata",
			expectedCode: 200,
			expectedBody: metadata,
		},
	}
	for _, tc := range tests {
		req, _ := http.NewRequest("GET", tc.route, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)

		err := json.Unmarshal(w.Body.Bytes(), metadata)
		assert.NoError(t, err)
	}
}

func TestProvisionedCapacityHandler(t *testing.T) {
	var prov *models.Provisioned
	router := SetupMockRouter()
	tests := []struct {
		description  string
		route        string
		expectedCode int
		expectedBody *models.Provisioned
	}{
		{
			description:  "GET /onboarding/provisioned",
			route:        "/api/v1/onboarding/provisioned",
			expectedCode: 200,
			expectedBody: prov,
		},
	}
	for _, tc := range tests {
		req, _ := http.NewRequest("GET", tc.route, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, tc.expectedCode, w.Code, tc.description)

		err := json.Unmarshal(w.Body.Bytes(), &prov)
		if err != nil {
			t.Fatalf("could not unmarshal body to struct: %v", err)
		}
	}
}

func TestCreatePaymentAddressHandler(t *testing.T) {
	router := SetupMockRouter()
	tests := []struct {
		description  string
		route        string
		query        string
		expectedCode int
	}{
		{
			description:  "cardano blockchain query",
			query:        "?blockchain=cardano",
			expectedCode: 200,
		},
		{
			description:  "ethereum blockchain query",
			query:        "?blockchain=ethereum",
			expectedCode: 200,
		},
		{
			description:  "empty blockchain query",
			query:        "",
			expectedCode: 200,
		},
	}
	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/onboarding/address/new"+tc.query, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
	}
}

func TestOnboardHandler(t *testing.T) {
	onboarding.AFS.Fs = afero.NewMemMapFs()
	onboarding.AFS.Mkdir(config.GetConfig().General.MetadataPath, 0777)

	router := SetupMockRouter()

	key, err := onboarding.CreatePaymentAddress("cardano")
	if err != nil {
		t.Errorf("could not create wallet addr: %v", err)
	}
	capacity := models.CapacityForNunet{
		Memory:         4096,
		CPU:            4096,
		Channel:        "nunet-test",
		PaymentAddress: key.Address,
	}
	bodyBytes, _ := json.Marshal(capacity)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/onboarding/onboard", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)

	assert.Equal(t, "", w.Body.String())
	assert.Equal(t, 200, w.Code)
}

func TestOffboardHandler(t *testing.T) {
	router := SetupMockRouter()
	tests := []struct {
		description  string
		route        string
		query        string
		expectedCode int
	}{
		{
			description:  "force query true",
			query:        "?force=true",
			expectedCode: 200,
		},
		{
			description:  "force query false",
			query:        "?force=false",
			expectedCode: 200,
		},
		{
			description:  "invalid force query",
			query:        "?force=foobar",
			expectedCode: 400,
		},
		{
			description:  "missing force query",
			query:        "",
			expectedCode: 200,
		},
	}
	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/onboarding/offboard"+tc.query, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description)
	}
}

func TestOnboardStatusHandler(t *testing.T) {
	router := SetupMockRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/onboarding/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestResourceConfigHandler(t *testing.T) {
	router := SetupMockRouter()
	capacity := models.CapacityForNunet{ServerMode: true}
	bodyBytes, _ := json.Marshal(capacity)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/onboarding/resource-config", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
