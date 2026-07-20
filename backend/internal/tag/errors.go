package tag

import "errors"

var (
	ErrNotFound = errors.New("tag not found")
	ErrConflict = errors.New("tag already exists")
)
