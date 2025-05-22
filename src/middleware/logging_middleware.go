package middleware

import (
	"log"
	"net/http"
	"strings" // Added for authLog
	"time"
)

// LoggingMiddleware logs details about each incoming request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Log request details
		authHeader := r.Header.Get("Authorization")
		authLog := "NotPresent"
		if authHeader != "" {
			// Basic masking for Bearer token, just show type and if present
			if len(authHeader) > 7 && strings.ToLower(authHeader[:7]) == "bearer " {
				authLog = "BearerPresent"
			} else {
				authLog = "PresentNotBearer"
			}
		}

		log.Printf(
			"Incoming Request: Method=%s URL=%s RemoteAddr=%s UserAgent=%s Authorization=%s",
			r.Method,
			r.URL.String(),
			r.RemoteAddr,
			r.Header.Get("User-Agent"),
			authLog,
		)

		// Serve the request
		next.ServeHTTP(w, r)

		// Log response status and duration (optional, but good practice)
		// To get status code, we'd need to wrap http.ResponseWriter
		// For simplicity in this step, we'll just log after the handler returns
		log.Printf(
			"Request Handled: Method=%s URL=%s Duration=%s",
			r.Method,
			r.URL.String(),
			time.Since(startTime),
		)
	})
}
