package api

import (
	"encoding/json"
	"net/http"
)

// writeJSON writes data as JSON to the response writer with proper headers.
// If encoding fails, writes a fallback plain text error message.
func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Fallback to plain text if JSON encoding fails
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}
