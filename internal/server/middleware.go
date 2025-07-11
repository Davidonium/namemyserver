package server

import (
	"log/slog"
	"net/http"
	"slices"

	"github.com/davidonium/namemyserver/internal/vite"
)

type MiddlewareFunc func(http.Handler) http.Handler

func viteMiddleware(assets *vite.Assets) MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// inject assets service into the context so that it's available to go-templ components
			ctx := vite.NewContextWithAssets(r.Context(), assets)
			r = r.WithContext(ctx)
			h.ServeHTTP(w, r)
		})
	}
}

// appMiddleware returns an adapter to be able to return errors in http handlers. This way, a centralized place
// for error handling can be set up through the use of the errorHandler argument.
func appMiddleware(
	logger *slog.Logger,
	errorHandler ErrorHandler,
) func(appHandlerFunc) http.Handler {
	return func(h appHandlerFunc) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			if err := h(rw, r); err != nil {
				logger.Error("error serving http request",
					slog.Any("err", err),
					slog.String("request.uri", r.RequestURI),
				)

				errorHandler(rw, r, err)
				return
			}
		})
	}
}

func chainMiddleware(mds []MiddlewareFunc) MiddlewareFunc {
	cloned := slices.Clone(mds)
	slices.Reverse(cloned)

	return func(h http.Handler) http.Handler {
		for _, m := range cloned {
			h = m(h)
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		})
	}
}
