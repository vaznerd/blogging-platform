package tag

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Tag struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

type Repository interface {
	Create(ctx context.Context, name string) (*Tag, error)
	GetByName(ctx context.Context, name string) (*Tag, error)
	List(ctx context.Context, limit, offset int) ([]*Tag, error)
	GetByPostID(ctx context.Context, postID string) ([]*Tag, error)
	AttachToPost(ctx context.Context, postID, tagID string) error
	DetachFromPost(ctx context.Context, postID, tagID string) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, name string) (*Tag, error) {
	t := &Tag{}
	err := r.db.QueryRow(
		ctx,
		`INSERT INTO tags (name)
		 VALUES ($1)
		 RETURNING id, name, created_at`,
		name,
	).Scan(&t.ID, &t.Name, &t.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrConflict
		}
		return nil, err
	}
	return t, nil
}

func (r *repository) GetByName(ctx context.Context, name string) (*Tag, error) {
	t := &Tag{}
	err := r.db.QueryRow(
		ctx,
		`SELECT id, name, created_at
		 FROM tags WHERE name = $1`,
		name,
	).Scan(&t.ID, &t.Name, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]*Tag, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT id, name, created_at
		 FROM tags
		 ORDER BY name ASC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*Tag
	for rows.Next() {
		t := &Tag{}
		if scanErr := rows.Scan(&t.ID, &t.Name, &t.CreatedAt); scanErr != nil {
			return nil, scanErr
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *repository) GetByPostID(ctx context.Context, postID string) ([]*Tag, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT t.id, t.name, t.created_at
		 FROM tags t
		 INNER JOIN post_tags pt ON pt.tag_id = t.id
		 WHERE pt.post_id = $1
		 ORDER BY t.name ASC`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*Tag
	for rows.Next() {
		t := &Tag{}
		if scanErr := rows.Scan(&t.ID, &t.Name, &t.CreatedAt); scanErr != nil {
			return nil, scanErr
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *repository) AttachToPost(ctx context.Context, postID, tagID string) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO post_tags (post_id, tag_id)
		 VALUES ($1, $2)
		 ON CONFLICT (post_id, tag_id) DO NOTHING`,
		postID, tagID,
	)
	return err
}

func (r *repository) DetachFromPost(ctx context.Context, postID, tagID string) error {
	tag, err := r.db.Exec(
		ctx,
		`DELETE FROM post_tags WHERE post_id = $1 AND tag_id = $2`,
		postID, tagID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
