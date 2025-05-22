package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"ollama-openai-proxy/src/models" // Adjust if your module path is different
	"strings"
	"testing"
	"time"
)

// --- NON-STREAMING TESTS ---

func TestChatHandler_NonStreaming_Success(t *testing.T) {
	var openAIRequest models.OpenAIChatRequest
	var receivedAuthHeader string

	mockOpenAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		if receivedAuthHeader != "Bearer testtoken" {
			t.Errorf("Mock OpenAI: Expected Authorization header 'Bearer testtoken', got '%s'", receivedAuthHeader)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&openAIRequest); err != nil {
			t.Errorf("Mock OpenAI: Could not decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if openAIRequest.Stream {
			t.Error("Mock OpenAI: Expected Stream=false, got true")
		}

		respTime := time.Now().Unix()
		resp := models.OpenAIChatResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: respTime,
			Model:   openAIRequest.Model, // Echo back the requested model
			Choices: []models.OpenAIChatChoice{
				{Index: 0, Message: models.OpenAIChatMessage{Role: "assistant", Content: "Hello there!"}, FinishReason: "stop"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockOpenAIServer.Close()

	ollamaReqPayload := models.OllamaChatRequest{
		Model:    "gpt-3.5-turbo-0125",
		Messages: []models.OllamaChatMessage{{Role: "user", Content: "Hi"}},
		Stream:   false,
	}
	reqBytes, _ := json.Marshal(ollamaReqPayload)
	req, _ := http.NewRequest("POST", "/api/chat", bytes.NewBuffer(reqBytes))
	req.Header.Set("Authorization", "Bearer testtoken")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	ChatHandler(rr, req, mockOpenAIServer.URL)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Handler returned wrong status: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}

	if receivedAuthHeader != "Bearer testtoken" {
		t.Errorf("Authorization header not correctly forwarded to OpenAI mock.")
	}

	var actualResp models.OllamaChatResponse
	if err := json.NewDecoder(rr.Body).Decode(&actualResp); err != nil {
		t.Fatalf("Could not decode non-streaming response: %v", err)
	}

	if actualResp.Message.Content != "Hello there!" {
		t.Errorf("Unexpected Message.Content: got '%s' want 'Hello there!'", actualResp.Message.Content)
	}
	if actualResp.Message.Role != "assistant" {
		t.Errorf("Unexpected Message.Role: got '%s' want 'assistant'", actualResp.Message.Role)
	}
	if !actualResp.Done {
		t.Errorf("Expected Done=true, got false")
	}
	if actualResp.Model != ollamaReqPayload.Model {
		t.Errorf("Unexpected Model: got '%s' want '%s'", actualResp.Model, ollamaReqPayload.Model)
	}
}

func TestChatHandler_NonStreaming_OpenAIError(t *testing.T) {
	mockOpenAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorPayload := map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Invalid API key.",
				"type":    "invalid_request_error",
				"param":   nil,
				"code":    "invalid_api_key",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorPayload)
	}))
	defer mockOpenAIServer.Close()

	ollamaReqPayload := models.OllamaChatRequest{
		Model:    "gpt-3.5-turbo",
		Messages: []models.OllamaChatMessage{{Role: "user", Content: "Test"}},
		Stream:   false,
	}
	reqBytes, _ := json.Marshal(ollamaReqPayload)
	req, _ := http.NewRequest("POST", "/api/chat", bytes.NewBuffer(reqBytes))
	req.Header.Set("Authorization", "Bearer invalidtoken")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	ChatHandler(rr, req, mockOpenAIServer.URL)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Fatalf("Handler returned wrong status for OpenAI error: got %v want %v. Body: %s", status, http.StatusUnauthorized, rr.Body.String())
	}

	var actualErrorPayload map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&actualErrorPayload); err != nil {
		t.Fatalf("Could not decode error response: %v. Body: %s", err, rr.Body.String())
	}
	if _, ok := actualErrorPayload["error"]; !ok {
		t.Errorf("Expected error payload to be forwarded, got: %+v", actualErrorPayload)
	}
	errMap, ok := actualErrorPayload["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("Error payload 'error' field is not a map: %+v", actualErrorPayload["error"])
	}
	if errMap["code"] != "invalid_api_key" {
		t.Errorf("Expected error code 'invalid_api_key', got: %s", errMap["code"])
	}
}

