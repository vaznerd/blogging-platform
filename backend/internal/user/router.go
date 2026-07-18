package user

import (
	"log/slog"
	"net/http"

	"github.com/resend/resend-go/v3"
)

func RegisterRoutes(mux *http.ServeMux, service *Service, log *slog.Logger, mail *resend.Client, validateMW func(http.Handler) http.Handler) {
	h := NewHandler(service, log, mail)

	// Public
	mux.HandleFunc("GET "+RouteGetUser, h.GetUser)

	// Protected
	mux.Handle("PATCH "+RouteUpdateMe, validateMW(http.HandlerFunc(h.UpdateMe)))
	mux.Handle("DELETE "+RouteDeleteMe, validateMW(http.HandlerFunc(h.DeleteMe)))
}
