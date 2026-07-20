package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Session struct {
	SessionID        int64
	UserID           string
	RefreshTokenHash []byte
	UserAgent        string
	IPAddress        *string
	ExpiresAt        time.Time
	LastUsedAt       time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
}

type VerificationToken struct {
	ID        int64
	UserID    string
	TokenHash []byte
	ExpiresAt time.Time
	CreatedAt time.Time
}

type PasswordResetToken struct {
	ID        int64
	UserID    string
	TokenHash []byte
	ExpiresAt time.Time
	CreatedAt time.Time
}

type RefreshTokenRepository interface {
	CreateSession(
		ctx context.Context,
		userID string,
		refreshTokenHash []byte,
		userAgent string,
		ipAddress string,
		expiresAt time.Time,
	) error
	GetSessionByRefreshTokenHash(ctx context.Context, hash []byte) (*Session, error)
	UpdateLastUsedAt(ctx context.Context, sessionID int64) error
	RevokeSession(ctx context.Context, sessionID int64) error
	RevokeAllUserSessions(ctx context.Context, userID string) error
	DeleteExpiredSessions(ctx context.Context) (int64, error)

	InsertVerificationToken(ctx context.Context, userID string, tokenHash []byte, expiresAt time.Time) error
	GetVerificationToken(ctx context.Context, tokenHash []byte) (*VerificationToken, error)
	DeleteVerificationTokens(ctx context.Context, userID string) error

	InsertPasswordResetToken(ctx context.Context, userID string, tokenHash []byte, expiresAt time.Time) error
	GetPasswordResetToken(ctx context.Context, tokenHash []byte) (*PasswordResetToken, error)
	DeletePasswordResetTokens(ctx context.Context, userID string) error
}

type refreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) RefreshTokenRepository {
	return &refreshTokenRepository{
		db: db,
	}
}

func (r *refreshTokenRepository) CreateSession(
	ctx context.Context,
	userID string,
	refreshTokenHash []byte,
	userAgent string,
	ipAddress string,
	expiresAt time.Time,
) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO sessions (
			user_id,
			refresh_token_hash,
			user_agent,
			ip_address,
			expires_at
		)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID, refreshTokenHash, userAgent, ipAddress, expiresAt,
	)
	return err
}

func (r *refreshTokenRepository) GetSessionByRefreshTokenHash(
	ctx context.Context,
	hash []byte,
) (*Session, error) {
	s := &Session{}
	err := r.db.QueryRow(
		ctx,
		`SELECT
			session_id,
			user_id,
			refresh_token_hash,
			user_agent,
			ip_address,
			expires_at,
			last_used_at,
			revoked_at,
			created_at
		 FROM sessions
		 WHERE refresh_token_hash = $1`,
		hash,
	).Scan(
		&s.SessionID, &s.UserID, &s.RefreshTokenHash, &s.UserAgent,
		&s.IPAddress, &s.ExpiresAt, &s.LastUsedAt, &s.RevokedAt, &s.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *refreshTokenRepository) UpdateLastUsedAt(
	ctx context.Context,
	sessionID int64,
) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE sessions SET last_used_at = NOW() WHERE session_id = $1`,
		sessionID,
	)
	return err
}

func (r *refreshTokenRepository) RevokeSession(
	ctx context.Context,
	sessionID int64,
) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE sessions
		SET revoked_at = NOW()
		WHERE session_id = $1`,
		sessionID,
	)
	return err
}

func (r *refreshTokenRepository) RevokeAllUserSessions(
	ctx context.Context,
	userID string,
) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE sessions
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
	)
	return err
}

func (r *refreshTokenRepository) DeleteExpiredSessions(
	ctx context.Context,
) (int64, error) {
	tag, err := r.db.Exec(
		ctx,
		`DELETE FROM sessions
		WHERE expires_at < NOW() OR revoked_at IS NOT NULL`,
	)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *refreshTokenRepository) InsertVerificationToken(
	ctx context.Context,
	userID string,
	tokenHash []byte,
	expiresAt time.Time,
) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	return err
}

func (r *refreshTokenRepository) GetVerificationToken(
	ctx context.Context,
	tokenHash []byte,
) (*VerificationToken, error) {
	vt := &VerificationToken{}
	err := r.db.QueryRow(
		ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at
		 FROM email_verification_tokens
		 WHERE token_hash = $1 AND expires_at > NOW()`,
		tokenHash,
	).Scan(&vt.ID, &vt.UserID, &vt.TokenHash, &vt.ExpiresAt, &vt.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrVerificationTokenInvalid
	}
	if err != nil {
		return nil, err
	}
	return vt, nil
}

func (r *refreshTokenRepository) DeleteVerificationTokens(
	ctx context.Context,
	userID string,
) error {
	_, err := r.db.Exec(
		ctx,
		`DELETE FROM email_verification_tokens WHERE user_id = $1`,
		userID,
	)
	return err
}

func (r *refreshTokenRepository) InsertPasswordResetToken(
	ctx context.Context,
	userID string,
	tokenHash []byte,
	expiresAt time.Time,
) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	return err
}

func (r *refreshTokenRepository) GetPasswordResetToken(
	ctx context.Context,
	tokenHash []byte,
) (*PasswordResetToken, error) {
	prt := &PasswordResetToken{}
	err := r.db.QueryRow(
		ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at
		 FROM password_reset_tokens
		 WHERE token_hash = $1 AND expires_at > NOW()`,
		tokenHash,
	).Scan(&prt.ID, &prt.UserID, &prt.TokenHash, &prt.ExpiresAt, &prt.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPasswordResetTokenInvalid
	}
	if err != nil {
		return nil, err
	}
	return prt, nil
}

func (r *refreshTokenRepository) DeletePasswordResetTokens(
	ctx context.Context,
	userID string,
) error {
	_, err := r.db.Exec(
		ctx,
		`DELETE FROM password_reset_tokens WHERE user_id = $1`,
		userID,
	)
	return err
}
