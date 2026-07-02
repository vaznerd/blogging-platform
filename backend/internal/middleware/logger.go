package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const ReqIDKey contextKey = "req_id"

func LoggingMiddleware(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := uuid.New().String()

			ctx := context.WithValue(r.Context(), ReqIDKey, reqID)
			r = r.WithContext(ctx)

			rw := &ResponseWriter{ResponseWriter: w, Status: http.StatusOK}
			w.Header().Set("X-Request-ID", reqID)
			next.ServeHTTP(rw, r)

			var level slog.Level
			switch {
			case rw.Status >= http.StatusInternalServerError:
				level = slog.LevelError
			case rw.Status >= http.StatusBadRequest:
				level = slog.LevelWarn
			default:
				level = slog.LevelInfo
			}

			log.Log(ctx, level, "http_request",
				slog.String("req_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.Status),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
				slog.String("ip", r.RemoteAddr),
				slog.String("user_agent", r.Header.Get("User-Agent")),
			)
		})
	}
}
