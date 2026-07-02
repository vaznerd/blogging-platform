package user

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type UserRepository interface {
}

type Repository struct {
	db  *pgxpool.Pool
	rdb *redis.Client
	log *slog.Logger
}

func NewRepository(db *pgxpool.Pool, rdb *redis.Client, log *slog.Logger) *Repository {
	return &Repository{
		db:  db,
		rdb: rdb,
		log: log,
	}
}
