package middleware

import (
	"net/http"
	"slices"
)

type Middleware func(http.Handler) http.Handler

func CreateStack(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for _, middleware := range slices.Backward(middlewares) {
			next = middleware(next)
		}
		return next
	}
}
