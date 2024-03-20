package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/internal/config"
	library "gitlab.com/nunet/device-management-service/lib"
	"gitlab.com/nunet/device-management-service/libp2p"
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

	_, err := WriteMockMetadata(onboarding.AFS.Fs)
	assert.NoError(t, err)

	router := SetupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/onboarding/metadata", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var metadata *models.MetadataV2
	err = json.Unmarshal(w.Body.Bytes(), &metadata)
	assert.NoError(t, err)
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
		description  string
		query        string
		value        string
		expectedCode int
	}{
		{
			description:  "valid cardano query",
			query:        "?blockchain=",
			value:        "cardano",
			expectedCode: 200,
		},
		{
			description:  "valid ethereum query",
			query:        "?blockchain=",
			value:        "ethereum",
			expectedCode: 200,
		},
		{
			description:  "empty query and value",
			query:        "",
			value:        "",
			expectedCode: 200,
		},
		{
			description:  "query with empty value",
			query:        "?blockchain=",
			value:        "",
			expectedCode: 500,
		},
	}
	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/onboarding/address/new"+tc.query+tc.value, nil)
		router.ServeHTTP(w, req)

		if w.Code != tc.expectedCode {
			t.Errorf("%s: wanted code %d, got %d", tc.description, tc.expectedCode, w.Code)
		}
		if w.Code == 200 {
			var keypair *models.BlockchainAddressPrivKey
			err := json.Unmarshal(w.Body.Bytes(), &keypair)
			if err != nil {
				t.Errorf("could not unmarshal blockchain keypair: %v", err)
			}
			if tc.value == "cardano" || tc.value == "" {
				assert.True(t, keypair.Mnemonic != "", tc.description)
			} else if tc.value == "ethereum" {
				assert.True(t, keypair.PrivateKey != "", tc.description)
			}
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

	config.SetConfig("general.metadata_path", ".")
	onboarding.AFS.Fs = afero.NewMemMapFs()

	// TODO: test cases with invalid onboard parameters
	onboardBody, err := onboardTestBody(0.5)
	if err != nil {
		t.Fatalf("failed to generate onboard request body: %v", err)
	}
	bodyBytes, _ := json.Marshal(&onboardBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/onboarding/onboard", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("wanted 200 code, got %d", w.Code)
		t.Errorf("got response: %s", w.Body.String())
	}

	var (
		avalRes       models.AvailableResources
		nodeInfo      models.Libp2pInfo
		freeRes       models.FreeResources
		elasticTokens []models.ElasticToken
		logbinAuth    models.LogBinAuth
	)
	// assert metadata :165
	res := db.Limit(1).Find(&avalRes)
	if res.RowsAffected == 0 {
		t.Errorf("available resources was not created")
	}
	res = db.Limit(1).Find(&nodeInfo)
	if res.RowsAffected == 0 {
		t.Errorf("node info was not created")
	}
	res = db.Limit(1).Find(&freeRes)
	if res.RowsAffected == 0 {
		t.Errorf("free resources was not created")
	}
	res = db.Limit(1).Find(&elasticTokens)
	if res.RowsAffected == 0 {
		t.Errorf("heartbeat tokens was not created")
	}
	res = db.Limit(1).Find(&logbinAuth)
	if res.RowsAffected == 0 {
		t.Errorf("logbin auth was not created")
	}

	// assert PeerStore :213

	// assert Host :220
	host := libp2p.GetP2P().Host
	if host == nil {
		t.Errorf("wanted initialized host, got nil")
	}

	var metadata *models.MetadataV2
	err = json.Unmarshal(w.Body.Bytes(), &metadata)
	assert.NoError(t, err)
}

func TestOffboardHandler(t *testing.T) {
	// TODO: Shut up loggers
	router := SetupTestRouter()

	config.SetConfig("general.metadata_path", ".")
	onboarding.AFS.Fs = afero.NewMemMapFs()

	// TODO: force query only ignore errors, add test case

	tests := []struct {
		description  string
		onboarded    bool
		query        string
		expectedCode int
	}{
		{
			description:  "onboarded machine without query",
			onboarded:    true,
			query:        "",
			expectedCode: 200,
		},
		{
			description:  "onboarded machine with query, not forced",
			onboarded:    true,
			query:        "?force=false",
			expectedCode: 200,
		},
		{
			description:  "onboarded machine with query, forced",
			onboarded:    true,
			query:        "?force=true",
			expectedCode: 200,
		},
		{
			description:  "not onboarded machine without query",
			onboarded:    false,
			query:        "",
			expectedCode: 500,
		},
		{
			description:  "not onboarded machine with query, not forced",
			onboarded:    false,
			query:        "?force=false",
			expectedCode: 500,
		},
		{
			description:  "not onboarded machine with query, forced",
			onboarded:    false,
			query:        "?force=true",
			expectedCode: 500,
		},
	}
	for _, tc := range tests {
		var (
			p2pInfo models.Libp2pInfo
			avalRes models.AvailableResources
		)

		t.Logf("%s: started test", tc.description)

		db, err := ConnectTestDatabase()
		if err != nil {
			t.Fatalf("failed to connect to database: %v", err)
		}

		if tc.onboarded {
			onboardBody, err := onboardTestBody(0.5)
			if err != nil {
				t.Errorf("generating onboard test body: wanted nil, got %v", err)
			}
			_, err = onboarding.Onboard(context.Background(), *onboardBody)
			if err != nil {
				t.Fatalf("failed to onboard: %v", err)
			}
		}

		t.Logf("%s: offboarding started", tc.description)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/onboarding/offboard"+tc.query, nil)
		router.ServeHTTP(w, req)

		if w.Code != tc.expectedCode {
			t.Errorf("%s: wanted code %d, got %d", tc.description, tc.expectedCode, w.Code)
			t.Errorf("%s: response: %s", tc.description, w.Body.String())
		}
		if w.Code == 200 {
			res := db.Limit(1).Find(&p2pInfo)
			if res.RowsAffected != 0 {
				t.Errorf("record failed to be deleted")
			}

			res = db.Limit(1).Find(&avalRes)
			if res.RowsAffected != 0 {
				t.Errorf("record failed to be deleted")
			}
		} else if w.Code == 400 {
			res := db.Limit(1).Find(&p2pInfo)
			if res.RowsAffected == 0 {
				t.Errorf("record should exist, but got deleted")
			}
			res = db.Limit(1).Find(&avalRes)
			if res.RowsAffected == 0 {
				t.Errorf("record should exist, but got deleted")
			}
		}
		CleanupTestDatabase(db)
	}

}

