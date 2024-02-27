package api

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetFreeResourcesHandler(t *testing.T) {
	router := SetupMockRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/telemetry/free", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
