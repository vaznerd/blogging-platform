package auth

import "errors"

var (
	ErrInvalidToken              = errors.New("invalid token")
	ErrTokenExpired              = errors.New("token expired")
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrEmailAlreadyExists        = errors.New("email already exists")
	ErrUsernameAlreadyExists     = errors.New("username already exists")
	ErrUserDoesNotExist          = errors.New("user does not exist")
	ErrSessionNotFound          = errors.New("session not found")
	ErrSessionExpired           = errors.New("session expired")
	ErrSessionRevoked           = errors.New("session revoked")
	ErrEmailNotVerified         = errors.New("email not verified")
	ErrEmailAlreadyVerified     = errors.New("email already verified")
	ErrVerificationTokenInvalid = errors.New("verification token invalid")
	ErrVerificationTokenExpired = errors.New("verification token expired")
	ErrPasswordResetTokenInvalid = errors.New("password reset token invalid")
	ErrPasswordResetTokenExpired = errors.New("password reset token expired")
)
