package category

import (
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, name, slug, description string) (*Category, error) {
	return s.repo.Create(ctx, name, slug, description)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*Category, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*Category, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *Service) Update(ctx context.Context, id, name, slug, description string) (*Category, error) {
	return s.repo.Update(ctx, id, name, slug, description)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) AttachToPost(ctx context.Context, postID, categoryID string) error {
	return s.repo.AttachToPost(ctx, postID, categoryID)
}

func (s *Service) DetachFromPost(ctx context.Context, postID, categoryID string) error {
	return s.repo.DetachFromPost(ctx, postID, categoryID)
}

func (s *Service) ListByPostID(ctx context.Context, postID string) ([]*Category, error) {
	return s.repo.ListByPostID(ctx, postID)
}
