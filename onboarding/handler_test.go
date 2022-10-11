package onboarding_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/nunet/device-management-service/routes"

	"github.com/stretchr/testify/assert"
)

func TestCreatePaymentAddressRoute(t *testing.T) {
	router := routes.SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/onboarding/address/new", nil)
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
	router := routes.SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/onboarding/provisioned", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "cpu")
	assert.Contains(t, w.Body.String(), "memory")
}

// Following test has extenrnal dependencies. Behavior changes depending on the presence of metadata.json file.
// func TestMetadataRoute(t *testing.T) {
// 	router := routes.SetupRouter()

// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/api/v1/metadata", nil)
// 	router.ServeHTTP(w, req)

// 	assert.NotNil(t, w.Body)
// 	assert.Equal(t, 200, w.Code)
// 	assert.Contains(t, w.Body.String(), "name")
// 	assert.Contains(t, w.Body.String(), "resource")
// 	assert.Contains(t, w.Body.String(), "available")
// 	assert.Contains(t, w.Body.String(), "reserved")
// 	assert.Contains(t, w.Body.String(), "network")
// 	assert.Contains(t, w.Body.String(), "public_key")
// }
