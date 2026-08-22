package auth

import "errors"

var (
	ErrInvalidToken              = errors.New("invalid token")
	ErrSessionNotFound           = errors.New("session not found")
	ErrSessionExpired            = errors.New("session expired")
	ErrSessionRevoked            = errors.New("session revoked")
	ErrVerificationTokenInvalid  = errors.New("verification token invalid")
	ErrPasswordResetTokenInvalid = errors.New("password reset token invalid")
)

const (
	ErrInternalServerMsg = "internal server error"
	ErrInvalidBodyMsg    = "invalid request body"
	ErrInvalidRefreshMsg = "invalid or expired refresh token"
)

const (
	MinPasswordLength = 8
)

const DummyBcryptHash = "$2a$12$anEPqngbDcEREqqvh3Cr8uOn9YwD/EzgslHp8wBnI414AM27jiQV2"

type messageResponse struct {
	Message string `json:"message"`
}
