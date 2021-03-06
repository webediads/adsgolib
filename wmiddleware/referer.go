package wmiddleware

import (
	"context"
	"net/http"

	"github.com/webediads/adsgolib/wcontext"
)

// Referer implements the check of the request params against their hash
func Referer(contextKeyReferer wcontext.Key) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			ctx := r.Context()

			referer := r.Header.Get("Referer")
			ctx = context.WithValue(r.Context(), contextKeyReferer, referer)

			next.ServeHTTP(w, r.WithContext(ctx))

		}
		return http.HandlerFunc(fn)
	}
}
