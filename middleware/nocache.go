package middleware

import (
	"net/http"
	"time"
)

// NoCache adds headers preventing the browser to cache the response
func NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Expires", "Tue, 25 Apr 1978 20:20:00 GMT")
		w.Header().Set("Last-Modified", time.Now().Format(time.RFC1123))
		w.Header().Set("Cache-Control", "private, no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		// for CORS
		if r.Header.Get("Origin") != "" {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		ctx := r.Context()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
