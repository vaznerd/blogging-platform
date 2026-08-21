package post

import (
	"log/slog"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, service Service, log *slog.Logger, authMW func(http.Handler) http.Handler) {
	h := NewHandler(service, log)

	mux.HandleFunc("GET "+RoutePosts, h.ListPosts)
	mux.HandleFunc("GET "+RoutePost, h.GetPost)
	mux.HandleFunc("GET "+RoutePostsByAuthor, h.ListPostsByAuthor)

	mux.Handle("POST "+RoutePosts, authMW(http.HandlerFunc(h.CreatePost)))
	mux.Handle("PUT "+RoutePost, authMW(http.HandlerFunc(h.UpdatePost)))
	mux.Handle("DELETE "+RoutePost, authMW(http.HandlerFunc(h.DeletePost)))
}
