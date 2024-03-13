package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/firecracker"
)

// TODO: test it with incorrect bind json
func TestStartCustomHandler(t *testing.T) {
	router := SetupTestRouter()

	body := firecracker.CustomVM{
		KernelImagePath: "/foo/bar",
		FilesystemPath:  "/baz/foo",
		VCPUCount:       1,
		MemSizeMib:      5,
		TapDevice:       "baz",
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/vm/start-custom", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "VM started successfully")
}

func TestStartDefaultHandler(t *testing.T) {
	router := SetupTestRouter()

	body := firecracker.DefaultVM{
		KernelImagePath: "/foo/bar",
		FilesystemPath:  "/baz/foo",
		PublicKey:       "foobar",
		NodeID:          "foobaz",
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/vm/start-default", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "VM started successfully")
}
