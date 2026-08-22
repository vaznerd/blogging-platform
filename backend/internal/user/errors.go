package user

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrNotFound        = errors.New("user not found")
	ErrForbidden       = errors.New("forbidden")
	ErrConflict        = errors.New("email or username already exists")
	ErrInvalidInput    = errors.New("invalid input")
	ErrInvalidPassword = errors.New("invalid password")

	ErrInvalidUsername = fmt.Errorf("%w: username", ErrInvalidInput)
	ErrInvalidBio      = fmt.Errorf("%w: bio", ErrInvalidInput)
	ErrInvalidAvatar   = fmt.Errorf("%w: avatar_url", ErrInvalidInput)
)

const (
	MinUsernameLength  = 3
	MaxUsernameLength  = 30
	MaxBioLength       = 500
	MaxAvatarURLLength = 512
)

var usernameRegex = regexp.MustCompile(`^[a-z0-9_-]+$`)
