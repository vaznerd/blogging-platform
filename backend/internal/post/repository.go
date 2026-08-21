package post

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Post struct {
	ID              string
	AuthorID        string
	Title           string
	Slug            string
	MarkdownContent string
	HTMLContent     string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	PublishedAt     *time.Time
}

type Repository interface {
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

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(
	ctx context.Context, authorID, title, slug, markdownContent, htmlContent, status string,
) (*Post, error) {
	p := &Post{}
	err := r.db.QueryRow(
		ctx,
		`INSERT INTO posts (author_id, title, slug, markdown_content, html_content, status)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, author_id, title, slug, markdown_content, html_content, status, created_at, updated_at, published_at`,
		authorID, title, slug, markdownContent, htmlContent, status,
	).Scan(
		&p.ID, &p.AuthorID, &p.Title, &p.Slug,
		&p.MarkdownContent, &p.HTMLContent, &p.Status,
		&p.CreatedAt, &p.UpdatedAt, &p.PublishedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSlugAlreadyExists
		}
		if errors.As(err, &pgErr) && pgErr.Code == "22P02" {
			return nil, ErrInvalidStatus
		}
		return nil, err
	}
	return p, nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*Post, error) {
	p := &Post{}
	err := r.db.QueryRow(
		ctx,
		`SELECT id, author_id, title, slug, markdown_content, html_content, status, created_at, updated_at, published_at
		 FROM posts WHERE id = $1`,
		id,
	).Scan(
		&p.ID, &p.AuthorID, &p.Title, &p.Slug,
		&p.MarkdownContent, &p.HTMLContent, &p.Status,
		&p.CreatedAt, &p.UpdatedAt, &p.PublishedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *repository) GetBySlug(ctx context.Context, authorID, slug string) (*Post, error) {
	p := &Post{}
	err := r.db.QueryRow(
		ctx,
		`SELECT id, author_id, title, slug, markdown_content, html_content, status, created_at, updated_at, published_at
		 FROM posts WHERE author_id = $1 AND slug = $2`,
		authorID, slug,
	).Scan(
		&p.ID, &p.AuthorID, &p.Title, &p.Slug,
		&p.MarkdownContent, &p.HTMLContent, &p.Status,
		&p.CreatedAt, &p.UpdatedAt, &p.PublishedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]*Post, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT id, author_id, title, slug, markdown_content, html_content, status, created_at, updated_at, published_at
		 FROM posts
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		p := &Post{}
		if scanErr := rows.Scan(
			&p.ID, &p.AuthorID, &p.Title, &p.Slug,
			&p.MarkdownContent, &p.HTMLContent, &p.Status,
			&p.CreatedAt, &p.UpdatedAt, &p.PublishedAt,
		); scanErr != nil {
			return nil, scanErr
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *repository) ListByUsername(ctx context.Context, username string, limit, offset int) ([]*Post, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT p.id, p.author_id, p.title, p.slug,
		 p.markdown_content, p.html_content, p.status,
		 p.created_at, p.updated_at, p.published_at
		 FROM posts p
		 INNER JOIN users u ON u.id = p.author_id
		 WHERE u.username = $1
		 ORDER BY p.created_at DESC
		 LIMIT $2 OFFSET $3`,
		username, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		p := &Post{}
		if scanErr := rows.Scan(
			&p.ID, &p.AuthorID, &p.Title, &p.Slug,
			&p.MarkdownContent, &p.HTMLContent, &p.Status,
			&p.CreatedAt, &p.UpdatedAt, &p.PublishedAt,
		); scanErr != nil {
			return nil, scanErr
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *repository) Update(
	ctx context.Context, id, title, slug, markdownContent, htmlContent, status string,
) (*Post, error) {
	p := &Post{}
	err := r.db.QueryRow(
		ctx,
		`UPDATE posts
		 SET title = $1, slug = $2, markdown_content = $3, html_content = $4, status = $5, updated_at = NOW()
		 WHERE id = $6
		 RETURNING id, author_id, title, slug, markdown_content, html_content, status, created_at, updated_at, published_at`,
		title, slug, markdownContent, htmlContent, status, id,
	).Scan(
		&p.ID, &p.AuthorID, &p.Title, &p.Slug,
		&p.MarkdownContent, &p.HTMLContent, &p.Status,
		&p.CreatedAt, &p.UpdatedAt, &p.PublishedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSlugAlreadyExists
		}
		if errors.As(err, &pgErr) && pgErr.Code == "22P02" {
			return nil, ErrInvalidStatus
		}
		return nil, err
	}
	return p, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(
		ctx,
		`DELETE FROM posts WHERE id = $1`, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
