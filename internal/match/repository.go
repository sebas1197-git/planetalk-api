// Database access for the match module.
package match

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("match not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateMatch inserts a new active match and returns its id.
func (r *Repository) CreateMatch(ctx context.Context, userA, userB, channel string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO matches (user_a, user_b, channel_name) VALUES ($1, $2, $3) RETURNING id`,
		userA, userB, channel).Scan(&id)
	return id, err
}

// GetMatch loads a match by id.
func (r *Repository) GetMatch(ctx context.Context, id string) (*Match, error) {
	var m Match
	err := r.db.QueryRow(ctx,
		`SELECT id, user_a, user_b, channel_name, status, started_at, ended_at
		 FROM matches WHERE id = $1`, id).
		Scan(&m.ID, &m.UserA, &m.UserB, &m.ChannelName, &m.Status, &m.StartedAt, &m.EndedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// EndMatch marks an active match as ended.
func (r *Repository) EndMatch(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE matches SET status = 'ended', ended_at = now()
		 WHERE id = $1 AND status = 'active'`, id)
	return err
}

// PartnerSummary returns the lightweight profile of one user.
func (r *Repository) PartnerSummary(ctx context.Context, userID string) (*Partner, error) {
	var p Partner
	err := r.db.QueryRow(ctx,
		`SELECT id, display_name, avatar_url FROM users WHERE id = $1`, userID).
		Scan(&p.ID, &p.DisplayName, &p.AvatarURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}
