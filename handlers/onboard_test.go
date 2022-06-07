package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/nunet/device-management-service/routes"

	"github.com/stretchr/testify/assert"
)

func TestOnboardedRoute(t *testing.T) {
    router := routes.SetupRouter()

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/onboard", nil)
    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), "will show onboarded info")
}


func TestOnboardRoute(t *testing.T) {
    router := routes.SetupRouter()

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/onboard", nil)
    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), "will onboard a new device")
}

func TestProvisionedRoute(t *testing.T) {
	router := routes.SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/provisioned", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "cpu")
	assert.Contains(t, w.Body.String(), "memory")
}
