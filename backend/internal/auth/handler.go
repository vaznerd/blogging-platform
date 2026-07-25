package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"regexp"
	"slices"
	"strings"

	"codeberg.org/vaznerd/blogging-platform/internal/middleware"
	"codeberg.org/vaznerd/blogging-platform/internal/user"
	"github.com/resend/resend-go/v3"
)

type Handler struct {
	service        *Service
	user           user.Service
	log            *slog.Logger
	mail           *resend.Client
	frontendURL    string
	trustedProxies []netip.Prefix
}

type errorResponse struct {
	Error string `json:"error"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func NewHandler(
	service *Service,
	usr user.Service,
	log *slog.Logger,
	mail *resend.Client,
	frontendURL string,
	trustedProxies []string,
) *Handler {
	prefixes := make([]netip.Prefix, 0, len(trustedProxies))
	for _, raw := range trustedProxies {
		prefix, err := netip.ParsePrefix(raw)
		if err != nil {
			addr, err2 := netip.ParseAddr(raw)
			if err2 != nil {
				log.Warn("invalid trusted proxy, skipping", "value", raw, "error", err)
				continue
			}
			bits := addr.BitLen()
			prefix = netip.PrefixFrom(addr, bits)
		}
		prefixes = append(prefixes, prefix)
	}
	return &Handler{
		service:        service,
		user:           usr,
		log:            log,
		mail:           mail,
		frontendURL:    frontendURL,
		trustedProxies: prefixes,
	}
}

func (h *Handler) isTrustedProxy(addr netip.Addr) bool {
	for _, prefix := range h.trustedProxies {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func (h *Handler) extractIP(r *http.Request) string {
	peer, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		peer = r.RemoteAddr
	}

	peerAddr, err := netip.ParseAddr(peer)
	if err != nil {
		return peer
	}

	if !h.isTrustedProxy(peerAddr) {
		return peerAddr.String()
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for _, raw := range slices.Backward(parts) {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			addr, parseErr := netip.ParseAddr(raw)
			if parseErr != nil {
				continue
			}
			if !h.isTrustedProxy(addr) {
				return addr.String()
			}
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if addr, parseErr := netip.ParseAddr(strings.TrimSpace(xri)); parseErr == nil {
			return addr.String()
		}
	}

	return peerAddr.String()
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func (h *Handler) sendVerificationEmail(ctx context.Context, userID, email string) {
	token, err := h.service.GenerateVerificationToken()
	if err != nil {
		h.log.ErrorContext(ctx, "failed to generate verification token", "error", err)
		return
	}

	if storeErr := h.service.StoreVerificationToken(ctx, userID, token); storeErr != nil {
		h.log.ErrorContext(ctx, "failed to store verification token", "error", storeErr)
		return
	}

	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", h.frontendURL, token)

	params := &resend.SendEmailRequest{
		From:    "Blogging Platform <onboarding@resend.dev>",
		To:      []string{email},
		Subject: "Verify your email",
		Html: fmt.Sprintf(
			`<p>Click <a href="%s">here</a> to verify your email.</p>`,
			verifyURL,
		),
	}

	if _, sendErr := h.mail.Emails.SendWithContext(ctx, params); sendErr != nil {
		h.log.ErrorContext(ctx, "failed to send verification email", "email", email, "error", sendErr)
	}
}

func (h *Handler) parseRegister(w http.ResponseWriter, r *http.Request) (string, string, string, bool) {
	w.Header().Set("Content-Type", "application/json")

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.WarnContext(r.Context(), "invalid json in Register", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInvalidBodyMsg})
		return "", "", "", false
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)

	if req.Email == "" || req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		h.log.WarnContext(r.Context(), "missing fields in Register")
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "email, username, and password are required"})
		return "", "", "", false
	}

	if !isValidEmail(req.Email) {
		w.WriteHeader(http.StatusBadRequest)
		h.log.WarnContext(r.Context(), "invalid email in Register")
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "invalid email"})
		return "", "", "", false
	}

	if len(req.Username) < minUsernameLength || len(req.Username) > maxUsernameLength {
		w.WriteHeader(http.StatusBadRequest)
		h.log.WarnContext(r.Context(), "invalid username length in Register")
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "username must be between 3 and 30 characters"})
		return "", "", "", false
	}

	if len(req.Password) < minPasswordLength {
		w.WriteHeader(http.StatusBadRequest)
		h.log.WarnContext(r.Context(), "password too short in Register")
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "password must be at least 8 characters"})
		return "", "", "", false
	}

	return req.Email, req.Username, req.Password, true
}

func (h *Handler) parseLogin(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	w.Header().Set("Content-Type", "application/json")

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.WarnContext(r.Context(), "invalid json in Login", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInvalidBodyMsg})
		return "", "", false
	}

	req.Email = strings.TrimSpace(req.Email)

	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		h.log.WarnContext(r.Context(), "missing fields in Login")
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "email and password are required"})
		return "", "", false
	}

	if !isValidEmail(req.Email) {
		w.WriteHeader(http.StatusBadRequest)
		h.log.WarnContext(r.Context(), "invalid email in Login")
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "invalid email"})
		return "", "", false
	}

	return req.Email, req.Password, true
}

func (h *Handler) parseRefreshToken(w http.ResponseWriter, r *http.Request) (string, bool) {
	w.Header().Set("Content-Type", "application/json")

	var req refreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.WarnContext(r.Context(), "invalid json in Refresh", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInvalidBodyMsg})
		return "", false
	}

	if req.RefreshToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		h.log.WarnContext(r.Context(), "missing refresh token in Refresh")
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "refresh_token is required"})
		return "", false
	}

	return req.RefreshToken, true
}

func (h *Handler) parseEmail(w http.ResponseWriter, r *http.Request) (string, bool) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.WarnContext(r.Context(), "invalid json", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInvalidBodyMsg})
		return "", false
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || !isValidEmail(req.Email) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "valid email is required"})
		return "", false
	}
	return req.Email, true
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	email, username, password, ok := h.parseRegister(w, r)
	if !ok {
		return
	}

	hashedPassword, err := h.service.HashPassword(password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to hash password", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	userID, err := h.user.Create(r.Context(), email, username, hashedPassword)
	if errors.Is(err, user.ErrConflict) {
		w.WriteHeader(http.StatusConflict)
		h.log.WarnContext(r.Context(), "email or username already exists", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "email or username already exists"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to create user", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	accessToken, err := h.service.GenerateAccessToken(userID, "user")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to generate access token", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	refreshToken, err := h.service.GenerateRefreshToken(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to generate refresh token", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	hash := HashRefreshToken(refreshToken)
	if sessErr := h.service.CreateSession(r.Context(), userID, hash, r.UserAgent(), h.extractIP(r)); sessErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to create session", "error", sessErr)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	go h.sendVerificationEmail(context.Background(), userID, email)

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(tokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	email, password, ok := h.parseLogin(w, r)
	if !ok {
		return
	}

	u, err := h.user.GetByEmail(r.Context(), email)
	if errors.Is(err, user.ErrNotFound) {
		_ = h.service.ComparePassword(dummyBcryptHash, password)
		w.WriteHeader(http.StatusUnauthorized)
		h.log.WarnContext(r.Context(), "invalid credentials in Login")
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "invalid credentials"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to get user by email", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	if !h.service.ComparePassword(u.PasswordHash, password) {
		w.WriteHeader(http.StatusUnauthorized)
		h.log.WarnContext(r.Context(), "invalid credentials in Login")
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "invalid credentials"})
		return
	}

	accessToken, err := h.service.GenerateAccessToken(u.ID, u.Role)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to generate access token", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	refreshToken, err := h.service.GenerateRefreshToken(u.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to generate refresh token", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	hash := HashRefreshToken(refreshToken)
	if sessErr := h.service.CreateSession(r.Context(), u.ID, hash, r.UserAgent(), h.extractIP(r)); sessErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to create session", "error", sessErr)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(tokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "unauthorized"})
		return
	}

	refreshToken, ok := h.parseRefreshToken(w, r)
	if !ok {
		return
	}

	hash := HashRefreshToken(refreshToken)
	session, err := h.service.GetSessionByRefreshTokenHash(r.Context(), hash)
	if errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrSessionRevoked) || errors.Is(err, ErrSessionExpired) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to get session", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	if session.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "forbidden"})
		return
	}

	if revokeErr := h.service.RevokeSession(r.Context(), session.ID); revokeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to revoke session", "error", revokeErr)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	refreshToken, ok := h.parseRefreshToken(w, r)
	if !ok {
		return
	}

	hash := HashRefreshToken(refreshToken)
	session, err := h.service.GetSessionByRefreshTokenHash(r.Context(), hash)
	if errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrSessionRevoked) || errors.Is(err, ErrSessionExpired) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "invalid or expired refresh token"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to get session", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	u, err := h.user.GetByID(r.Context(), session.UserID)
	if errors.Is(err, user.ErrNotFound) {
		_ = h.service.RevokeSession(r.Context(), session.ID)
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "invalid or expired refresh token"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to get user", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	newAccessToken, err := h.service.GenerateAccessToken(u.ID, u.Role)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to generate access token", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	newRefreshToken, err := h.service.GenerateRefreshToken(u.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to generate refresh token", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	newHash := HashRefreshToken(newRefreshToken)
	if sessErr := h.service.CreateSession(r.Context(), u.ID, newHash, r.UserAgent(), h.extractIP(r)); sessErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to create new session", "error", sessErr)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	if revokeErr := h.service.RevokeSession(r.Context(), session.ID); revokeErr != nil {
		h.log.ErrorContext(r.Context(), "failed to revoke old session during refresh", "error", revokeErr)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(tokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	})
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.WarnContext(r.Context(), "invalid json in VerifyEmail", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInvalidBodyMsg})
		return
	}

	if req.Token == "" {
		w.WriteHeader(http.StatusBadRequest)
		h.log.WarnContext(r.Context(), "missing token in VerifyEmail")
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "token is required"})
		return
	}

	if err := h.service.VerifyEmail(r.Context(), req.Token); err != nil {
		if errors.Is(err, ErrVerificationTokenInvalid) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(errorResponse{Error: "invalid or expired verification token"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to verify email", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(messageResponse{Message: "email verified successfully"})
}

func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	email, ok := h.parseEmail(w, r)
	if !ok {
		return
	}

	u, err := h.user.GetByEmail(r.Context(), email)
	if errors.Is(err, user.ErrNotFound) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(messageResponse{
			Message: "if the email exists, a verification link has been sent",
		})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to get user by email", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	if !u.IsEmailVerified {
		go h.sendVerificationEmail(context.Background(), u.ID, email)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(messageResponse{
		Message: "if the email exists, a verification link has been sent",
	})
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	email, ok := h.parseEmail(w, r)
	if !ok {
		return
	}

	u, err := h.user.GetByEmail(r.Context(), email)
	if errors.Is(err, user.ErrNotFound) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(messageResponse{Message: "if the email exists, a reset link has been sent"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to get user by email", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	token, err := h.service.GeneratePasswordResetToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to generate reset token", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	if storeErr := h.service.StorePasswordResetToken(r.Context(), u.ID, token); storeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to store reset token", "error", storeErr)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", h.frontendURL, token)
	params := &resend.SendEmailRequest{
		From:    "Blogging Platform <onboarding@resend.dev>",
		To:      []string{email},
		Subject: "Reset your password",
		Html: fmt.Sprintf(
			`<p>Click <a href="%s">here</a> to reset your password. This link expires in 1 hour.</p>`,
			resetURL,
		),
	}
	go func() {
		if _, sendErr := h.mail.Emails.SendWithContext(context.Background(), params); sendErr != nil {
			h.log.ErrorContext(context.Background(), "failed to send reset email", "email", email, "error", sendErr)
		}
	}()

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(messageResponse{Message: "if the email exists, a reset link has been sent"})
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "unauthorized"})
		return
	}

	u, err := h.user.GetByID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to get user", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	token, err := h.service.GeneratePasswordResetToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to generate reset token", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	if storeErr := h.service.StorePasswordResetToken(r.Context(), u.ID, token); storeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to store reset token", "error", storeErr)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", h.frontendURL, token)
	params := &resend.SendEmailRequest{
		From:    "Blogging Platform <onboarding@resend.dev>",
		To:      []string{u.Email},
		Subject: "Change your password",
		Html: fmt.Sprintf(
			`<p>Click <a href="%s">here</a> to change your password. This link expires in 1 hour.</p>`,
			resetURL,
		),
	}
	go func() {
		if _, sendErr := h.mail.Emails.SendWithContext(context.Background(), params); sendErr != nil {
			h.log.ErrorContext(
				context.Background(),
				"failed to send change password email",
				"email", u.Email,
				"error", sendErr,
			)
		}
	}()

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(messageResponse{Message: "a password reset link has been sent to your email"})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.WarnContext(r.Context(), "invalid json in ResetPassword", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInvalidBodyMsg})
		return
	}

	req.Token = strings.TrimSpace(req.Token)

	if req.Token == "" || req.NewPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "token and new_password are required"})
		return
	}

	if len(req.NewPassword) < minPasswordLength {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "new password must be at least 8 characters"})
		return
	}

	userID, err := h.service.VerifyPasswordResetToken(r.Context(), req.Token)
	if errors.Is(err, ErrPasswordResetTokenInvalid) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: "invalid or expired reset token"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to verify reset token", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	hashedPassword, err := h.service.HashPassword(req.NewPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to hash password", "error", err)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	if updateErr := h.user.UpdatePassword(r.Context(), userID, hashedPassword); updateErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.ErrorContext(r.Context(), "failed to update password", "error", updateErr)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: ErrInternalServerMsg})
		return
	}

	if revokeErr := h.service.RevokeAllUserSessions(r.Context(), userID); revokeErr != nil {
		h.log.ErrorContext(r.Context(), "failed to revoke sessions after password reset", "error", revokeErr)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(messageResponse{Message: "password reset successfully"})
}
