package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"ollama-openai-proxy/src/models"
)

func TestGetVersionHandler_Success(t *testing.T) {
	mockVersion := "0.0.1"

	req, err := http.NewRequest("GET", "/api/version", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer testtoken")

	rr := httptest.NewRecorder()
	// Call GetVersionHandler with mock server's URL and no filter
	GetVersionHandler(rr, req, mockVersion)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}

	// Check response body
	expectedResponse := models.OllamaVersionResponse{Version: mockVersion}

	var actualResponse models.OllamaVersionResponse
	if err := json.NewDecoder(rr.Body).Decode(&actualResponse); err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	if !reflect.DeepEqual(actualResponse, expectedResponse) {
		t.Errorf("Handler returned unexpected body: got %+v want %+v", actualResponse, expectedResponse)
	}
}

func TestGetVersionHandler_MissingAuthHeader(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/version", nil)
	if err != nil {
		t.Fatal(err)
	}
	// No Authorization header set

	rr := httptest.NewRecorder()
	// The openAIBaseURL and allowedModelsList don't matter as auth should fail first
	GetVersionHandler(rr, req, "dummy")

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code for missing auth: got %v want %v. Body: %s", status, http.StatusUnauthorized, rr.Body.String())
	}
	expectedErrorMsg := "Unauthorized: Missing Authorization header"
	if !strings.Contains(rr.Body.String(), expectedErrorMsg) {
		t.Errorf("Handler returned wrong error message: got '%s', expected to contain '%s'", rr.Body.String(), expectedErrorMsg)
	}
}

func TestGetVersionHandler_MethodNotAllowed(t *testing.T) {
	req, err := http.NewRequest("POST", "/api/version", nil) // Using POST
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer testtoken")

	rr := httptest.NewRecorder()
	GetVersionHandler(rr, req, "dummy")

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status code for wrong method: got %v want %v. Body: %s", status, http.StatusMethodNotAllowed, rr.Body.String())
	}
    expectedErrorMsg := "Method not allowed"
    if !strings.Contains(rr.Body.String(), expectedErrorMsg) {
        t.Errorf("Handler returned wrong error message: got '%s', expected to contain '%s'", rr.Body.String(), expectedErrorMsg)
    }
}
