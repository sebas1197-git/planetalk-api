// Database access for the chat module.
package chat

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// SaveMessage inserts a message and fills in its id + created_at.
func (r *Repository) SaveMessage(ctx context.Context, m *Message) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO messages (match_id, sender_id, body, translated_body, src_lang, dst_lang)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		m.MatchID, m.SenderID, m.Body, m.TranslatedBody, m.SrcLang, m.DstLang).
		Scan(&m.ID, &m.CreatedAt)
}

// ListMessages returns a page of a conversation, newest first.
func (r *Repository) ListMessages(ctx context.Context, matchID string, limit, offset int) ([]Message, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, match_id, sender_id, body, translated_body, src_lang, dst_lang, created_at
		 FROM messages WHERE match_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`, matchID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Message{}
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.MatchID, &m.SenderID, &m.Body,
			&m.TranslatedBody, &m.SrcLang, &m.DstLang, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// CountMessages returns how many messages a conversation has (for pagination).
func (r *Repository) CountMessages(ctx context.Context, matchID string) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM messages WHERE match_id = $1`, matchID).Scan(&n)
	return n, err
}

// GetUserLanguage returns a user's preferred language, or "" if not set.
func (r *Repository) GetUserLanguage(ctx context.Context, userID string) (string, error) {
	var lang *string
	err := r.db.QueryRow(ctx, `SELECT language FROM users WHERE id = $1`, userID).Scan(&lang)
	if err == pgx.ErrNoRows || lang == nil {
		return "", nil
	}
	return *lang, err
}
