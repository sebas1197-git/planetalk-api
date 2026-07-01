// Database access for the moderation module.
package moderation

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrUserNotFound is returned when the target user doesn't exist (FK violation).
var ErrUserNotFound = errors.New("user not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateReport records a report.
func (r *Repository) CreateReport(ctx context.Context, reporterID, reportedID, reason string, context_ *string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO reports (reporter_id, reported_id, reason, context)
		 VALUES ($1, $2, $3, $4)`, reporterID, reportedID, reason, context_)
	return mapFKError(err)
}

// AddBlock records that blocker blocked blocked (idempotent).
func (r *Repository) AddBlock(ctx context.Context, blockerID, blockedID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO blocks (blocker_id, blocked_id) VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`, blockerID, blockedID)
	return mapFKError(err)
}

// RemoveBlock removes a block.
func (r *Repository) RemoveBlock(ctx context.Context, blockerID, blockedID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM blocks WHERE blocker_id = $1 AND blocked_id = $2`, blockerID, blockedID)
	return err
}

// IsBlocked reports whether either user has blocked the other.
func (r *Repository) IsBlocked(ctx context.Context, a, b string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM blocks
			WHERE (blocker_id = $1 AND blocked_id = $2)
			   OR (blocker_id = $2 AND blocked_id = $1)
		)`, a, b).Scan(&exists)
	return exists, err
}

// ListBlocked returns the users blockerID has blocked.
func (r *Repository) ListBlocked(ctx context.Context, blockerID string) ([]BlockedUser, error) {
	rows, err := r.db.Query(ctx,
		`SELECT u.id, u.display_name, u.avatar_url
		 FROM blocks b JOIN users u ON u.id = b.blocked_id
		 WHERE b.blocker_id = $1
		 ORDER BY b.created_at DESC`, blockerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []BlockedUser{}
	for rows.Next() {
		var u BlockedUser
		if err := rows.Scan(&u.ID, &u.DisplayName, &u.AvatarURL); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// mapFKError turns a foreign-key violation (target user doesn't exist) into
// ErrUserNotFound so the handler can return a clean 404.
func mapFKError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return ErrUserNotFound
	}
	return err
}
