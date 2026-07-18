package middleware

import "net/http"

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	userRoleKey contextKey = "user_role"
)

func GetUserID(r *http.Request) (string, bool) {
	id, ok := r.Context().Value(userIDKey).(string)
	return id, ok && id != ""
}

func GetUserRole(r *http.Request) (string, bool) {
	role, ok := r.Context().Value(userRoleKey).(string)
	return role, ok
}
