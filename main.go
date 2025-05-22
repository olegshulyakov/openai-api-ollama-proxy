package main

import (
	"log"
	"net/http"

	// "os" // os is now used within config.LoadConfig()
	"ollama-openai-proxy/src/config" // Add this import
	"ollama-openai-proxy/src/handlers"
	"ollama-openai-proxy/src/middleware"
)

// healthCheckHandler remains the same
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// NotImplementedHandler handles requests to not implemented methods
func NotImplementedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func main() {
	cfg := config.LoadConfig() // Load configuration

	mux := http.NewServeMux()

	mux.HandleFunc("/", healthCheckHandler)
	mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetVersionHandler(w, r, cfg.Version)
	})
	mux.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetModelsHandler(w, r, cfg.OpenAIBaseURL, cfg.OpenAIAllowedModels)
	})
	mux.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		handlers.ChatHandler(w, r, cfg.OpenAIBaseURL)
	})
	mux.HandleFunc("/api/generate", NotImplementedHandler)
	mux.HandleFunc("/api/pull", NotImplementedHandler)
	mux.HandleFunc("/api/push", NotImplementedHandler)
	mux.HandleFunc("/api/create", NotImplementedHandler)
	mux.HandleFunc("/api/ps", NotImplementedHandler)
	mux.HandleFunc("/api/copy", NotImplementedHandler)
	mux.HandleFunc("/api/delete", NotImplementedHandler)
	mux.HandleFunc("/api/show", NotImplementedHandler)
	mux.HandleFunc("/api/embed", NotImplementedHandler)
	mux.HandleFunc("/api/embeddings", NotImplementedHandler)

	loggedMux := middleware.LoggingMiddleware(mux)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, loggedMux); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
