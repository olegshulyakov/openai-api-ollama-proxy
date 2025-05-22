package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"ollama-openai-proxy/src/models"
)

// GetVersionHandler handles requests to /api/version.
// Signature changed to accept config values.
func GetVersionHandler(w http.ResponseWriter, r *http.Request, version string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authToken := r.Header.Get("Authorization")
	if authToken == "" {
		http.Error(w, "Unauthorized: Missing Authorization header", http.StatusUnauthorized)
		return
	}

	ollamaResponse := models.OllamaVersionResponse{Version: version}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ollamaResponse); err != nil {
		log.Printf("Error encoding Ollama response: %v", err)
	}
}
