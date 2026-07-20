package tag

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

func (s *Service) Create(ctx context.Context, name string) (*Tag, error) {
	return s.repo.Create(ctx, name)
}

func (s *Service) GetByName(ctx context.Context, name string) (*Tag, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*Tag, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *Service) GetByPostID(ctx context.Context, postID string) ([]*Tag, error) {
	return s.repo.GetByPostID(ctx, postID)
}

func (s *Service) AttachToPost(ctx context.Context, postID, tagID string) error {
	return s.repo.AttachToPost(ctx, postID, tagID)
}

func (s *Service) DetachFromPost(ctx context.Context, postID, tagID string) error {
	return s.repo.DetachFromPost(ctx, postID, tagID)
}
