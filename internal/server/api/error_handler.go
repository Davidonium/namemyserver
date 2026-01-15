package api

import (
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
			Type:   "invalid_request",
			Title:  "Invalid request format",
		}
		detail := err.Error()
		problem.Detail = &detail

		writeJSON(w, http.StatusBadRequest, problem)
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

		problem := ProblemDetail{
			Status: http.StatusInternalServerError,
			Type:   "internal_error",
			Title:  "Internal Server Error",
		}

		var detail string
		if debug {
			detail = fmt.Sprintf("Internal error: %v (type: %T)", err, err)
		} else {
			detail = "An unexpected error occurred while processing your request"
		}
		problem.Detail = &detail

		writeJSON(w, http.StatusInternalServerError, problem)
	}
}
