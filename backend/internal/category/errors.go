package category

import "errors"

var (
	ErrNotFound  = errors.New("category not found")
	ErrConflict  = errors.New("category already exists")
	ErrForbidden = errors.New("forbidden")
)
