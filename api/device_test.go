package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/buger/jsonparser"
	"github.com/spf13/afero"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/telemetry"
)

func TestDeviceStatusHandler(t *testing.T) {
	router := SetupTestRouter()

	config.SetConfig("general.metadata_path", ".")
	libp2p.FS = afero.NewMemMapFs()

	_, err := WriteMockMetadata(libp2p.FS)
	if err != nil {
		t.Fatalf("impossible write test metadata: %v", err)
	}

	tests := []struct {
		description  string
		available    bool
		expectedCode int
		expectedMsg  string
	}{
		{
			description:  "device online",
			available:    true,
			expectedCode: 200,
		},
		{
			description:  "device offline",
			available:    false,
			expectedCode: 200,
		},
	}
	for _, tc := range tests {
		db, err := ConnectTestDatabase()
		if err != nil {
			t.Fatalf("could not connect to database: %v", err)
		}

		priv, pub, err := libp2p.GenerateKey(0)
		if err != nil {
			t.Errorf("failed to generate test key: %v", err)
		}
		err = libp2p.SaveNodeInfo(priv, pub, true, tc.available)
		if err != nil {
			t.Errorf("failed to save node info: %v", err)
		}
		err = telemetry.CalcFreeResAndUpdateDB()
		if err != nil {
			t.Errorf("failed to update free resources: %v", err)
		}
		err = libp2p.RunNode(priv, true, tc.available)
		if err != nil {
			t.Errorf("failed to run node: %v", err)
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/device/status", nil)
		router.ServeHTTP(w, req)

		if w.Code != tc.expectedCode {
			t.Errorf("%s: wanted code %d, got %d", tc.description, tc.expectedCode, w.Code)
			t.Errorf("%s: response: %s", tc.description, w.Body.String())
		}
		if w.Code == 200 {
			status, err := jsonparser.GetBoolean(w.Body.Bytes(), "online")
			if err != nil {
				t.Errorf("failed to get boolean value from response body: %v", err)
			}
			if status != tc.available {
				t.Errorf("%s: wanted online set to %t, got %t", tc.description, tc.available, status)
			}
			libp2p.ShutdownNode()
		}
		CleanupTestDatabase(db)
	}
}

func TestChangeDeviceStatusHandler(t *testing.T) {
	router := SetupTestRouter()

	tests := []struct {
		description  string
		payload      []byte
		expectedCode int
		expectedMsg  string
	}{
		{
			description:  "change status to online",
			payload:      []byte(`{"is_available": true}`),
			expectedCode: 200,
			expectedMsg:  "device status set to online",
		},
		{
			description:  "change status to offline",
			payload:      []byte(`{"is_available": false}`),
			expectedCode: 200,
			expectedMsg:  "device status set to offline",
		},
		{
			description:  "invalid payload",
			payload:      []byte(`{"is_available": 350}`),
			expectedCode: 400,
		},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/device/status", bytes.NewBuffer(tc.payload))
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code, tc.description, w.Body.String())
		assert.Contains(t, w.Body.String(), tc.expectedMsg, tc.description)
	}
}