func TestOnboardStatusHandler(t *testing.T) {
	router := SetupTestRouter()

	config.SetConfig("general.metadata_path", ".")
	onboarding.AFS.Fs = afero.NewMemMapFs()

	tests := []struct {
		description string
		onboarded   bool
		wantCode    int
	}{
		{
			description: "machine onboarded",
			onboarded:   true,
			wantCode:    200,
		},
		{
			description: "machine not onboarded",
			onboarded:   false,
			wantCode:    200,
		},
	}
	for _, tc := range tests {
		db, err := ConnectTestDatabase()
		if err != nil {
			t.Fatalf("failed to connect to database: %v", err)
		}

		if tc.onboarded {
			onboardBody, err := onboardTestBody(0.5)
			if err != nil {
				t.Fatalf("could not generate onboard body: %v", err)
			}
			_, err = onboarding.Onboard(context.Background(), *onboardBody)
			if err != nil {
				t.Fatalf("failed to onboard: %v", err)
			}
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/onboarding/status", nil)
		router.ServeHTTP(w, req)

		if w.Code != tc.wantCode {
			t.Errorf("%s: wanted code %d, got %d", tc.description, tc.wantCode, w.Code)
			t.Errorf("%s: response: %s", tc.description, w.Body.String())
		}
		if w.Code == 200 {
			var status models.OnboardingStatus
			err := json.Unmarshal(w.Body.Bytes(), &status)
			if err != nil {
				t.Errorf("%s: unmarshal error: wanted nil, got %v", tc.description, err)
			}
			assert.Equal(t, tc.onboarded, status.Onboarded)
		}
		CleanupTestDatabase(db)
	}
}

func TestResourceConfigHandler(t *testing.T) {
	router := SetupTestRouter()

	config.SetConfig("general.metadata_path", ".")
	onboarding.AFS.Fs = afero.NewMemMapFs()

	tests := []struct {
		description string
		onboarded   bool
		wantCode    int
	}{
		{
			description: "onboarded with valid body",
			onboarded:   true,
			wantCode:    200,
		},
		{
			description: "not onboarded",
			onboarded:   false,
			wantCode:    500,
		},
	}
	for _, tc := range tests {
		var (
			onboardPercent = 0.5
			configPercent  = 0.6
		)

		t.Logf("%s: starting test case", tc.description)

		db, err := ConnectTestDatabase()
		if err != nil {
			t.Fatalf("failed to connect to database: %v", err)
		}

		if tc.onboarded {
			t.Logf("%s: onboarding device", tc.description)

			// onboard device
			onboardBody, err := onboardTestBody(onboardPercent)
			if err != nil {
				t.Errorf("could not set onboard body: %v", err)
			}
			_, err = onboarding.Onboard(context.Background(), *onboardBody)
			if err != nil {
				t.Fatalf("failed to onboard: %v", err)
			}

			// TODO: store current Memory and CPU info
		}
		capacity, err := onboardTestBody(configPercent)
		if err != nil {
			t.Fatalf("could not create resource config body: %v", err)
		}
		bodyBytes, _ := json.Marshal(capacity)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/onboarding/resource-config", bytes.NewBuffer(bodyBytes))
		router.ServeHTTP(w, req)

		if w.Code != tc.wantCode {
			t.Errorf("%s: wanted response code %d, got %d", tc.description, tc.wantCode, w.Code)
			t.Errorf("%s: response: %s", tc.description, w.Body.String())
		}
		if w.Code == 200 {
			var (
				aval models.AvailableResources
				free models.FreeResources
			)
			res := db.Limit(1).Find(&aval)
			if res.RowsAffected == 0 {
				t.Errorf("could not find available resources table")
			}
			res = db.Limit(1).Find(&free)
			if res.RowsAffected == 0 {
				t.Errorf("could not find free resources table")
			}

			// TODO: compare updated with previous record
		}
		CleanupTestDatabase(db)
	}
}

func onboardTestBody(p float64) (*models.CapacityForNunet, error) {
	resources := library.GetTotalProvisioned()
	avalMem := p * float64(resources.Memory)
	avalCPU := p * resources.CPU
	addr, err := onboarding.CreatePaymentAddress("cardano")
	if err != nil {
		return nil, fmt.Errorf("failed to generate payment address")
	}
	return &models.CapacityForNunet{
		Memory:         int64(avalMem),
		CPU:            int64(avalCPU),
		PaymentAddress: addr.Address,
		Channel:        "nunet-test",
	}, nil
}
