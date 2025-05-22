package config

import (
	"os"
	"strings"
)

// AppConfig holds the application configuration.
type AppConfig struct {
	Port                string
	OpenAIBaseURL       string
	OpenAIAllowedModels []string
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() AppConfig {
	port := os.Getenv("PROXY_PORT")
	if port == "" {
		port = "8080" // Default port
	}

	openAIBaseURL := os.Getenv("OPENAI_API_BASE_URL")
	if openAIBaseURL == "" {
		openAIBaseURL = "https://api.openai.com" // Default OpenAI URL
	}

	allowedModelsEnv := os.Getenv("OPENAI_ALLOWED_MODELS")
	var allowedModelsList []string
	if allowedModelsEnv != "" {
		trimmedModels := strings.TrimSpace(allowedModelsEnv)
		if trimmedModels != "" { // Ensure not to split an empty or whitespace-only string
			allowedModelsList = strings.Split(trimmedModels, ",")
			// Trim whitespace from each model ID
			for i, model := range allowedModelsList {
				allowedModelsList[i] = strings.TrimSpace(model)
			}
		} else {
            allowedModelsList = []string{} // Ensure it's an empty list if env var is whitespace
        }
	} else {
        allowedModelsList = []string{} // Ensure it's an empty list if env var is not set
    }


	return AppConfig{
		Port:                port,
		OpenAIBaseURL:       openAIBaseURL,
		OpenAIAllowedModels: allowedModelsList, // This will be an empty slice if not set or empty
	}
}
