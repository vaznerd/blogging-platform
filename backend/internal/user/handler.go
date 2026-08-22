package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

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

type publicProfileResponse struct {
	Username  string    `json:"username"`
	Bio       string    `json:"bio"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

type updateMeRequest struct {
	Username  *string `json:"username"`
	Bio       *string `json:"bio"`
	AvatarURL *string `json:"avatar_url"`
}

type deleteUserRequest struct {
	Password string `json:"password"`
}

const (
	msgUnauthorized        = "unauthorized"
	msgUserNotFound        = "user not found"
	msgInternalServerError = "internal server error"
)

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

func toPublicProfileResponse(u *User) publicProfileResponse {
	return publicProfileResponse{
		Username:  u.Username,
		Bio:       u.Bio,
		AvatarURL: u.AvatarURL,
		CreatedAt: u.CreatedAt,
	}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: msgUnauthorized})
		return
	}

	user, err := h.service.GetByID(r.Context(), userID)
	if errors.Is(err, ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: msgUserNotFound})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to get user", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: msgInternalServerError})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toProfileResponse(user))
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userName := r.PathValue("username")
	user, err := h.service.GetByUserName(r.Context(), userName)
	if errors.Is(err, ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: msgUserNotFound})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to get user", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: msgInternalServerError})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toPublicProfileResponse(user))
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: msgUnauthorized})
		return
	}

	var req updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	user, err := h.service.UpdateMe(r.Context(), userID, req.Username, req.Bio, req.AvatarURL)
	if errors.Is(err, ErrInvalidUsername) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResponse{
			Error: "username must be 3-30 characters and contain only lowercase letters, digits, '_' or '-'",
		})
		return
	}
	if errors.Is(err, ErrInvalidBio) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "bio must be at most 500 characters"})
		return
	}
	if errors.Is(err, ErrInvalidAvatar) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "avatar_url must be at most 512 characters"})
		return
	}
	if errors.Is(err, ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: msgUserNotFound})
		return
	}
	if errors.Is(err, ErrConflict) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "username already taken"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to update user", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: msgInternalServerError})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toProfileResponse(user))
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: msgUnauthorized})
		return
	}

	var req deleteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}
	if req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "password is required"})
		return
	}

	err := h.service.DeleteAccount(r.Context(), userID, req.Password)
	if errors.Is(err, ErrInvalidPassword) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "invalid password"})
		return
	}
	if errors.Is(err, ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: msgUserNotFound})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to delete account", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: msgInternalServerError})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
