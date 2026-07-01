// Database access for the feed module.
package feed

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("post not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreatePost inserts a post and fills in its id + created_at.
func (r *Repository) CreatePost(ctx context.Context, p *Post) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO posts (user_id, media_url, media_type, caption, visibility)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		p.UserID, p.MediaURL, p.MediaType, p.Caption, p.Visibility).
		Scan(&p.ID, &p.CreatedAt)
}

// GetPost loads a single post.
func (r *Repository) GetPost(ctx context.Context, id string) (*Post, error) {
	var p Post
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, media_url, media_type, caption, visibility, created_at
		 FROM posts WHERE id = $1`, id).
		Scan(&p.ID, &p.UserID, &p.MediaURL, &p.MediaType, &p.Caption, &p.Visibility, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// DeletePost removes a post owned by ownerID. Returns ErrNotFound if there's no
// such post for that owner (also covers "not yours").
func (r *Repository) DeletePost(ctx context.Context, id, ownerID string) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM posts WHERE id = $1 AND user_id = $2`, id, ownerID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListByUser returns a page of a user's posts limited to the given visibilities.
func (r *Repository) ListByUser(ctx context.Context, userID string, visibilities []string, limit, offset int) ([]Post, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, media_url, media_type, caption, visibility, created_at
		 FROM posts
		 WHERE user_id = $1 AND visibility = ANY($2)
		 ORDER BY created_at DESC
		 LIMIT $3 OFFSET $4`, userID, visibilities, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Post{}
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.MediaURL, &p.MediaType, &p.Caption, &p.Visibility, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// CountByUser counts a user's posts within the given visibilities (for pagination).
func (r *Repository) CountByUser(ctx context.Context, userID string, visibilities []string) (int, error) {
	var n int
	err := r.db.QueryRow(ctx,
		`SELECT count(*) FROM posts WHERE user_id = $1 AND visibility = ANY($2)`,
		userID, visibilities).Scan(&n)
	return n, err
}

// AreFriends reports whether a and b are accepted friends (either direction).
func (r *Repository) AreFriends(ctx context.Context, a, b string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM friends
			WHERE status = 'accepted'
			  AND ((user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1))
		)`, a, b).Scan(&exists)
	return exists, err
}
