package server

import (
	"errors"
	"io/fs"
	"log/slog"
	"net/http"

	domain "github.com/davidonium/serverplate/internal/serverplate"
	"github.com/davidonium/serverplate/internal/templates"
)

func addRoutes(
	m *http.ServeMux,
	svcs *Services,
) {
	// register the assets on the root as the last step to avoid conflicts with the home handler
	if svcs.Config.AssetsUseManifest {
		stfs, err := fs.Sub(svcs.Assets.FS, "frontend/dist")
		if err != nil {
			svcs.Logger.Error(
				"failed to create assets filesystem, a 404 will be returned for assets requests",
				slog.Any("err", err),
			)
		} else {
			m.Handle("/static/", http.StripPrefix("/static", http.FileServerFS(stfs)))
		}
	}

	c := chainMiddleware([]MiddlewareFunc{
		viteMiddleware(svcs.Assets),
	})
	app := appMiddleware(svcs.Logger, WebErrorHandler(svcs.Logger, svcs.Config.Debug))

	m.Handle("GET /health", healthHandler())
	m.Handle("GET /api/openapi.json", openapiHandler(svcs.Logger))
	m.Handle("GET /api", c(app(apiDocsHandler())))
	m.Handle("GET /{$}", c(app(homeHandler(svcs.PairStore))))
	m.Handle("GET /stats", c(app(statsHandler(svcs.PairStore))))
	m.Handle("GET /generate", c(app(generateHandler(svcs.Generator))))
	m.Handle("GET /config/stats", c(app(configStatsHandler(svcs.PairStore))))
	m.Handle("GET /buckets", c(app(bucketListHandler(svcs.BucketStore))))
	m.Handle("GET /buckets/{id}", c(app(bucketDetailsHandler(svcs.BucketStore))))
	m.Handle("GET /buckets/create", c(app(bucketCreateHandler())))
	m.Handle("POST /buckets", c(app(bucketCreateSubmitHandler(svcs.BucketStore))))
	m.Handle("POST /buckets/{id}/archive", c(app(bucketArchiveHandler(svcs.BucketStore))))
	m.Handle("POST /buckets/{id}/recover", c(app(bucketRecoverHandler(svcs.BucketStore))))

	m.Handle("/{path...}", c(app(notFoundHandler())))
}

type appHandlerFunc func(http.ResponseWriter, *http.Request) error

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

func WebErrorHandler(logger *slog.Logger, debug bool) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		switch {
		case errors.Is(err, domain.ErrBucketNotFound):
			c := templates.NotFoundPage(templates.NotFoundViewModel{
				Message: "Bucket not found",
			})
			if err := component(w, r, http.StatusNotFound, c); err != nil {
				logger.Error("failure rendering 404 page",
					slog.Any("err", err),
					slog.String("request.uri", r.RequestURI),
				)
			}
		default:
			c := templates.InternalErrorPage(templates.InternalErrorViewModel{
				Err:      err,
				PrintErr: debug,
			})
			if err := component(w, r, http.StatusInternalServerError, c); err != nil {
				logger.Error("failure rendering error page",
					slog.Any("err", err),
					slog.String("request.uri", r.RequestURI),
				)
			}
			return
		}

	}
}
