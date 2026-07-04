package server

import (
	"log/slog"
	"net/http"
	"net/http/pprof"

	"codeberg.org/vaznerd/blogging-platform/internal/auth"
	"codeberg.org/vaznerd/blogging-platform/internal/middleware"
	"codeberg.org/vaznerd/blogging-platform/internal/user"
	"github.com/resend/resend-go/v3"
)

func NewRouter(userService *user.Service, authService *auth.Service, log *slog.Logger, mail *resend.Client) http.Handler {
	mux := http.NewServeMux()
	user.RegisterRoutes(mux, userService, log, mail)
	auth.RegisterRoutes(mux, authService, log, mail)

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			return
		}
	})

	stack := middleware.CreateStack(
		middleware.RecoveryMiddleware(log),
		middleware.LoggingMiddleware(log),
		middleware.CorsMiddleware(),
	)
	return stack(mux)
}
