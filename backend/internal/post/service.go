package post

import (
	"context"
)

type Service interface {
	Create(
		ctx context.Context, authorID, title, slug, markdownContent, htmlContent, status string,
	) (*Post, error)
	GetByID(ctx context.Context, id string) (*Post, error)
	GetBySlug(ctx context.Context, authorID, slug string) (*Post, error)
	List(ctx context.Context, limit, offset int) ([]*Post, error)
	ListByUsername(ctx context.Context, username string, limit, offset int) ([]*Post, error)
	Update(
		ctx context.Context, id, title, slug, markdownContent, htmlContent, status string,
	) (*Post, error)
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

func (s *service) Create(
	ctx context.Context, authorID, title, slug, markdownContent, htmlContent, status string,
) (*Post, error) {
	return s.repo.Create(ctx, authorID, title, slug, markdownContent, htmlContent, status)
}

func (s *service) GetByID(ctx context.Context, id string) (*Post, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetBySlug(ctx context.Context, authorID, slug string) (*Post, error) {
	return s.repo.GetBySlug(ctx, authorID, slug)
}

func (s *service) List(ctx context.Context, limit, offset int) ([]*Post, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *service) ListByUsername(ctx context.Context, username string, limit, offset int) ([]*Post, error) {
	return s.repo.ListByUsername(ctx, username, limit, offset)
}

func (s *service) Update(
	ctx context.Context, id, title, slug, markdownContent, htmlContent, status string,
) (*Post, error) {
	return s.repo.Update(ctx, id, title, slug, markdownContent, htmlContent, status)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
