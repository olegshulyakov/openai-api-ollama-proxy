package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"ollama-openai-proxy/internal/models" // Assuming module name is ollama-openai-proxy
	"reflect"
	"strings"
	"testing"
	"time"
)

// Helper function to create OllamaModel for expected results
func makeOllamaModel(id string, created int64) models.OllamaModel {
	return models.OllamaModel{
		Name:       id,
		Model:      id,
		ModifiedAt: time.Unix(created, 0).UTC().Format(time.RFC3339),
		Size:       0,
		Digest:     "",
		Details:    models.OllamaModelDetails{}, // Ensure this matches actual default details
	}
}

func TestGetModelsHandler_Success_NoFilter(t *testing.T) {
	// Mock OpenAI Server
	mockOpenAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check auth header
		if r.Header.Get("Authorization") != "Bearer testtoken" {
			t.Errorf("Expected Authorization header 'Bearer testtoken', got '%s'", r.Header.Get("Authorization"))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// Sample OpenAI response
		response := models.OpenAIModelsResponse{
			Object: "list",
			Data: []models.OpenAIModel{
				{ID: "gpt-4", Object: "model", Created: 1687882411, OwnedBy: "openai"},
				{ID: "gpt-3.5-turbo", Object: "model", Created: 1677610602, OwnedBy: "openai"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockOpenAIServer.Close()

	req, err := http.NewRequest("GET", "/api/tags", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer testtoken")

	rr := httptest.NewRecorder()
	// Call GetModelsHandler with mock server's URL and no filter
	GetModelsHandler(rr, req, mockOpenAIServer.URL, []string{})

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}

	// Check response body
	expectedModels := []models.OllamaModel{
		makeOllamaModel("gpt-4", 1687882411),
		makeOllamaModel("gpt-3.5-turbo", 1677610602),
	}
	expectedResponse := models.OllamaTagsResponse{Models: expectedModels}

	var actualResponse models.OllamaTagsResponse
	if err := json.NewDecoder(rr.Body).Decode(&actualResponse); err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	if !reflect.DeepEqual(actualResponse, expectedResponse) {
		t.Errorf("Handler returned unexpected body: got %+v want %+v", actualResponse, expectedResponse)
	}
}

func TestGetModelsHandler_Success_WithFilter(t *testing.T) {
	mockOpenAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer testtoken" {
			t.Errorf("Expected Authorization header 'Bearer testtoken', got '%s'", r.Header.Get("Authorization"))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		response := models.OpenAIModelsResponse{
			Object: "list",
			Data: []models.OpenAIModel{
				{ID: "gpt-4", Object: "model", Created: 1687882411, OwnedBy: "openai"},
				{ID: "gpt-3.5-turbo", Object: "model", Created: 1677610602, OwnedBy: "openai"},
				{ID: "dall-e-3", Object: "model", Created: 1698742000, OwnedBy: "openai"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockOpenAIServer.Close()

	req, err := http.NewRequest("GET", "/api/tags", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer testtoken")

	rr := httptest.NewRecorder()
	// Filter for "gpt-4" and "dall-e-3"
	GetModelsHandler(rr, req, mockOpenAIServer.URL, []string{"gpt-4", "dall-e-3"})

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}

	expectedModels := []models.OllamaModel{
		makeOllamaModel("gpt-4", 1687882411),
		makeOllamaModel("dall-e-3", 1698742000),
	}
	expectedResponse := models.OllamaTagsResponse{Models: expectedModels}
	var actualResponse models.OllamaTagsResponse
	if err := json.NewDecoder(rr.Body).Decode(&actualResponse); err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}
	// Sort slices before comparison if order is not guaranteed by the handler (it is in this case, but good practice)
	// For this specific handler, the order is preserved from the (filtered) OpenAI response.
	if !reflect.DeepEqual(actualResponse, expectedResponse) {
		t.Errorf("Handler returned unexpected body for filtered request: got %+v want %+v", actualResponse, expectedResponse)
	}
}

func TestGetModelsHandler_OpenAIError(t *testing.T) {
	mockOpenAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate an OpenAI API error
		http.Error(w, `{"error": {"message": "Internal Server Error", "type": "server_error"}}`, http.StatusInternalServerError)
	}))
	defer mockOpenAIServer.Close()

	req, err := http.NewRequest("GET", "/api/tags", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer testtoken")

	rr := httptest.NewRecorder()
	GetModelsHandler(rr, req, mockOpenAIServer.URL, []string{})

	// The handler forwards OpenAI's status code
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusInternalServerError, rr.Body.String())
	}
    // Check if the body contains part of the error message (handler might not pass full JSON)
    // Based on current GetModelsHandler, it passes "Failed to fetch models from OpenAI: 500 Internal Server Error"
    // This might change if error propagation is made more sophisticated.
    // For now, we check the status code. A more robust test could check the body if the handler guarantees a specific error format.
    // For example, if GetModelsHandler were to parse the JSON error and return it:
    // var errResp map[string]interface{}
    // if json.NewDecoder(rr.Body).Decode(&errResp) != nil {
    //  t.Fatalf("Could not decode error response: %v", err)
    // }
    // if _, ok := errResp["error"]; !ok {
    //  t.Errorf("Expected error field in JSON response, got: %s", rr.Body.String())
    // }
    // Current behavior: returns plain text error: "Failed to fetch models from OpenAI: " + resp.Status
    expectedErrorSubstring := "Failed to fetch models from OpenAI: 500 Internal Server Error"
    if !strings.Contains(rr.Body.String(), expectedErrorSubstring) {
        t.Errorf("Handler returned unexpected error body: got '%s', expected to contain '%s'", rr.Body.String(), expectedErrorSubstring)
    }
}

func TestGetModelsHandler_MissingAuthHeader(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/tags", nil)
	if err != nil {
		t.Fatal(err)
	}
	// No Authorization header set

	rr := httptest.NewRecorder()
	// The openAIBaseURL and allowedModelsList don't matter as auth should fail first
	GetModelsHandler(rr, req, "http://dummyurl", []string{})

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code for missing auth: got %v want %v. Body: %s", status, http.StatusUnauthorized, rr.Body.String())
	}
	expectedErrorMsg := "Unauthorized: Missing Authorization header"
	if !strings.Contains(rr.Body.String(), expectedErrorMsg) {
		t.Errorf("Handler returned wrong error message: got '%s', expected to contain '%s'", rr.Body.String(), expectedErrorMsg)
	}
}

func TestGetModelsHandler_MethodNotAllowed(t *testing.T) {
	req, err := http.NewRequest("POST", "/api/tags", nil) // Using POST
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer testtoken")

	rr := httptest.NewRecorder()
	GetModelsHandler(rr, req, "http://dummyurl", []string{})

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status code for wrong method: got %v want %v. Body: %s", status, http.StatusMethodNotAllowed, rr.Body.String())
	}
    expectedErrorMsg := "Method not allowed"
    if !strings.Contains(rr.Body.String(), expectedErrorMsg) {
        t.Errorf("Handler returned wrong error message: got '%s', expected to contain '%s'", rr.Body.String(), expectedErrorMsg)
    }
}

