package tag

import "errors"

var (
	ErrNotFound      = errors.New("tag not found")
	ErrAlreadyExists = errors.New("tag already exists")
)
