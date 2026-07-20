package auth

import (
	"log/slog"
	"net/http"

	"codeberg.org/vaznerd/blogging-platform/internal/middleware"
	"codeberg.org/vaznerd/blogging-platform/internal/user"
	"github.com/resend/resend-go/v3"
)

func RegisterRoutes(
	mux *http.ServeMux,
	service *Service,
	usr user.Service,
	log *slog.Logger,
	mail *resend.Client,
	frontendURL string,
) {
	h := NewHandler(service, usr, log, mail, frontendURL)
	authMW := middleware.Auth(service.ValidateToken, log)

	mux.HandleFunc("POST "+RouteRegister, h.Register)
	mux.HandleFunc("POST "+RouteLogin, h.Login)
	mux.HandleFunc("POST "+RouteRefresh, h.Refresh)
	mux.HandleFunc("POST "+RouteVerifyEmail, h.VerifyEmail)
	mux.HandleFunc("POST "+RouteResendVerification, h.ResendVerification)
	mux.HandleFunc("POST "+RouteForgotPassword, h.ForgotPassword)
	mux.HandleFunc("POST "+RouteResetPassword, h.ResetPassword)

	mux.Handle("POST "+RouteLogout, authMW(http.HandlerFunc(h.Logout)))
}
