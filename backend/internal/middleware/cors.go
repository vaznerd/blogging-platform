package middleware

import (
	"github.com/go-chi/cors"
)

const (
	CORSMaxAge = 300
)

func CorsMiddleware(origin string) Middleware {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{origin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           CORSMaxAge,
	})

	return c.Handler
}
