package user

import (
	"context"
)

type Service interface {
	Create(ctx context.Context, email, username, passwordHash string) (string, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, userID string) (*User, error)
	MarkEmailVerified(ctx context.Context, userID string) error
	UpdatePassword(ctx context.Context, userID string, passwordHash string) error
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
	return s.repo.CreateUser(ctx, email, username, passwordHash)
}

func (s *service) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *service) GetByID(ctx context.Context, userID string) (*User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *service) MarkEmailVerified(ctx context.Context, userID string) error {
	return s.repo.MarkEmailVerified(ctx, userID)
}

func (s *service) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	return s.repo.UpdatePassword(ctx, userID, passwordHash)
}
