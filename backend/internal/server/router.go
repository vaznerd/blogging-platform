package server

import (
	"log/slog"
	"net/http"
	"net/http/pprof"

	"codeberg.org/vaznerd/blogging-platform/internal/auth"
	"codeberg.org/vaznerd/blogging-platform/internal/category"
	"codeberg.org/vaznerd/blogging-platform/internal/middleware"
	"codeberg.org/vaznerd/blogging-platform/internal/tag"
	"codeberg.org/vaznerd/blogging-platform/internal/user"
	"github.com/resend/resend-go/v3"
)

func NewRouter(
	userService user.Service,
	authService *auth.Service,
	categoryService *category.Service,
	tagService *tag.Service,
	log *slog.Logger,
	mail *resend.Client,
	debug bool,
	corsOrigin string,
	frontendURL string,
) http.Handler {
	mux := http.NewServeMux()
	authMW := middleware.Auth(authService.ValidateToken, log)

	auth.RegisterRoutes(mux, authService, userService, log, mail, frontendURL)
	user.RegisterRoutes(mux, userService, log, authMW)
	category.RegisterRoutes(mux, categoryService, log, authMW)
	tag.RegisterRoutes(mux, tagService, log, authMW)

	if debug {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	}

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.DebugContext(r.Context(), "health: failed to write response", "error", err)
		}
	})

	stack := middleware.CreateStack(
		middleware.RecoveryMiddleware(log),
		middleware.LoggingMiddleware(log),
		middleware.CorsMiddleware(corsOrigin),
	)
	return stack(mux)
}
