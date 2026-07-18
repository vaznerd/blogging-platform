package post

import "errors"

var (
	ErrNotFound         = errors.New("post not found")
	ErrForbidden        = errors.New("forbidden")
	ErrSlugAlreadyExists = errors.New("slug already exists")
	ErrInvalidStatus    = errors.New("invalid post status")
)
