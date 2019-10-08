package middlewares

import (
	"context"
	"fmt"
	"net/http"
)

// Referer implements the check of the request params against their hash
func Referer(contextKeyReferer key) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			ctx := r.Context()

			referer := r.Header.Get("Referer")
			fmt.Println(referer)
			ctx = context.WithValue(r.Context(), contextKeyReferer, referer)

			next.ServeHTTP(w, r.WithContext(ctx))

		}
		return http.HandlerFunc(fn)
	}
}
