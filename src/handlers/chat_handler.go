package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	// "os" // No longer needed for OPENAI_API_BASE_URL here
	"strings"
	"time"

	"ollama-openai-proxy/src/models"
)

// ChatHandler signature changed
func ChatHandler(w http.ResponseWriter, r *http.Request, openAIBaseURL string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// openAIBaseURL is now passed as an argument
	apiURL := openAIBaseURL + "/v1/chat/completions"

	// (Re-inserting the full ChatHandler logic with modification for openAIBaseURL)
	authToken := r.Header.Get("Authorization")
	if authToken == "" {
		http.Error(w, "Unauthorized: Missing Authorization header", http.StatusUnauthorized)
		return
	}

	var ollamaReq models.OllamaChatRequest
	// Read r.Body once
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request: Could not read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close() // Ensure body is closed

	if err := json.Unmarshal(bodyBytes, &ollamaReq); err != nil {
		http.Error(w, "Bad request: Could not decode JSON", http.StatusBadRequest)
		return
	}

	openAIReq := models.OpenAIChatRequest{
		Model:    ollamaReq.Model,
		Messages: make([]models.OpenAIChatMessage, len(ollamaReq.Messages)),
		Stream:   ollamaReq.Stream,
	}
	for i, msg := range ollamaReq.Messages {
		openAIReq.Messages[i] = models.OpenAIChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	reqBodyBytes, err := json.Marshal(openAIReq)
	if err != nil {
		log.Printf("Error marshalling OpenAI request: %v", err)
		http.Error(w, "Failed to marshal OpenAI request", http.StatusInternalServerError)
		return
	}

	httpClient := &http.Client{}
	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBodyBytes)) // apiURL uses the passed-in openAIBaseURL
	if err != nil {
		log.Printf("Error creating request to OpenAI: %v", err)
		http.Error(w, "Failed to create request to OpenAI", http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Authorization", authToken)
	httpReq.Header.Set("Content-Type", "application/json")
	if ollamaReq.Stream {
		httpReq.Header.Set("Accept", "text/event-stream")
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Error making request to OpenAI: %v", err)
		http.Error(w, "Failed to communicate with OpenAI API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the entire response body once
	respBodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("Error reading OpenAI response body: %v", readErr)
		http.Error(w, "Failed to read response from OpenAI", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("OpenAI API Error: Status %d, Body: %s", resp.StatusCode, string(respBodyBytes))
		var errorResp map[string]interface{}
		if json.Unmarshal(respBodyBytes, &errorResp) == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			json.NewEncoder(w).Encode(errorResp)
			return
		}
		http.Error(w, "OpenAI API request failed: "+resp.Status, resp.StatusCode)
		return
	}

	if ollamaReq.Stream {
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		// w.WriteHeader(http.StatusOK) // Implicitly called on first Write/Flush

		flusher, ok := w.(http.Flusher)
		if !ok {
			log.Println("Streaming unsupported: Flusher not available")
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		scanner := bufio.NewScanner(bytes.NewReader(respBodyBytes)) // Use the already read body
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				jsonData := strings.TrimPrefix(line, "data: ")
				if jsonData == "[DONE]" {
					finalChunk := models.OllamaStreamChunk{
						Model:     ollamaReq.Model,
						CreatedAt: time.Now().UTC().Format(time.RFC3339),
						Message:   models.OllamaChatMessage{Role: "assistant", Content: ""},
						Done:      true,
					}
					if err := json.NewEncoder(w).Encode(finalChunk); err != nil {
						log.Printf("Error encoding final stream chunk: %v", err)
						// Connection might be closed, stop processing
						return
					}
					if _, err := w.Write([]byte("\n")); err != nil {
						log.Printf("Error writing newline after final chunk: %v", err)
						return
					}
					flusher.Flush()
					break // Exit loop after [DONE]
				}

				var openAIChunk models.OpenAIStreamChunk
				if err := json.Unmarshal([]byte(jsonData), &openAIChunk); err != nil {
					log.Printf("Error unmarshalling OpenAI stream chunk '%s': %v", jsonData, err)
					continue // Skip malformed chunk
				}

				// Process valid chunks that have content or role
				if len(openAIChunk.Choices) > 0 && (openAIChunk.Choices[0].Delta.Content != "" || openAIChunk.Choices[0].Delta.Role != "") {
					ollamaChunk := models.OllamaStreamChunk{
						Model:     openAIChunk.Model,
						CreatedAt: time.Now().UTC().Format(time.RFC3339),
						Message: models.OllamaChatMessage{
							Role:    openAIChunk.Choices[0].Delta.Role, // Use role from delta
							Content: openAIChunk.Choices[0].Delta.Content,
						},
						Done: false,
					}
					if ollamaChunk.Message.Role == "" {
						ollamaChunk.Message.Role = "assistant" // Default role if not in delta
					}
					if ollamaChunk.Model == "" { // Fallback if model not in chunk
						ollamaChunk.Model = ollamaReq.Model
					}

					if err := json.NewEncoder(w).Encode(ollamaChunk); err != nil {
						log.Printf("Error encoding Ollama stream chunk: %v", err)
						return // Stop streaming if encode fails
					}
					if _, err := w.Write([]byte("\n")); err != nil {
						log.Printf("Error writing newline: %v", err)
						return // Stop streaming if write fails
					}
					flusher.Flush()
				}
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading stream from OpenAI: %v", err)
		}

	} else { // Non-streaming
		var openAIResp models.OpenAIChatResponse
		if err := json.Unmarshal(respBodyBytes, &openAIResp); err != nil {
			log.Printf("Error unmarshalling OpenAI non-stream response: %v. Body: %s", err, string(respBodyBytes))
			http.Error(w, "Failed to decode OpenAI response", http.StatusInternalServerError)
			return
		}

		if len(openAIResp.Choices) == 0 {
			log.Printf("No choices found in OpenAI non-stream response. Body: %s", string(respBodyBytes))
			http.Error(w, "No content choices from OpenAI", http.StatusInternalServerError)
			return
		}

		ollamaResp := models.OllamaChatResponse{
			Model:     openAIResp.Model,
			CreatedAt: time.Unix(openAIResp.Created, 0).UTC().Format(time.RFC3339),
			Message: models.OllamaChatMessage{
				Role:    "assistant", // Default role for response
				Content: openAIResp.Choices[0].Message.Content,
			},
			Done: true,
		}
		if openAIResp.Choices[0].Message.Role != "" {
			ollamaResp.Message.Role = openAIResp.Choices[0].Message.Role
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(ollamaResp); err != nil {
			log.Printf("Error encoding Ollama non-stream response: %v", err)
		}
	}
}
