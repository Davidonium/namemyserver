package server

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/davidonium/namemyserver"
	"github.com/davidonium/namemyserver/internal/templates"
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

	c := chainMiddleware([]MiddlewareFunc{
		viteMiddleware(svcs.Assets),
	})
	app := appMiddleware(svcs.Logger, WebErrorHandler(svcs.Logger, svcs.Config.Debug))

	m.Handle("GET /health", healthHandler())
	m.Handle("GET /{$}", c(app(homeHandler(svcs.PairStore))))
	m.Handle("GET /stats", c(app(statsHandler(svcs.PairStore))))
	m.Handle("POST /generate", c(app(generateHandler(svcs.Generator))))
	m.Handle("GET /config/stats", c(app(configStatsHandler(svcs.PairStore))))

	m.Handle("GET /buckets", c(app(bucketListHandler(svcs.BucketStore))))
	m.Handle("GET /buckets/{id}", c(app(bucketDetailsHandler(svcs.BucketStore))))
	m.Handle("GET /buckets/create", c(app(bucketCreateHandler(svcs.Logger, svcs.Generator))))
	m.Handle("POST /buckets", c(app(bucketCreateSubmitHandler(svcs.BucketStore))))
	m.Handle("POST /buckets/{id}/archive", c(app(bucketArchiveHandler(svcs.BucketStore))))
}

func addAPIRoutes(m *http.ServeMux, svcs *Services) {
	app := appMiddleware(svcs.Logger, APIErrorHandler(svcs.Logger, svcs.Config.Debug))

	m.Handle("POST /generate", app(apiGenerateHandler(svcs.Generator)))
	m.Handle("POST /buckets", app(apiCreateBucketHandler(svcs.BucketStore)))
	m.Handle("GET /buckets", app(apiBucketListHandler(svcs.BucketStore)))
	m.Handle("POST /buckets/{name}/pop", app(apiPopBucketNameHandler(svcs.BucketStore)))
}

type appHandlerFunc func(http.ResponseWriter, *http.Request) error

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

func WebErrorHandler(logger *slog.Logger, debug bool) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, err error) {
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
	}
}

func APIErrorHandler(logger *slog.Logger, debug bool) ErrorHandler {
	return func(w http.ResponseWriter, _ *http.Request, err error) {
		// TODO proper error handling and maybe use the problem detail RFC https://www.rfc-editor.org/rfc/rfc7807
		res := map[string]any{
			"title": "Internal Server Error",
		}
		if debug {
			res["internal.error"] = err.Error()
		}

		if err := encode(w, http.StatusInternalServerError, res); err != nil {
			logger.Error("could not write error response", slog.Any("err", err))
		}
	}
}
