// This file is the "repository": the only place that talks to the database for
// auth. Keeping SQL here (not in handlers) means the rest of the code works with
// Go types, and you can change the DB without touching business logic.
package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository holds the DB pool and runs the auth-related queries.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// SaveOTP stores a new OTP row (phone + hashed code + expiry).
// TODO: INSERT INTO otp_codes (phone, code_hash, expires_at) VALUES ($1,$2,$3)
func (r *Repository) SaveOTP(ctx context.Context, phone, codeHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO otp_codes (phone, code_hash, expires_at) VALUES ($1, $2, $3)`, phone, codeHash, expiresAt)
	return err
}

// LatestOTP returns the most recent, non-expired OTP for a phone.
// TODO: SELECT id, code_hash, attempts FROM otp_codes
//
//	WHERE phone=$1 AND expires_at > now() ORDER BY created_at DESC LIMIT 1
func (r *Repository) LatestOTP(ctx context.Context, phone string) (otpID, codeHash string, attempts int, err error) {
	err = r.db.QueryRow(ctx,
		`SELECT id, code_hash, attempts FROM otp_codes
			 WHERE phone=$1 AND expires_at > now()
			 ORDER BY created_at DESC LIMIT 1`, phone).Scan(&otpID, &codeHash, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", 0, nil // no valid code = not an error, just empty
	}
	return otpID, codeHash, attempts, err

}

// IncrementAttempts bumps the attempts counter to limit brute-force guessing.
// TODO: UPDATE otp_codes SET attempts = attempts + 1 WHERE id=$1
func (r *Repository) IncrementAttempts(ctx context.Context, otpID string) error {
	_, err := r.db.Exec(ctx, `UPDATE otp_codes SET attempts = attempts + 1 WHERE id = $1`, otpID)
	return err
}

// DeleteOTPsForPhone removes used/old codes once login succeeds.
// TODO: DELETE FROM otp_codes WHERE phone=$1
func (r *Repository) DeleteOTPsForPhone(ctx context.Context, phone string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM otp_codes WHERE phone = $1`, phone)
	return err
}

// UpsertUser finds the user by phone, or creates one if new. Returns the user id.
// "Upsert" = update-or-insert. For phone login, first OTP verify creates them.
// TODO:
//
//	INSERT INTO users (phone) VALUES ($1)
//	ON CONFLICT (phone) DO UPDATE SET last_seen = now()
//	RETURNING id
func (r *Repository) UpsertUser(ctx context.Context, phone string) (userID string, err error) {
	err = r.db.QueryRow(ctx,
		`INSERT INTO users (phone) VALUES ($1)
               ON CONFLICT (phone) DO UPDATE SET last_seen = now()
               RETURNING id`, phone).Scan(&userID)
	return userID, err
}
