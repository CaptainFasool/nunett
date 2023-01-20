package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/onboarding"
)

func SetUpRouter() *gin.Engine {
	router := gin.Default()
	v1 := router.Group("/api/v1")

	onboardingRoute := v1.Group("/onboarding")
	{
		onboardingRoute.GET("/metadata", onboarding.GetMetadata)
		onboardingRoute.POST("/onboard", onboarding.Onboard)
		onboardingRoute.GET("/provisioned", onboarding.ProvisionedCapacity)
		onboardingRoute.GET("/address/new", onboarding.CreatePaymentAddress)
	}
	return router
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
	expectedResponse := `{"error":"request body is empty"}`
	router := SetUpRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/onboarding/onboard", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, w.Body.String(), expectedResponse)
}

func TestOnboard_CardanoCapacityTooLow(t *testing.T) {
	expectedResponse := `{"error":"cardano node requires 10000MB of RAM and 6000MHz CPU"}`
	router := SetUpRouter()
	w := httptest.NewRecorder()

	requestPayload := struct {
		Memory      int64  `json:"memory"`
		CPU         int64  `json:"cpu"`
		Channel     string `json:"channel"`
		PaymentAddr string `json:"payment_addr"`
		Cardano     bool   `json:"cardano"`
	}{
		Memory:      3000,
		CPU:         4000,
		Channel:     "nunet-test",
		PaymentAddr: "mockaddress", // XXX needs to be real address when validation is implemented
		Cardano:     true,
	}
	jsonPayload, _ := json.Marshal(requestPayload)
	req, _ := http.NewRequest("POST", "/api/v1/onboarding/onboard", bytes.NewBuffer((jsonPayload)))
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	assert.Equal(t, w.Body.String(), expectedResponse)
}

// Test that the function creates the metadata struct correctly and returns it
// // when the request is valid.
// func TestOnboard_ValidRequest(t *testing.T) {
// 	FS = afero.NewMemMapFs()
// 	AFS = &afero.Afero{Fs: FS}
// 	AFS.Mkdir("/etc/nunet", 0777)
// 	db.ConnectDatabase()
// 	expectedResponse := `{"error":"cardano node requires 10000MB of RAM and 6000MHz CPU"}`
// 	router := SetUpRouter()
// 	w := httptest.NewRecorder()

// 	requestPayload := struct {
// 		Memory      int64  `json:"memory"`
// 		CPU         int64  `json:"cpu"`
// 		Channel     string `json:"channel"`
// 		PaymentAddr string `json:"payment_addr"`
// 		Cardano     bool   `json:"cardano"`
// 	}{
// 		Memory:      11000,
// 		CPU:         7000,
// 		Channel:     "nunet-test",
// 		PaymentAddr: "mockaddress", // XXX needs to be real address when validation is implemented
// 		Cardano:     true,
// 	}
// 	jsonPayload, _ := json.Marshal(requestPayload)
// 	req, _ := http.NewRequest("POST", "/api/v1/onboarding/onboard", bytes.NewBuffer((jsonPayload)))
// 	router.ServeHTTP(w, req)
// 	assert.Equal(t, 400, w.Code)
// 	assert.Equal(t, w.Body.String(), expectedResponse)
// }
