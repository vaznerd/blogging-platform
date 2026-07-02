package auth

import "github.com/jackc/pgx/v5/pgxpool"

type RefreshTokenRepository interface {
}

type refreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}
