package comment

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Handler struct {
	service Service
	log     *slog.Logger
}

func NewHandler(service Service, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) CreateComment(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: "not implemented"})
}

func (h *Handler) GetComment(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: "not implemented"})
}

func (h *Handler) ListComments(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: "not implemented"})
}

func (h *Handler) UpdateComment(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: "not implemented"})
}

func (h *Handler) DeleteComment(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: "not implemented"})
}
