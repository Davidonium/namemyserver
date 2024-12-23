package server

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/davidonium/namemyserver"
	"github.com/davidonium/namemyserver/internal/templates"
	"github.com/davidonium/namemyserver/internal/vite"
)

func addRoutes(
	m *http.ServeMux,
	svcs *Services,
) {
	// register the assets on the root as the last step to avoid conflicts with the home handler
	if svcs.Config.AssetsUseManifest {
		fileServer, err := fs.Sub(namemyserver.FrontendFS, "frontend/dist")
		if err != nil {
			svcs.Logger.Error("failed to create assets filesystem, a 404 will be returned for assets requests", slog.Any("err", err))
		} else {
			m.Handle("/static/", http.StripPrefix("/static", http.FileServerFS(fileServer)))
		}
	}

	app := appMiddleware(svcs.Logger, svcs.Assets)
	m.Handle("/", app(homeHandler()))
	m.Handle("GET /health", healthHandler())
}

func healthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
}

type appHandlerFunc func(http.ResponseWriter, *http.Request) error

// appMiddleware builds a middleware that injects the assets service for templ components and handles errors
// it allows the user to write handlers that return an error and are handled centraly.
func appMiddleware(logger *slog.Logger, assets *vite.Assets) func(appHandlerFunc) http.Handler {
	return func(h appHandlerFunc) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			// inject assets service into the context so that it's available to go-templ components
			ctx := vite.NewContextWithAssets(r.Context(), assets)
			r = r.WithContext(ctx)

			err := h(rw, r)
			if err != nil {
				logger.Error("error serving http request",
					slog.Any("err", err),
					slog.String("request_uri", r.RequestURI),
				)

				c := templates.InternalErrorPage()

				rw.Header().Set("Content-Type", "text/html")
				rw.WriteHeader(http.StatusInternalServerError)
				if err := c.Render(r.Context(), rw); err != nil {
					logger.Error("failure rendering error page",
						slog.Any("err", err),
						slog.String("request_uri", r.RequestURI),
					)
				}
				return
			}
		})
	}
}
