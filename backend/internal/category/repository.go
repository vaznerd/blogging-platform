package category

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Category struct {
	ID          string
	Name        string
	Slug        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Repository interface {
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

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, name, slug, description string) (*Category, error) {
	c := &Category{}
	err := r.db.QueryRow(
		ctx,
		`INSERT INTO categories (name, slug, description)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, slug, description, created_at, updated_at`,
		name, slug, description,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrConflict
		}
		return nil, err
	}
	return c, nil
}

func (r *repository) GetBySlug(ctx context.Context, slug string) (*Category, error) {
	c := &Category{}
	err := r.db.QueryRow(
		ctx,
		`SELECT id, name, slug, description, created_at, updated_at
		 FROM categories WHERE slug = $1`,
		slug,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*Category, error) {
	c := &Category{}
	err := r.db.QueryRow(
		ctx,
		`SELECT id, name, slug, description, created_at, updated_at
		 FROM categories WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]*Category, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT id, name, slug, description, created_at, updated_at
		 FROM categories
		 ORDER BY name ASC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		c := &Category{}
		if scanErr := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt); scanErr != nil {
			return nil, scanErr
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *repository) Update(ctx context.Context, id, name, slug, description string) (*Category, error) {
	c := &Category{}
	err := r.db.QueryRow(
		ctx,
		`UPDATE categories
		 SET name = $1, slug = $2, description = $3, updated_at = NOW()
		 WHERE id = $4
		 RETURNING id, name, slug, description, created_at, updated_at`,
		name, slug, description, id,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt)
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
		`DELETE FROM categories WHERE id = $1`, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) AttachToPost(ctx context.Context, postID, categoryID string) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO post_categories (post_id, category_id)
		 VALUES ($1, $2)
		 ON CONFLICT (post_id, category_id) DO NOTHING`,
		postID, categoryID,
	)
	return err
}

func (r *repository) DetachFromPost(ctx context.Context, postID, categoryID string) error {
	tag, err := r.db.Exec(
		ctx,
		`DELETE FROM post_categories WHERE post_id = $1 AND category_id = $2`,
		postID, categoryID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) ListByPostID(ctx context.Context, postID string) ([]*Category, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT c.id, c.name, c.slug, c.description, c.created_at, c.updated_at
		 FROM categories c
		 INNER JOIN post_categories pc ON pc.category_id = c.id
		 WHERE pc.post_id = $1
		 ORDER BY c.name ASC`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		c := &Category{}
		if scanErr := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt); scanErr != nil {
			return nil, scanErr
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}
