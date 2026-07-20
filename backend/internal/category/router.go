package category

import (
	"log/slog"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, service *Service, log *slog.Logger, authMW func(http.Handler) http.Handler) {
	h := NewHandler(service, log)

	mux.HandleFunc("GET "+RouteCategories, h.ListCategories)
	mux.HandleFunc("GET "+RouteCategoryBySlug, h.GetCategory)

	mux.Handle("POST "+RouteCategories, authMW(http.HandlerFunc(h.CreateCategory)))
	mux.Handle("PUT "+RouteCategoryBySlug, authMW(http.HandlerFunc(h.UpdateCategory)))
	mux.Handle("DELETE "+RouteCategoryBySlug, authMW(http.HandlerFunc(h.DeleteCategory)))
	mux.Handle("POST "+RouteCategories+"/post", authMW(http.HandlerFunc(h.AttachCategoryToPost)))
	mux.Handle("DELETE "+RouteCategories+"/post", authMW(http.HandlerFunc(h.DetachCategoryFromPost)))
}
