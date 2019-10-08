package middleware

import (
	"net/http"
	"strings"

	chi "github.com/go-chi/chi"
)

// StripBeginningSlashes is a middleware that will match request paths with an
// extra beginnning slash, strip it from the path and continue routing through the mux
func StripBeginningSlashes(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var path string
		rctx := chi.RouteContext(r.Context())
		if rctx.RoutePath != "" {
			path = rctx.RoutePath
		} else {
			path = r.URL.Path
		}
		if len(path) > 1 {
			rctx.RoutePath = "/" + strings.Trim(path, "/")
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