func TestGetModelsHandler_EmptyOpenAIResponse(t *testing.T) {
    mockOpenAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Authorization") != "Bearer testtoken" {
            t.Errorf("Expected Authorization header 'Bearer testtoken', got '%s'", r.Header.Get("Authorization"))
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        response := models.OpenAIModelsResponse{ // Empty data
            Object: "list",
            Data:   []models.OpenAIModel{},
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    }))
    defer mockOpenAIServer.Close()

    req, err := http.NewRequest("GET", "/api/tags", nil)
    if err != nil {
        t.Fatal(err)
    }
    req.Header.Set("Authorization", "Bearer testtoken")

    rr := httptest.NewRecorder()
    GetModelsHandler(rr, req, mockOpenAIServer.URL, []string{})

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
    }

    expectedResponse := models.OllamaTagsResponse{Models: []models.OllamaModel{}} // Expect empty models list
    var actualResponse models.OllamaTagsResponse
    if err := json.NewDecoder(rr.Body).Decode(&actualResponse); err != nil {
        t.Fatalf("Could not decode response: %v", err)
    }
    if !reflect.DeepEqual(actualResponse, expectedResponse) {
        t.Errorf("Handler returned unexpected body for empty OpenAI response: got %+v want %+v", actualResponse, expectedResponse)
    }
}

func TestGetModelsHandler_OpenAIUnmarshallingError(t *testing.T) {
    mockOpenAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"object": "list", "data": "not_an_array"}`)) // Malformed JSON
    }))
    defer mockOpenAIServer.Close()

    req, err := http.NewRequest("GET", "/api/tags", nil)
    if err != nil {
        t.Fatal(err)
    }
    req.Header.Set("Authorization", "Bearer testtoken")

    rr := httptest.NewRecorder()
    GetModelsHandler(rr, req, mockOpenAIServer.URL, []string{})

    if status := rr.Code; status != http.StatusInternalServerError {
        t.Errorf("Handler returned wrong status code for OpenAI unmarshal error: got %v want %v. Body: %s", status, http.StatusInternalServerError, rr.Body.String())
    }
    expectedErrorSubstring := "Failed to decode response from OpenAI"
    if !strings.Contains(rr.Body.String(), expectedErrorSubstring) {
        t.Errorf("Handler returned unexpected error body: got '%s', expected to contain '%s'", rr.Body.String(), expectedErrorSubstring)
    }
}
