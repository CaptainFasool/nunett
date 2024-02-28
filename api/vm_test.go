package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/firecracker"
)

func (h *MockHandler) StartDefaultHandler(c *gin.Context) {
	var body firecracker.DefaultVM
	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "VM started successfully"})
}

func (h *MockHandler) StartCustomHandler(c *gin.Context) {
	var body firecracker.CustomVM
	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "VM started successfully"})
}

// TODO: test it with incorrect bind json
func TestStartCustomHandler(t *testing.T) {
	router := SetupMockRouter()

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
	assert.Contains(t, "VM started successfully", w.Body.String())
}

func TestStartDefaultHandler(t *testing.T) {
	router := SetupMockRouter()

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
	assert.Contains(t, "VM started successfully", w.Body.String())
}
