package wmiddleware

// The original work was derived from Goji's middleware, source:
// https://github.com/zenazn/goji/tree/master/web/middleware

import (
	"context"
	"net/http"

	"git.webedia-group.net/tools/adsgolib/wcontext"
)

// UserAgent stores the useragent in the context for easier use
func UserAgent(contextKeyUserAgent wcontext.Key) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			ctx := r.Context()

			userAgent := r.Header.Get("User-Agent")
			ctx = context.WithValue(r.Context(), contextKeyUserAgent, userAgent)

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
