package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	// "os" // No longer needed for env vars here
	// "strings" // strings.Split is used in config.LoadConfig, not directly here anymore for allowedModelsList
	"time"
	"ollama-openai-proxy/internal/models"
)

// GetModelsHandler handles requests to /api/tags.
// Signature changed to accept config values.
func GetModelsHandler(w http.ResponseWriter, r *http.Request, openAIBaseURL string, allowedModelsList []string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// openAIBaseURL and allowedModelsList are now passed as arguments
	// Default values are handled by config.LoadConfig()

	authToken := r.Header.Get("Authorization")
	if authToken == "" {
		http.Error(w, "Unauthorized: Missing Authorization header", http.StatusUnauthorized)
		return
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", openAIBaseURL+"/v1/models", nil) // Use passed-in openAIBaseURL
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Failed to create request to OpenAI", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", authToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching models from OpenAI: %v", err)
		http.Error(w, "Failed to fetch models from OpenAI", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("OpenAI API error: %s", resp.Status)
		// TODO: It might be useful to relay more specific error information if possible
		http.Error(w, "Failed to fetch models from OpenAI: "+resp.Status, resp.StatusCode)
		return
	}

	var openAIResp models.OpenAIModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		log.Printf("Error decoding OpenAI response: %v", err)
		http.Error(w, "Failed to decode response from OpenAI", http.StatusInternalServerError)
		return
	}

	var filteredOpenAIModels []models.OpenAIModel
	if len(allowedModelsList) > 0 { // Use passed-in allowedModelsList
		allowedMap := make(map[string]bool)
		for _, modelID := range allowedModelsList {
			// Assuming model IDs are already trimmed by LoadConfig
			allowedMap[modelID] = true 
		}
		for _, model := range openAIResp.Data {
			if _, ok := allowedMap[model.ID]; ok {
				filteredOpenAIModels = append(filteredOpenAIModels, model)
			}
		}
	} else {
		filteredOpenAIModels = openAIResp.Data
	}
	
	ollamaModels := make([]models.OllamaModel, len(filteredOpenAIModels))
	for i, openAIModel := range filteredOpenAIModels {
		ollamaModels[i] = models.OllamaModel{
			Name:       openAIModel.ID,
			Model:      openAIModel.ID,
			ModifiedAt: time.Unix(openAIModel.Created, 0).UTC().Format(time.RFC3339),
			Size:       0, // Not available from OpenAI
			Digest:     "", // Not available from OpenAI
			Details: models.OllamaModelDetails{ // Populate with defaults or leave empty
				ParentModel:       "",
				Format:            "", 
				Family:            "", 
				Families:          nil,
				ParameterSize:     "", 
				QuantizationLevel: "", 
			},
		}
	}

	ollamaResponse := models.OllamaTagsResponse{Models: ollamaModels}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ollamaResponse); err != nil {
		log.Printf("Error encoding Ollama response: %v", err)
	}
}
