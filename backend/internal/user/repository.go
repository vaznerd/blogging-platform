package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID              string
	Email           string
	Username        string
	PasswordHash    string
	Bio             string
	AvatarURL       string
	Role            string
	IsEmailVerified bool
}

type Repository interface {
	CreateUser(ctx context.Context, email, username, passwordHash string) (string, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, userID string) (*User, error)
	MarkEmailVerified(ctx context.Context, userID string) error
	UpdatePassword(ctx context.Context, userID string, passwordHash string) error
}

type pgRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &pgRepository{
		db: db,
	}
}

func (r *pgRepository) CreateUser(ctx context.Context, email, username, passwordHash string) (string, error) {
	var id string
	err := r.db.QueryRow(
		ctx,
		`INSERT INTO users (email, username, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		email, username, passwordHash,
	).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", ErrConflict
		}
		return "", err
	}
	return id, nil
}

func (r *pgRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(
		ctx,
		`SELECT id, email, username, password_hash, bio, avatar_url, role, is_email_verified FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.Bio, &u.AvatarURL, &u.Role, &u.IsEmailVerified)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *pgRepository) GetByID(ctx context.Context, userID string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(
		ctx,
		`SELECT id, email, username, password_hash, bio, avatar_url, role, is_email_verified
		 FROM users WHERE id = $1`,
		userID,
	).Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.Bio, &u.AvatarURL, &u.Role, &u.IsEmailVerified)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *pgRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	tag, err := r.db.Exec(
		ctx,
		`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`,
		passwordHash, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgRepository) MarkEmailVerified(ctx context.Context, userID string) error {
	tag, err := r.db.Exec(
		ctx,
		`UPDATE users SET is_email_verified = TRUE WHERE id = $1`,
		userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
