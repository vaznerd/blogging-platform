package user

import (
	"log/slog"

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
