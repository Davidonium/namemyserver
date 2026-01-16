package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/davidonium/namemyserver/internal/server/api"
	"github.com/davidonium/namemyserver/internal/templates"
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
		spec, err := api.GetSwagger()
		if err != nil {
			logger.Error("failed to load OpenAPI spec",
				slog.Any("err", err),
				slog.String("request.uri", r.RequestURI),
			)
			http.Error(w, "Failed to load OpenAPI spec", http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(spec)
		if err != nil {
			logger.Error("failed to serialize OpenAPI spec",
				slog.Any("err", err),
				slog.String("request.uri", r.RequestURI),
			)
			http.Error(w, "Failed to serialize OpenAPI spec", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
}

func notFoundHandler() appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		c := templates.NotFoundPage(templates.NotFoundViewModel{
			Message: "Page not found",
		})
		return component(w, r, http.StatusNotFound, c)
	}
}
