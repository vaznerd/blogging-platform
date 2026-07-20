package user

import "errors"

var (
	ErrNotFound  = errors.New("user not found")
	ErrForbidden = errors.New("forbidden")
	ErrConflict  = errors.New("email or username already exists")
)