func TestChatHandler_MissingAuthHeader(t *testing.T) {
	ollamaReqPayload := models.OllamaChatRequest{
		Model:    "gpt-3.5-turbo",
		Messages: []models.OllamaChatMessage{{Role: "user", Content: "Test"}},
	}
	reqBytes, _ := json.Marshal(ollamaReqPayload)
	req, _ := http.NewRequest("POST", "/api/chat", bytes.NewBuffer(reqBytes))
	// No Authorization header

	rr := httptest.NewRecorder()
	ChatHandler(rr, req, "http://dummyurl") // URL doesn't matter as auth check is first

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status: got %v want %v. Body: %s", status, http.StatusUnauthorized, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Missing Authorization header") {
		t.Errorf("Expected error message about missing auth header, got: %s", rr.Body.String())
	}
}

func TestChatHandler_BadRequestBody(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/chat", bytes.NewBufferString("{not_json"))
	req.Header.Set("Authorization", "Bearer testtoken")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	ChatHandler(rr, req, "http://dummyurl")

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status: got %v want %v. Body: %s", status, http.StatusBadRequest, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Could not decode JSON") {
		t.Errorf("Expected error message about decoding JSON, got: %s", rr.Body.String())
	}
}

func TestChatHandler_MethodNotAllowed(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/chat", nil) // Using GET
	req.Header.Set("Authorization", "Bearer testtoken")

	rr := httptest.NewRecorder()
	ChatHandler(rr, req, "http://dummyurl")

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status: got %v want %v. Body: %s", status, http.StatusMethodNotAllowed, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Method not allowed") {
		t.Errorf("Expected 'Method not allowed' error, got: %s", rr.Body.String())
	}
}

// --- STREAMING TESTS ---

func TestChatHandler_Streaming_Success(t *testing.T) {
	var openAIRequest models.OpenAIChatRequest
	var receivedAuthHeader string

	mockOpenAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		if receivedAuthHeader != "Bearer testtoken" {
			t.Errorf("Mock OpenAI: Auth header mismatch. Got: %s", receivedAuthHeader)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&openAIRequest); err != nil {
			t.Errorf("Mock OpenAI: Could not decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if !openAIRequest.Stream {
			t.Error("Mock OpenAI: Expected Stream=true, got false")
		}
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("Mock OpenAI: Expected Accept: text/event-stream, got: %s", r.Header.Get("Accept"))
		}


		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		modelName := openAIRequest.Model // Use the requested model name in response
		chunks := []string{
			fmt.Sprintf(`data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"%s","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}`, modelName),
			fmt.Sprintf(`data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"%s","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}`, modelName),
			fmt.Sprintf(`data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"%s","choices":[{"index":0,"delta":{"content":" world"},"finish_reason":null}]}`, modelName),
			fmt.Sprintf(`data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"%s","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`, modelName),
			`data: [DONE]`,
		}
		for _, chunk := range chunks {
			_, err := io.WriteString(w, chunk+"\n\n") // SSE spec uses two newlines
			if err != nil {
				t.Errorf("Mock OpenAI: Error writing chunk: %v", err)
				return
			}
			w.(http.Flusher).Flush()
			time.Sleep(5 * time.Millisecond)
		}
	}))
	defer mockOpenAIServer.Close()

	ollamaReqPayload := models.OllamaChatRequest{
		Model:    "gpt-streaming-model",
		Messages: []models.OllamaChatMessage{{Role: "user", Content: "Hi stream"}},
		Stream:   true,
	}
	reqBytes, _ := json.Marshal(ollamaReqPayload)
	req, _ := http.NewRequest("POST", "/api/chat", bytes.NewBuffer(reqBytes))
	req.Header.Set("Authorization", "Bearer testtoken")

	rr := httptest.NewRecorder()
	ChatHandler(rr, req, mockOpenAIServer.URL)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Handler returned wrong status for streaming: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/x-ndjson" {
		t.Errorf("Expected Content-Type application/x-ndjson, got %s", contentType)
	}
	if receivedAuthHeader != "Bearer testtoken" {
		t.Errorf("Authorization header not correctly forwarded to OpenAI mock for streaming.")
	}


	var receivedOllamaChunks []models.OllamaStreamChunk
	scanner := bufio.NewScanner(rr.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 { // Skip empty lines if any
			continue
		}
		var chunk models.OllamaStreamChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			t.Fatalf("Could not unmarshal stream chunk '%s': %v", string(line), err)
		}
		receivedOllamaChunks = append(receivedOllamaChunks, chunk)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading streaming response: %v", err)
	}

	if len(receivedOllamaChunks) != 4 {
		t.Fatalf("Expected 4 Ollama chunks, got %d: %+v", len(receivedOllamaChunks), receivedOllamaChunks)
	}

	// Chunk 0: Role
	if receivedOllamaChunks[0].Message.Role != "assistant" || receivedOllamaChunks[0].Message.Content != "" || receivedOllamaChunks[0].Done {
		t.Errorf("Chunk 0 (role) unexpected: %+v", receivedOllamaChunks[0])
	}
    if receivedOllamaChunks[0].Model != ollamaReqPayload.Model {
        t.Errorf("Chunk 0 Model mismatch: got %s, want %s", receivedOllamaChunks[0].Model, ollamaReqPayload.Model)
    }

	// Chunk 1: "Hello"
	if receivedOllamaChunks[1].Message.Role != "assistant" || receivedOllamaChunks[1].Message.Content != "Hello" || receivedOllamaChunks[1].Done {
		t.Errorf("Chunk 1 ('Hello') unexpected: %+v", receivedOllamaChunks[1])
	}
    if receivedOllamaChunks[1].Model != ollamaReqPayload.Model {
        t.Errorf("Chunk 1 Model mismatch: got %s, want %s", receivedOllamaChunks[1].Model, ollamaReqPayload.Model)
    }

	// Chunk 2: " world"
	if receivedOllamaChunks[2].Message.Role != "assistant" || receivedOllamaChunks[2].Message.Content != " world" || receivedOllamaChunks[2].Done {
		t.Errorf("Chunk 2 (' world') unexpected: %+v", receivedOllamaChunks[2])
	}
    if receivedOllamaChunks[2].Model != ollamaReqPayload.Model {
        t.Errorf("Chunk 2 Model mismatch: got %s, want %s", receivedOllamaChunks[2].Model, ollamaReqPayload.Model)
    }


	// Chunk 3: Final Done chunk
	if !receivedOllamaChunks[3].Done || receivedOllamaChunks[3].Message.Content != "" || receivedOllamaChunks[3].Message.Role != "assistant" {
		t.Errorf("Chunk 3 (final) unexpected: %+v. Expected Done=true, empty content, role assistant.", receivedOllamaChunks[3])
	}
    if receivedOllamaChunks[3].Model != ollamaReqPayload.Model { // Model name should be in final done chunk
        t.Errorf("Chunk 3 (final) Model mismatch: got %s, want %s", receivedOllamaChunks[3].Model, ollamaReqPayload.Model)
    }
}

