package user

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Create(ctx context.Context, email, username, passwordHash string) (string, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, userID string) (*User, error)
	GetByUserName(ctx context.Context, userName string) (*User, error)
	MarkEmailVerified(ctx context.Context, userID string) error
	UpdatePassword(ctx context.Context, userID string, passwordHash string) error
	UpdateMe(ctx context.Context, userID string, username, bio, avatarURL *string) (*User, error)
	DeleteAccount(ctx context.Context, userID, password string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Create(ctx context.Context, email, username, passwordHash string) (string, error) {
	username, err := normalizeUsername(username)
	if err != nil {
		return "", err
	}
	return s.repo.CreateUser(ctx, email, username, passwordHash)
}

func (s *service) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *service) GetByID(ctx context.Context, userID string) (*User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *service) GetByUserName(ctx context.Context, userName string) (*User, error) {
	return s.repo.GetByUserName(ctx, strings.ToLower(strings.TrimSpace(userName)))
}

func (s *service) MarkEmailVerified(ctx context.Context, userID string) error {
	return s.repo.MarkEmailVerified(ctx, userID)
}

func (s *service) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	return s.repo.UpdatePassword(ctx, userID, passwordHash)
}

func (s *service) UpdateMe(ctx context.Context, userID string, username, bio, avatarURL *string) (*User, error) {
	if username != nil {
		normalized, err := normalizeUsername(*username)
		if err != nil {
			return nil, err
		}
		username = &normalized
	}
	if bio != nil && len(*bio) > MaxBioLength {
		return nil, ErrInvalidBio
	}
	if avatarURL != nil && len(*avatarURL) > MaxAvatarURLLength {
		return nil, ErrInvalidAvatar
	}
	return s.repo.UpdateProfile(ctx, userID, username, bio, avatarURL)
}

func (s *service) DeleteAccount(ctx context.Context, userID, password string) error {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return ErrInvalidPassword
	}
	return s.repo.DeleteAccount(ctx, userID)
}

func normalizeUsername(username string) (string, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	if username == "me" {
		return "", ErrInvalidUsername
	}
	if len(username) < MinUsernameLength || len(username) > MaxUsernameLength || !usernameRegex.MatchString(username) {
		return "", ErrInvalidUsername
	}
	return username, nil
}
