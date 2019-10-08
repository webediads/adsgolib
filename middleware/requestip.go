package middleware

import (
	"context"
	"net/http"
)

// RequestIP handles reading the input data and sets the IP of the request in the context
func RequestIP(contextKeyIP key) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// default without proxy
			fromIP := r.RemoteAddr

			// check if the user is behind a proxy
			xForwardedFor := r.Header.Get("X-Forwarded-For")
			if xForwardedFor != "" {
				fromIP = xForwardedFor
			}

			ctx = context.WithValue(r.Context(), contextKeyIP, fromIP)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
