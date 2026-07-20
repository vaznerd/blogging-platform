package category

import (
	"log/slog"
	"net/http"
)

type Handler struct {
	service *Service
	log     *slog.Logger
}

func NewHandler(service *Service, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func (h *Handler) CreateCategory(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (h *Handler) GetCategory(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (h *Handler) ListCategories(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (h *Handler) AttachCategoryToPost(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (h *Handler) DetachCategoryFromPost(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}
