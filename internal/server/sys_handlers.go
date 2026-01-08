package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/davidonium/namemyserver/internal/server/api"
)

func healthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
}

func openapiHandler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the embedded OpenAPI spec
		spec, err := api.GetSwagger()
		if err != nil {
			logger.Error("failed to load OpenAPI spec",
				slog.Any("err", err),
				slog.String("request.uri", r.RequestURI),
			)
			http.Error(w, "Failed to load OpenAPI spec", http.StatusInternalServerError)
			return
		}

		// Serialize to JSON using standard library
		data, err := json.Marshal(spec)
		if err != nil {
			logger.Error("failed to serialize OpenAPI spec",
				slog.Any("err", err),
				slog.String("request.uri", r.RequestURI),
			)
			http.Error(w, "Failed to serialize OpenAPI spec", http.StatusInternalServerError)
			return
		}

		// Set appropriate headers and write response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
}
