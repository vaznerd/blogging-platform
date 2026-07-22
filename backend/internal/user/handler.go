package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"codeberg.org/vaznerd/blogging-platform/internal/middleware"
)

type Handler struct {
	service Service
	log     *slog.Logger
}

func NewHandler(service Service, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

type profileResponse struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	Username        string `json:"username"`
	Bio             string `json:"bio"`
	AvatarURL       string `json:"avatar_url"`
	Role            string `json:"role"`
	IsEmailVerified bool   `json:"is_email_verified"`
}

func toProfileResponse(u *User) profileResponse {
	return profileResponse{
		ID:              u.ID,
		Email:           u.Email,
		Username:        u.Username,
		Bio:             u.Bio,
		AvatarURL:       u.AvatarURL,
		Role:            u.Role,
		IsEmailVerified: u.IsEmailVerified,
	}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "unauthorized"})
		return
	}

	u, err := h.service.GetByID(r.Context(), userID)
	if errors.Is(err, ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "user not found"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to get user", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "internal server error"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toProfileResponse(u))
}

func (h *Handler) GetUser(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: "not implemented"})
}

func (h *Handler) UpdateMe(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: "not implemented"})
}

func (h *Handler) DeleteMe(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: "not implemented"})
}
