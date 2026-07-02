package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

func CreateStack(mw ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			x := mw[i]
			next = x(next)
		}
		return next
	}
}
