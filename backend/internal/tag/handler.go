package tag

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

func (h *Handler) ListTags(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (h *Handler) GetTag(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (h *Handler) CreateTag(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (h *Handler) DeleteTag(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (h *Handler) AttachTagToPost(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (h *Handler) DetachTagFromPost(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}
