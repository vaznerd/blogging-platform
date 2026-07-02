package auth

import (
	"time"

	"codeberg.org/vaznerd/blogging-platform/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	jwtSecret        string
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	refreshTokenRepo RefreshTokenRepository
}

func NewService(cfg *config.JWTConfig, db *pgxpool.Pool) *Service {
	return &Service{
		jwtSecret:        cfg.Secret,
		accessTokenTTL:   cfg.AccessTokenTTL,
		refreshTokenTTL:  cfg.RefreshTokenTTL,
		refreshTokenRepo: NewRefreshTokenRepository(db),
	}
}
