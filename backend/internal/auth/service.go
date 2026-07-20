package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"codeberg.org/vaznerd/blogging-platform/internal/config"
	"codeberg.org/vaznerd/blogging-platform/internal/user"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost            = 12
	tokenByteLen          = 32
	verificationTokenTTL  = 24 * time.Hour
	passwordResetTokenTTL = 1 * time.Hour
)

type Service struct {
	jwtSecret        string
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	refreshTokenRepo RefreshTokenRepository
	user             user.Service
}

func NewService(cfg *config.JWTConfig, repo RefreshTokenRepository, usr user.Service) *Service {
	return &Service{
		jwtSecret:        cfg.Secret,
		accessTokenTTL:   cfg.AccessTokenTTL,
		refreshTokenTTL:  cfg.RefreshTokenTTL,
		refreshTokenRepo: repo,
		user:             usr,
	}
}

func HashRefreshToken(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}

func ExtractUserIDFromClaims(claims jwt.MapClaims) (string, bool) {
	sub, ok := claims["sub"].(string)
	return sub, ok && sub != ""
}

func ExtractRoleFromClaims(claims jwt.MapClaims) (string, bool) {
	role, ok := claims["role"].(string)
	return role, ok
}

func (s *Service) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

func (s *Service) ComparePassword(hashed, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}

func (s *Service) signToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) GenerateAccessToken(userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(s.accessTokenTTL).Unix(),
		"iat":  time.Now().Unix(),
	}
	return s.signToken(claims)
}

func (s *Service) GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"type": "refresh",
		"exp":  time.Now().Add(s.refreshTokenTTL).Unix(),
		"iat":  time.Now().Unix(),
	}
	return s.signToken(claims)
}

func (s *Service) ValidateToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims["type"] == "refresh" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (s *Service) CreateSession(
	ctx context.Context,
	userID string,
	refreshTokenHash []byte,
	userAgent string,
	ipAddress string,
) error {
	return s.refreshTokenRepo.CreateSession(
		ctx, userID, refreshTokenHash, userAgent, ipAddress, time.Now().Add(s.refreshTokenTTL),
	)
}

func (s *Service) GetSessionByRefreshTokenHash(ctx context.Context, hash []byte) (*Session, error) {
	session, err := s.refreshTokenRepo.GetSessionByRefreshTokenHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if session.RevokedAt != nil {
		return nil, ErrSessionRevoked
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}
	return session, nil
}

func (s *Service) UpdateSessionLastUsed(ctx context.Context, sessionID int64) error {
	return s.refreshTokenRepo.UpdateLastUsedAt(ctx, sessionID)
}

func (s *Service) RevokeSession(ctx context.Context, sessionID int64) error {
	return s.refreshTokenRepo.RevokeSession(ctx, sessionID)
}

func (s *Service) RevokeAllUserSessions(ctx context.Context, userID string) error {
	return s.refreshTokenRepo.RevokeAllUserSessions(ctx, userID)
}

func (s *Service) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	return s.refreshTokenRepo.DeleteExpiredSessions(ctx)
}

func (s *Service) GenerateVerificationToken() (string, error) {
	b := make([]byte, tokenByteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *Service) StoreVerificationToken(ctx context.Context, userID string, token string) error {
	hash := sha256.Sum256([]byte(token))
	return s.refreshTokenRepo.InsertVerificationToken(ctx, userID, hash[:], time.Now().Add(verificationTokenTTL))
}

func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	hash := sha256.Sum256([]byte(token))
	vt, err := s.refreshTokenRepo.GetVerificationToken(ctx, hash[:])
	if err != nil {
		return err
	}

	if verifyErr := s.user.MarkEmailVerified(ctx, vt.UserID); verifyErr != nil {
		return verifyErr
	}

	return s.refreshTokenRepo.DeleteVerificationTokens(ctx, vt.UserID)
}

func (s *Service) GeneratePasswordResetToken() (string, error) {
	b := make([]byte, tokenByteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *Service) StorePasswordResetToken(ctx context.Context, userID string, token string) error {
	hash := sha256.Sum256([]byte(token))
	return s.refreshTokenRepo.InsertPasswordResetToken(ctx, userID, hash[:], time.Now().Add(passwordResetTokenTTL))
}

func (s *Service) DeletePasswordResetTokens(ctx context.Context, userID string) error {
	return s.refreshTokenRepo.DeletePasswordResetTokens(ctx, userID)
}

func (s *Service) VerifyPasswordResetToken(ctx context.Context, token string) (string, error) {
	hash := sha256.Sum256([]byte(token))
	prt, err := s.refreshTokenRepo.GetPasswordResetToken(ctx, hash[:])
	if err != nil {
		return "", err
	}
	if deleteErr := s.refreshTokenRepo.DeletePasswordResetTokens(ctx, prt.UserID); deleteErr != nil {
		return "", deleteErr
	}
	return prt.UserID, nil
}
