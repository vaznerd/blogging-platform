package middleware

import (
	"net/http"
	"slices"
)

type Middleware func(http.Handler) http.Handler

func CreateStack(mw ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for _, x := range slices.Backward(mw) {
			next = x(next)
		}
		return next
	}
}
