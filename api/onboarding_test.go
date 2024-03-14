package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding"
)

var testMetadata string = `
{
 "update_timestamp": 1698332902,
 "resource": {
  "memory_max": 31674,
  "total_core": 16,
  "cpu_max": 67198
 },
 "available": {
  "cpu": 42942,
  "memory": 10340
 },
 "reserved": {
  "cpu": 24256,
  "memory": 21334
 },
 "network": "nunet-team",
 "public_key": "addr_test1vzgxkngaw5dayp8xqzpmajrkm7f7fleyzqrjj8l8fp5e8jcc2p2dk",
 "allow_cardano": true
}`

func TestGetMetadataHandler(t *testing.T) {
	onboarding.AFS.Fs = afero.NewMemMapFs()

	// should I forcefully write the metadata or control it
	// making another API call? or maybe using tables?
	meta, err := WriteMockMetadata(onboarding.AFS.Fs)
	assert.NoError(t, err)

	router := SetupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/onboarding/metadata", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var metadata *models.MetadataV2
	err = json.Unmarshal(w.Body.Bytes(), &metadata)
	assert.NoError(t, err)
	assert.Equal(t, w.Body.String(), meta)
}

func TestProvisionedCapacityHandler(t *testing.T) {
	router := SetupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/onboarding/provisioned", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var prov *models.Provisioned
	err := json.Unmarshal(w.Body.Bytes(), &prov)
	assert.NoError(t, err)
}

func TestCreatePaymentAddressHandler(t *testing.T) {
	router := SetupTestRouter()
	tests := []struct {
		query        string
		value        string
		expectedCode int
	}{
		{
			query:        "?blockchain=",
			value:        "cardano",
			expectedCode: 200,
		},
		{
			query:        "?blockchain=",
			value:        "ethereum",
			expectedCode: 200,
		},
		{
			query:        "?blockchain=",
			value:        "",
			expectedCode: 200,
		},
		{
			query:        "",
			value:        "",
			expectedCode: 200,
		},
	}
	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/onboarding/address/new"+tc.query+tc.value, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code)

		var keypair *models.BlockchainAddressPrivKey
		err := json.Unmarshal(w.Body.Bytes(), &keypair)
		assert.NoError(t, err)
		if tc.value == "cardano" || tc.value == "" {
			assert.True(t, keypair.Mnemonic != "")
		} else if tc.value == "ethereum" {
			assert.True(t, keypair.PrivateKey != "")
		}
	}
}

func TestOnboardHandler(t *testing.T) {
	router := SetupTestRouter()
	db, err := ConnectTestDatabase()
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer CleanupTestDatabase(db)

	capacity := models.CapacityForNunet{
		Memory:         4096,
		CPU:            4096,
		Channel:        "nunet-test",
		PaymentAddress: "foobarfoobarfoobarfoobarfoobar",
	}
	bodyBytes, _ := json.Marshal(capacity)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/onboarding/onboard", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var metadata *models.MetadataV2
	err = json.Unmarshal(w.Body.Bytes(), &metadata)
	assert.NoError(t, err)
}

func TestOffboardHandler(t *testing.T) {
	router := SetupTestRouter()
	db, err := ConnectTestDatabase()
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer CleanupTestDatabase(db)

	tests := []struct {
		description  string
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
	router := SetupTestRouter()
	db, err := ConnectTestDatabase()
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer CleanupTestDatabase(db)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/onboarding/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestResourceConfigHandler(t *testing.T) {
	router := SetupTestRouter()
	db, err := ConnectTestDatabase()
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer CleanupTestDatabase(db)

	capacity := models.CapacityForNunet{ServerMode: true}
	bodyBytes, _ := json.Marshal(capacity)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/onboarding/resource-config", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
