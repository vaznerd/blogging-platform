package category

import (
	"context"
)

type Service interface {
	Create(ctx context.Context, name, slug, description string) (*Category, error)
	GetBySlug(ctx context.Context, slug string) (*Category, error)
	GetByID(ctx context.Context, id string) (*Category, error)
	List(ctx context.Context, limit, offset int) ([]*Category, error)
	Update(ctx context.Context, id, name, slug, description string) (*Category, error)
	Delete(ctx context.Context, id string) error
	AttachToPost(ctx context.Context, postID, categoryID string) error
	DetachFromPost(ctx context.Context, postID, categoryID string) error
	ListByPostID(ctx context.Context, postID string) ([]*Category, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Create(ctx context.Context, name, slug, description string) (*Category, error) {
	return s.repo.Create(ctx, name, slug, description)
}

func (s *service) GetBySlug(ctx context.Context, slug string) (*Category, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *service) GetByID(ctx context.Context, id string) (*Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, limit, offset int) ([]*Category, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *service) Update(ctx context.Context, id, name, slug, description string) (*Category, error) {
	return s.repo.Update(ctx, id, name, slug, description)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) AttachToPost(ctx context.Context, postID, categoryID string) error {
	return s.repo.AttachToPost(ctx, postID, categoryID)
}

func (s *service) DetachFromPost(ctx context.Context, postID, categoryID string) error {
	return s.repo.DetachFromPost(ctx, postID, categoryID)
}

func (s *service) ListByPostID(ctx context.Context, postID string) ([]*Category, error) {
	return s.repo.ListByPostID(ctx, postID)
}
