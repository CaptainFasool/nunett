package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/internal/config"
	library "gitlab.com/nunet/device-management-service/lib"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding"
	"gitlab.com/nunet/device-management-service/utils"
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
	// TODO: Shut up loggers
	router := SetupTestRouter()
	db, err := ConnectTestDatabase()
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer CleanupTestDatabase(db)

	onboarding.AFS.Fs = afero.NewMemMapFs()

	// TODO: Add checking of metadata path
	config.SetConfig("general.metadata_path", ".")
	err = os.Chmod(config.GetConfig().General.MetadataPath, 0777)
	if err != nil {
		t.Fatalf("failed to set permissions to metadata dir: %v", err)
	}

	tests := []struct {
		description  string
		onboarded    bool
		query        string
		expectedCode int
	}{
		{
			description:  "force query true",
			onboarded:    true,
			query:        "?force=true",
			expectedCode: 200,
		},
		{
			description:  "force query false",
			onboarded:    true,
			query:        "?force=false",
			expectedCode: 200,
		},
		{
			description:  "missing force query",
			onboarded:    true,
			query:        "",
			expectedCode: 200,
		},
		{
			description:  "invalid force query",
			onboarded:    true,
			query:        "?force=foobar",
			expectedCode: 400,
		},
	}
	for _, tc := range tests {
		var (
			p2pInfo models.Libp2pInfo
			avalRes models.AvailableResources
		)
		w := httptest.NewRecorder()
		if tc.onboarded {
			err := os.WriteFile(utils.GetMetadataFilePath(), []byte(testMetadata), 0644)
			if err != nil {
				t.Fatalf("could not write metadata file: %v", err)
			}

			onboardBody, err := onboardTestBody()
			if err != nil {
				t.Errorf("generating onboard test body: wanted nil, got %v", err)
			}

			_, err = onboarding.Onboard(context.Background(), *onboardBody)
			if err != nil {
				t.Fatalf("failed to onboard: %v", err)
			}
		}
		req, _ := http.NewRequest("DELETE", "/api/v1/onboarding/offboard"+tc.query, nil)
		router.ServeHTTP(w, req)
		if w.Code != tc.expectedCode {
			t.Fatalf("%s: wanted code %d, got %d", tc.description, tc.expectedCode, w.Code)
		}
		if w.Code == 200 {
			// assert node shutdown
			host := libp2p.GetP2P().Host
			if host != nil {
				t.Errorf("host should be nil, got %s", host.ID().String())
			}

			// assert libp2p info deletion
			res := db.Limit(1).Find(&p2pInfo)
			if res.RowsAffected != 0 {
				t.Errorf("record failed to be deleted")
			}

			// assert metadata deletion
			_, err = os.Stat(utils.GetMetadataFilePath())

			// assert available resources deletion
			res = db.Limit(1).Find(&avalRes)
			if res.RowsAffected != 0 {
				t.Errorf("record failed to be deleted")
			}
		} else if w.Code == 400 {
			// assert node still initialized
			host := libp2p.GetP2P().Host
			if host == nil {
				t.Errorf("host should have ID, but got nil")
			}

			// assert db records still exist
			res := db.Limit(1).Find(&p2pInfo)
			if res.RowsAffected == 0 {
				t.Errorf("record should exist, but got deleted")
			}
			res = db.Limit(1).Find(&avalRes)
			if res.RowsAffected == 0 {
				t.Errorf("record should exist, but got deleted")
			}
		}
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

func onboardTestBody() (*models.CapacityForNunet, error) {
	resources := library.GetTotalProvisioned()
	avalMem := 0.5 * float64(resources.Memory)
	avalCPU := 0.5 * resources.CPU
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
