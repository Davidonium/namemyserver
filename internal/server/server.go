package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-h/templ"

	"github.com/davidonium/namemyserver/internal/env"
	"github.com/davidonium/namemyserver/internal/vite"
)

type Services struct {
	Logger *slog.Logger
	Config env.Config
	Assets *vite.Assets
}

func New(svcs *Services) *http.Server {
	m := http.NewServeMux()

	addRoutes(m, svcs)

	return &http.Server{
		Addr:              "127.0.0.1:8080", // TODO parameterize
		Handler:           m,
		ReadHeaderTimeout: 3 * time.Second,
	}
}

func encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

func component(w http.ResponseWriter, r *http.Request, status int, c templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := c.Render(r.Context(), buf); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if _, err := w.Write(buf.Bytes()); err != nil {
		return err
	}

	return nil
}
