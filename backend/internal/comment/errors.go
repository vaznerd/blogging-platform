package comment

import "errors"

var (
	ErrNotFound  = errors.New("comment not found")
	ErrForbidden = errors.New("forbidden")
)
