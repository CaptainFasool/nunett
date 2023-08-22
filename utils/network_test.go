package utils

import (
	"io"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"fmt"

	"github.com/gin-gonic/gin"
)

type Info struct {
	Version string `json: "version"`
}

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
	req := httptest.NewRequest("GET", "/swagger/doc.json", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	resp, err := MakeInternalRequest(c, "GET", "/swagger/doc.json", nil)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	expectedContentType := "application/json; charset=utf-8"
	if resp.Header.Get("Content-Type") != expectedContentType {
		t.Errorf("Expected Content-Type header %s, but got %s", expectedContentType, resp.Header.Get("Content-Type"))
	}

	var info Info
	err = json.Unmarshal(body, &info)
	if err != nil {
		t.Errorf("Error unmarshaling response body: %v", err)
	}

	expectedInfo := "0.4.97"
	fmt.Println(info)
	if expectedInfo != info.Version {
		t.Errorf("Expected key %s, but got %s", expectedInfo, info.Version)
	}

	// Test with an invalid internal endpoint
	req = httptest.NewRequest("GET", "/swagger/doc.json", nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	resp, err = MakeInternalRequest(c, "GET", "", nil)
	if err == nil {
		t.Errorf("Expected an error, but got none")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, but got %d", http.StatusBadRequest, resp.StatusCode)
	}
}
