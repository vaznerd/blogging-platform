package tag

import (
	"context"
)

type Service interface {
	Create(ctx context.Context, name string) (*Tag, error)
	GetByName(ctx context.Context, name string) (*Tag, error)
	List(ctx context.Context, limit, offset int) ([]*Tag, error)
	GetByPostID(ctx context.Context, postID string) ([]*Tag, error)
	AttachToPost(ctx context.Context, postID, tagID string) error
	DetachFromPost(ctx context.Context, postID, tagID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Create(ctx context.Context, name string) (*Tag, error) {
	return s.repo.Create(ctx, name)
}

func (s *service) GetByName(ctx context.Context, name string) (*Tag, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *service) List(ctx context.Context, limit, offset int) ([]*Tag, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *service) GetByPostID(ctx context.Context, postID string) ([]*Tag, error) {
	return s.repo.GetByPostID(ctx, postID)
}

func (s *service) AttachToPost(ctx context.Context, postID, tagID string) error {
	return s.repo.AttachToPost(ctx, postID, tagID)
}

func (s *service) DetachFromPost(ctx context.Context, postID, tagID string) error {
	return s.repo.DetachFromPost(ctx, postID, tagID)
}