func TestChatHandler_Streaming_OpenAIReturnsErrorImmediately(t *testing.T) {
	mockOpenAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody models.OpenAIChatRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Bad request body from proxy", http.StatusBadRequest)
			return
		}
		if !reqBody.Stream { // This test is for when stream:true is requested
			t.Errorf("Mock OpenAI: Expected Stream=true from proxy, got false")
		}

		// Simulate OpenAI returning an error immediately (e.g., invalid model for streaming)
		errorPayload := map[string]interface{}{
			"error": map[string]interface{}{
				"message": "The model `some-model` does not support streaming.",
				"type":    "invalid_request_error",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // Or any other 4xx/5xx error
		json.NewEncoder(w).Encode(errorPayload)
	}))
	defer mockOpenAIServer.Close()

	ollamaReqPayload := models.OllamaChatRequest{
		Model:    "some-model",
		Messages: []models.OllamaChatMessage{{Role: "user", Content: "Stream test"}},
		Stream:   true,
	}
	reqBytes, _ := json.Marshal(ollamaReqPayload)
	req, _ := http.NewRequest("POST", "/api/chat", bytes.NewBuffer(reqBytes))
	req.Header.Set("Authorization", "Bearer testtoken")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	ChatHandler(rr, req, mockOpenAIServer.URL)

	if status := rr.Code; status != http.StatusBadRequest { // Should match OpenAI's error code
		t.Fatalf("Handler returned wrong status: got %v want %v. Body: %s", status, http.StatusBadRequest, rr.Body.String())
	}

	var actualErrorPayload map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&actualErrorPayload); err != nil {
		t.Fatalf("Could not decode error response: %v. Body: %s", err, rr.Body.String())
	}
	if _, ok := actualErrorPayload["error"]; !ok {
		t.Errorf("Expected error payload to be forwarded, got: %+v", actualErrorPayload)
	}
	errMap, ok := actualErrorPayload["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("Error payload 'error' field is not a map: %+v", actualErrorPayload["error"])
	}
	expectedMsg := "The model `some-model` does not support streaming."
	if errMap["message"] != expectedMsg {
		t.Errorf("Expected error message '%s', got: '%s'", expectedMsg, errMap["message"])
	}
}


