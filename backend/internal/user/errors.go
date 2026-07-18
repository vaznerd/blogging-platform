package user

import "errors"

var (
	ErrNotFound  = errors.New("user not found")
	ErrForbidden = errors.New("forbidden")
)
