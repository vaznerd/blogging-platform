package comment

import (
	"log/slog"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, service Service, log *slog.Logger, authMW func(http.Handler) http.Handler) {
	h := NewHandler(service, log)

	mux.HandleFunc("GET "+RoutePostComments, h.ListComments)
	mux.HandleFunc("GET "+RouteComment, h.GetComment)

	mux.Handle("POST "+RoutePostComments, authMW(http.HandlerFunc(h.CreateComment)))
	mux.Handle("PUT "+RouteComment, authMW(http.HandlerFunc(h.UpdateComment)))
	mux.Handle("DELETE "+RouteComment, authMW(http.HandlerFunc(h.DeleteComment)))
}
