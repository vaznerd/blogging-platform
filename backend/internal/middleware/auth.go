package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type TokenValidator func(token string) (jwt.MapClaims, error)

const errUnauthorized = "unauthorized"

func Auth(validate TokenValidator, log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{errKey: errUnauthorized})
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			claims, err := validate(tokenStr)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{errKey: errUnauthorized})
				return
			}

			userID, ok := claims["sub"].(string)
			if !ok || userID == "" {
				log.WarnContext(r.Context(), "auth: missing or invalid 'sub' claim")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{errKey: errUnauthorized})
				return
			}

			role, ok := claims["role"].(string)
			if !ok {
				log.WarnContext(r.Context(), "auth: missing or invalid 'role' claim", "sub", userID)
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, userIDKey, userID)
			ctx = context.WithValue(ctx, userRoleKey, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
