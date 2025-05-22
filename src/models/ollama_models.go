package models

// OllamaModelDetails represents the details structure for an Ollama model.
type OllamaModelDetails struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// OllamaModel represents a single model in the Ollama format.
type OllamaModel struct {
	Name         string             `json:"name"`
	Model        string             `json:"model"`
	ModifiedAt   string             `json:"modified_at"`
	Size         int64              `json:"size"` // OpenAI doesn't provide this, default to 0
	Digest       string             `json:"digest"` // OpenAI doesn't provide this, default to ""
	Details      OllamaModelDetails `json:"details"`
}

// OllamaTagsResponse represents the response for Ollama's /api/tags endpoint.
type OllamaTagsResponse struct {
	Models []OllamaModel `json:"models"`
}
