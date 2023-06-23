package utils

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetInternalBaseURL(t *testing.T) {
	// Test with a valid internal endpoint
	endpoint, err := GetInternalBaseURL("/swagger/doc.json")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	expected := "http://localhost:9999/swagger/doc.json"
	if endpoint != expected {
		t.Errorf("Expected %s, but got %s", expected, endpoint)
	}

	// Test with an empty internal endpoint
	_, err = GetInternalBaseURL("")
	if err == nil {
		t.Errorf("Expected an error, but got none")
	}
}

func TestMakeInternalRequest(t *testing.T) {
	// Test with a valid request
	body := []byte(`{"version": "0.4.97"}`)
	req := httptest.NewRequest("GET", "/swagger/doc.json", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	resp := MakeInternalRequest(c, "GET", "/swagger/doc.json", body)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	// Test with an invalid internal endpoint
	body = []byte(`{"version": "0.4.97"}`)
	req = httptest.NewRequest("GET", "/swagger/doc.json", bytes.NewBuffer(body))
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	resp = MakeInternalRequest(c, "GET", "", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, but got %d", http.StatusBadRequest, resp.StatusCode)
	}
}
