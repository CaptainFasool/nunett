package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/models"
)

func TestGetFreeResourcesHandler(t *testing.T) {
	router := SetupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/telemetry/free", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var free *models.FreeResources
	err := json.Unmarshal(w.Body.Bytes(), &free)
	assert.NoError(t, err)
}
