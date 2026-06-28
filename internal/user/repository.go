// This file is the repository: the only place that runs SQL for the user module.
package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a user/row doesn't exist, so the handler can
// turn it into a 404 instead of a generic 500.
var ErrNotFound = errors.New("not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// UpdateProfileInput carries optional fields. nil = "don't change this field".
type UpdateProfileInput struct {
	DisplayName *string
	Bio         *string
	AvatarURL   *string
	Gender      *string
	Birthdate   *string // "YYYY-MM-DD"
	Country     *string
	Language    *string
}

// ---- Profiles -----------------------------------------------------------

// GetProfile loads one user plus their interests.
func (r *Repository) GetProfile(ctx context.Context, id string) (*Profile, error) {
	var p Profile
	err := r.db.QueryRow(ctx, `
		SELECT id, phone, display_name, bio, avatar_url, gender,
		       to_char(birthdate, 'YYYY-MM-DD'), country, language, created_at, last_seen
		FROM users WHERE id = $1`, id).
		Scan(&p.ID, &p.Phone, &p.DisplayName, &p.Bio, &p.AvatarURL, &p.Gender,
			&p.Birthdate, &p.Country, &p.Language, &p.CreatedAt, &p.LastSeen)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	interests, err := r.GetUserInterests(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Interests = interests
	return &p, nil
}

// UpdateProfile changes only the provided fields. COALESCE($n, column) keeps the
// existing value whenever the input is nil.
func (r *Repository) UpdateProfile(ctx context.Context, id string, in UpdateProfileInput) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET
			display_name = COALESCE($2, display_name),
			bio          = COALESCE($3, bio),
			avatar_url   = COALESCE($4, avatar_url),
			gender       = COALESCE($5, gender),
			birthdate    = COALESCE($6::date, birthdate),
			country      = COALESCE($7, country),
			language     = COALESCE($8, language)
		WHERE id = $1`,
		id, in.DisplayName, in.Bio, in.AvatarURL, in.Gender, in.Birthdate, in.Country, in.Language)
	return err
}

// ---- Interests ----------------------------------------------------------

// ListAllInterests returns the fixed interests catalogue.
func (r *Repository) ListAllInterests(ctx context.Context) ([]Interest, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name FROM interests ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanInterests(rows)
}

// GetUserInterests returns the interests a specific user picked.
func (r *Repository) GetUserInterests(ctx context.Context, userID string) ([]Interest, error) {
	rows, err := r.db.Query(ctx, `
		SELECT i.id, i.name FROM interests i
		JOIN user_interests ui ON ui.interest_id = i.id
		WHERE ui.user_id = $1
		ORDER BY i.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanInterests(rows)
}

// SetUserInterests replaces a user's interests with the given list, atomically.
func (r *Repository) SetUserInterests(ctx context.Context, userID string, interestIDs []int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // no-op if we Commit; undoes everything otherwise

	if _, err := tx.Exec(ctx, `DELETE FROM user_interests WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, id := range interestIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO user_interests (user_id, interest_id) VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`, userID, id); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ---- Friends ------------------------------------------------------------

// SendFriendRequest records a pending request from `me` to `target`.
func (r *Repository) SendFriendRequest(ctx context.Context, me, target string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO friends (user_id, friend_id, status) VALUES ($1, $2, 'pending')
		ON CONFLICT (user_id, friend_id) DO NOTHING`, me, target)
	return err
}

// AcceptFriendRequest accepts a pending request that `requester` sent to `me`.
func (r *Repository) AcceptFriendRequest(ctx context.Context, me, requester string) error {
	ct, err := r.db.Exec(ctx, `
		UPDATE friends SET status = 'accepted'
		WHERE user_id = $1 AND friend_id = $2 AND status = 'pending'`, requester, me)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound // no such pending request
	}
	return nil
}

// ListFriends returns accepted friends (in either direction).
func (r *Repository) ListFriends(ctx context.Context, me string) ([]UserSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.display_name, u.avatar_url, u.country
		FROM friends f
		JOIN users u ON u.id = CASE WHEN f.user_id = $1 THEN f.friend_id ELSE f.user_id END
		WHERE (f.user_id = $1 OR f.friend_id = $1) AND f.status = 'accepted'
		ORDER BY u.display_name`, me)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSummaries(rows)
}

// ListIncomingRequests returns people who sent `me` a pending request.
func (r *Repository) ListIncomingRequests(ctx context.Context, me string) ([]UserSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.display_name, u.avatar_url, u.country
		FROM friends f
		JOIN users u ON u.id = f.user_id
		WHERE f.friend_id = $1 AND f.status = 'pending'
		ORDER BY f.created_at DESC`, me)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSummaries(rows)
}

// RemoveFriend deletes the friendship/request in either direction.
func (r *Repository) RemoveFriend(ctx context.Context, me, other string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM friends
		WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)`, me, other)
	return err
}

// ---- small scan helpers -------------------------------------------------

func scanInterests(rows pgx.Rows) ([]Interest, error) {
	out := []Interest{}
	for rows.Next() {
		var it Interest
		if err := rows.Scan(&it.ID, &it.Name); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func scanSummaries(rows pgx.Rows) ([]UserSummary, error) {
	out := []UserSummary{}
	for rows.Next() {
		var u UserSummary
		if err := rows.Scan(&u.ID, &u.DisplayName, &u.AvatarURL, &u.Country); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
