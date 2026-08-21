package comment

import (
	"context"
)

type Service interface {
	Create(ctx context.Context, postID, authorID, content string) (*Comment, error)
	GetByID(ctx context.Context, id string) (*Comment, error)
	ListByPostID(ctx context.Context, postID string, limit, offset int) ([]*Comment, error)
	Update(ctx context.Context, id, content string) (*Comment, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Create(ctx context.Context, postID, authorID, content string) (*Comment, error) {
	return s.repo.Create(ctx, postID, authorID, content)
}

func (s *service) GetByID(ctx context.Context, id string) (*Comment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListByPostID(ctx context.Context, postID string, limit, offset int) ([]*Comment, error) {
	return s.repo.ListByPostID(ctx, postID, limit, offset)
}

func (s *service) Update(ctx context.Context, id, content string) (*Comment, error) {
	return s.repo.Update(ctx, id, content)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
