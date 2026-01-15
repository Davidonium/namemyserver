package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// RequestErrorHandler creates a handler for request parsing errors (malformed JSON, invalid types, etc.).
// These errors result in 400 Bad Request responses with ProblemDetail.
func RequestErrorHandler() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		problem := ProblemDetail{
			Status: http.StatusBadRequest,
			Type:   "request_error",
			Title:  "Invalid request format",
		}
		detail := err.Error()
		problem.Detail = &detail

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		if encErr := json.NewEncoder(w).Encode(problem); encErr != nil {
			// Fallback to plain text if JSON encoding fails
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
	}
}

// ErrorHandler creates a response error handler for the strict server that converts
// Go errors into RFC 7807 ProblemDetail JSON responses.
//
// In debug mode, the detail field includes the error message and type.
// In production mode, a generic error message is returned to avoid leaking implementation details.
func ErrorHandler(logger *slog.Logger, debug bool) func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("error serving API request",
			slog.Any("err", err),
			slog.String("request.uri", r.RequestURI),
			slog.String("request.method", r.Method),
		)

		// Build ProblemDetail response
		problem := ProblemDetail{
			Status: http.StatusInternalServerError,
			Type:   "internal_error",
			Title:  "Internal Server Error",
		}

		// Include debug information in detail field if debug mode enabled
		var detail string
		if debug {
			detail = fmt.Sprintf("Internal error: %v (type: %T)", err, err)
		} else {
			detail = "An unexpected error occurred while processing your request"
		}
		problem.Detail = &detail

		// Write JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		if encErr := json.NewEncoder(w).Encode(problem); encErr != nil {
			logger.Error("failed to encode error response",
				slog.Any("err", encErr),
			)
			// Fallback to plain text if JSON encoding fails
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
	}
}
