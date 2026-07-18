package user

import (
	"log/slog"
	"net/http"

	"github.com/resend/resend-go/v3"
)

type Handler struct {
	service *Service
	log     *slog.Logger
	mail    *resend.Client
}

func NewHandler(service *Service, log *slog.Logger, mail *resend.Client) *Handler {
	return &Handler{
		service: service,
		log:     log,
		mail:    mail,
	}
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (h *Handler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}
