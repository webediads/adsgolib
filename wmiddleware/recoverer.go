package wmiddleware

// The original work was derived from Goji's middleware, source:
// https://github.com/zenazn/goji/tree/master/web/middleware

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/go-chi/chi/middleware"
	"github.com/webediads/adsgolib/wconfig"
	"github.com/webediads/adsgolib/wlog"
)

// Recoverer is a middleware that recovers from panics, logs the panic (and a
// backtrace), and returns a HTTP 500 (Internal Server Error) status if
// possible.
func Recoverer() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {

					environment := wconfig.Config.GetEnvironment()

					envsWithDebug := []string{"dev", "staging"}

					_, found := find(envsWithDebug, environment)
					if found {
						w.Write([]byte(fmt.Sprintf("Panic: %+v", rvr)))
					}

					wlog.GetLogger().Critical(fmt.Sprintf("Panic: %+v", rvr), w, r)

					logEntry := middleware.GetLogEntry(r)
					if logEntry != nil {
						logEntry.Panic(rvr, debug.Stack())
					} else {
						fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
						debug.PrintStack()
					}

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			ctx := r.Context()
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
