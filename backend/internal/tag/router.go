package tag

import (
	"log/slog"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, service Service, log *slog.Logger, authMW func(http.Handler) http.Handler) {
	h := NewHandler(service, log)

	mux.HandleFunc("GET "+RouteTags, h.ListTags)
	mux.HandleFunc("GET "+RouteTagByName, h.GetTag)

	mux.Handle("POST "+RouteTags, authMW(http.HandlerFunc(h.CreateTag)))
	mux.Handle("DELETE "+RouteTagByName, authMW(http.HandlerFunc(h.DeleteTag)))
	mux.Handle("POST "+RouteTags+"/post", authMW(http.HandlerFunc(h.AttachTagToPost)))
	mux.Handle("DELETE "+RouteTags+"/post", authMW(http.HandlerFunc(h.DetachTagFromPost)))
}
