package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/a-h/templ"

	"github.com/davidonium/namemyserver/internal/env"
	"github.com/davidonium/namemyserver/internal/namemyserver"
	"github.com/davidonium/namemyserver/internal/server/api"
	"github.com/davidonium/namemyserver/internal/vite"
)

type Services struct {
	Logger      *slog.Logger
	Config      env.Config
	Assets      *vite.Assets
	Generator   *namemyserver.Generator
	PairStore   namemyserver.PairStore
	BucketStore namemyserver.BucketStore
}

func New(svcs *Services) *http.Server {
	m := http.NewServeMux()
	addRoutes(m, svcs)

	handlers := api.New(svcs.Generator, svcs.BucketStore)
	strictOptions := api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  api.RequestErrorHandler(),
		ResponseErrorHandlerFunc: api.ErrorHandler(svcs.Logger, svcs.Config.Debug),
	}
	strict := api.NewStrictHandlerWithOptions(handlers, nil, strictOptions)
	apiHandler := api.HandlerFromMuxWithBaseURL(strict, http.NewServeMux(), "/api")

	m.Handle("/api/", apiHandler)

	return &http.Server{
		Addr:              svcs.Config.ListenAddr,
		Handler:           m,
		ReadHeaderTimeout: 3 * time.Second,
	}
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
