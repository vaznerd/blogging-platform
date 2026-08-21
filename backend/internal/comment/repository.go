package comment

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Comment struct {
	ID        string
	PostID    string
	AuthorID  string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository interface {
	Create(ctx context.Context, postID, authorID, content string) (*Comment, error)
	GetByID(ctx context.Context, id string) (*Comment, error)
	ListByPostID(ctx context.Context, postID string, limit, offset int) ([]*Comment, error)
	Update(ctx context.Context, id, content string) (*Comment, error)
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, postID, authorID, content string) (*Comment, error) {
	c := &Comment{}
	err := r.db.QueryRow(
		ctx,
		`INSERT INTO comments (post_id, author_id, content)
		 VALUES ($1, $2, $3)
		 RETURNING id, post_id, author_id, content, created_at, updated_at`,
		postID, authorID, content,
	).Scan(&c.ID, &c.PostID, &c.AuthorID, &c.Content, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*Comment, error) {
	c := &Comment{}
	err := r.db.QueryRow(
		ctx,
		`SELECT id, post_id, author_id, content, created_at, updated_at
		 FROM comments WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.PostID, &c.AuthorID, &c.Content, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *repository) ListByPostID(ctx context.Context, postID string, limit, offset int) ([]*Comment, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT id, post_id, author_id, content, created_at, updated_at
		 FROM comments
		 WHERE post_id = $1
		 ORDER BY created_at ASC
		 LIMIT $2 OFFSET $3`,
		postID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		c := &Comment{}
		if scanErr := rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.Content, &c.CreatedAt, &c.UpdatedAt); scanErr != nil {
			return nil, scanErr
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (r *repository) Update(ctx context.Context, id, content string) (*Comment, error) {
	c := &Comment{}
	err := r.db.QueryRow(
		ctx,
		`UPDATE comments
		 SET content = $1, updated_at = NOW()
		 WHERE id = $2
		 RETURNING id, post_id, author_id, content, created_at, updated_at`,
		content, id,
	).Scan(&c.ID, &c.PostID, &c.AuthorID, &c.Content, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(
		ctx,
		`DELETE FROM comments WHERE id = $1`, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