func TestChatHandler_Streaming_OpenAIErrorAfterPartialStream_MalformedJSON(t *testing.T) {
    // This tests when OpenAI sends some valid chunks, then a malformed JSON chunk
    // The handler should process valid chunks and then likely stop or log an error.
    // The client will receive valid chunks up to the error.
    mockOpenAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/event-stream")
        w.Header().Set("Cache-Control", "no-cache")
        w.Header().Set("Connection", "keep-alive")

        flusher := w.(http.Flusher)

        // Valid chunk
        io.WriteString(w, `data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":1700000000,"model":"gpt-test","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}`+"\n\n")
        flusher.Flush()
        time.Sleep(5 * time.Millisecond)

        io.WriteString(w, `data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":1700000000,"model":"gpt-test","choices":[{"index":0,"delta":{"content":"First part. "},"finish_reason":null}]}`+"\n\n")
        flusher.Flush()
        time.Sleep(5 * time.Millisecond)

        // Malformed JSON chunk (e.g., missing closing brace)
        io.WriteString(w, `data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":1700000000,"model":"gpt-test","choices":[{"index":0,"delta":{"content":"Broken..."}`+"\n\n") // No closing }
        flusher.Flush()
        time.Sleep(5 * time.Millisecond)

        // This part might not be reached by the client if the connection breaks due to error
        // io.WriteString(w, `data: [DONE]`+"\n\n")
        // flusher.Flush()
    }))
    defer mockOpenAIServer.Close()

    ollamaReqPayload := models.OllamaChatRequest{Model: "gpt-test", Messages: []models.OllamaChatMessage{{Role: "user", Content: "Test partial stream"}}, Stream: true}
    reqBytes, _ := json.Marshal(ollamaReqPayload)
    req, _ := http.NewRequest("POST", "/api/chat", bytes.NewBuffer(reqBytes))
    req.Header.Set("Authorization", "Bearer testtoken")

    rr := httptest.NewRecorder()
    ChatHandler(rr, req, mockOpenAIServer.URL)

    if status := rr.Code; status != http.StatusOK { // Status OK because headers were already sent
        t.Errorf("Handler returned wrong status: got %v want %v", status, http.StatusOK)
    }

    var receivedOllamaChunks []models.OllamaStreamChunk
    scanner := bufio.NewScanner(rr.Body)
    for scanner.Scan() {
        line := scanner.Bytes()
        if len(line) == 0 { continue }
        var chunk models.OllamaStreamChunk
        // We expect unmarshalling to fail for the malformed chunk, but not for previous ones.
        // The handler's log should show an error for the malformed chunk.
        if err := json.Unmarshal(line, &chunk); err != nil {
            t.Logf("Error unmarshalling expectedly malformed chunk or subsequent data: '%s', error: %v", string(line), err)
            // This test primarily ensures that valid chunks before the error are received.
            // The behavior of the stream *after* a malformed chunk from OpenAI is less defined by this test.
            // The current handler logs the error and continues, which might mean the client sees nothing further or a broken stream.
            break // Stop processing on first error for test simplicity
        }
        receivedOllamaChunks = append(receivedOllamaChunks, chunk)
    }

    // Assert that we received the valid chunks before the error
    if len(receivedOllamaChunks) < 2 {
        t.Errorf("Expected at least 2 valid Ollama chunks before malformed data, got %d: %+v", len(receivedOllamaChunks), receivedOllamaChunks)
    } else {
        if receivedOllamaChunks[0].Message.Role != "assistant" {
            t.Errorf("Chunk 0 (role) unexpected: %+v", receivedOllamaChunks[0])
        }
        if receivedOllamaChunks[1].Message.Content != "First part. " {
            t.Errorf("Chunk 1 ('First part. ') unexpected: %+v", receivedOllamaChunks[1])
        }
    }
    // Further assertions could check logs if a logging mock were injected,
    // or verify that the stream does not contain a "DONE" message if OpenAI didn't send it.
    // For this test, confirming initial chunks are received is the main goal.
}
