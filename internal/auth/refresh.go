// Refresh tokens are OPAQUE (random) strings, not JWTs. We store only their
// SHA-256 hash in the database so a DB leak can't be used to forge sessions.
//
// Rotation: each time a refresh token is used, it is revoked and a brand-new one
// is issued. If a already-revoked token is presented again, that signals theft
// (someone kept an old copy), so we revoke ALL of that user's tokens.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// RefreshToken is a stored refresh-token record (without the raw value).
type RefreshToken struct {
	UserID    string
	ExpiresAt time.Time
	Revoked   bool
}

// generateRefreshToken returns a new random opaque token (64 hex chars).
func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashRefreshToken returns the SHA-256 hex of a raw token. Deterministic, so we
// can look it up; fast, because the token is already high-entropy random.
func hashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
