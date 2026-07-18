package auth

import (
	"log/slog"
	"net/http"

	"codeberg.org/vaznerd/blogging-platform/internal/middleware"
	"github.com/resend/resend-go/v3"
)

func RegisterRoutes(mux *http.ServeMux, service *Service, log *slog.Logger, mail *resend.Client) {
	h := NewHandler(service, log, mail)
	authMW := middleware.Auth(service.ValidateToken)

	// Public
	mux.HandleFunc("POST "+RouteRegister, h.Register)
	mux.HandleFunc("POST "+RouteLogin, h.Login)
	mux.HandleFunc("POST "+RouteRefresh, h.Refresh)
	mux.HandleFunc("POST "+RouteVerifyEmail, h.VerifyEmail)
	mux.HandleFunc("POST "+RouteResendVerification, h.ResendVerification)
	mux.HandleFunc("POST "+RouteForgotPassword, h.ForgotPassword)
	mux.HandleFunc("POST "+RouteResetPassword, h.ResetPassword)

	// Protected
	mux.Handle("POST "+RouteLogout, authMW(http.HandlerFunc(h.Logout)))
	mux.Handle("GET "+RouteMe, authMW(http.HandlerFunc(h.Me)))
}
