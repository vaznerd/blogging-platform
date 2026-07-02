package user

import (
	"log/slog"

	"github.com/resend/resend-go/v3"
)

type Service struct {
	repo UserRepository
	log  *slog.Logger
	mail *resend.Client
}

func NewService(repo UserRepository, log *slog.Logger, mail *resend.Client) *Service {
	return &Service{
		repo: repo,
		log:  log,
		mail: mail,
	}
}
