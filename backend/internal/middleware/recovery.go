package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
)

const (
	errKey            = "error"
	errInternalServer = "internal server error"
)

func RecoveryMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.ErrorContext(
						r.Context(),
						"panic recovered",
						"panic", rec,
						"path", r.URL.Path,
						"method", r.Method,
						"stack", string(debug.Stack()),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(map[string]string{errKey: errInternalServer})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
