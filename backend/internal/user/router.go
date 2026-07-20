package user

import (
	"log/slog"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, service Service, log *slog.Logger, authMW func(http.Handler) http.Handler) {
	h := NewHandler(service, log)

	mux.HandleFunc("GET "+RouteGetUser, h.GetUser)

	mux.Handle("GET "+RouteMe, authMW(http.HandlerFunc(h.Me)))
	mux.Handle("PATCH "+RouteMe, authMW(http.HandlerFunc(h.UpdateMe)))
	mux.Handle("DELETE "+RouteMe, authMW(http.HandlerFunc(h.DeleteMe)))
}
