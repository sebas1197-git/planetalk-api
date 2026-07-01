// This file is the "service": the brains of auth. It coordinates the repository
// (database), the token manager (JWTs), and the SMS sender. Handlers stay thin
// and just call these methods.
package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/sebas1197-git/planetalk/pkg/twilio"
)

const (
	otpTTL      = 5 * time.Minute // how long a code stays valid
	maxAttempts = 5               // wrong guesses allowed before a code is locked
)

// Sentinel errors so handlers can map them to stable API error codes.
var (
	ErrNoValidCode     = errors.New("no valid code; request a new one")
	ErrTooManyAttempts = errors.New("too many attempts; request a new code")
	ErrInvalidCode     = errors.New("invalid code")
	ErrInvalidRefresh  = errors.New("refresh token is invalid or expired")
)

// Service ties together everything auth needs.
type Service struct {
	repo   *Repository
	tokens *TokenManager
	sms    twilio.Sender
}

func NewService(repo *Repository, tokens *TokenManager, sms twilio.Sender) *Service {
	return &Service{repo: repo, tokens: tokens, sms: sms}
}

// RequestOTP generates a code, stores its HASH, sends it via SMS, and returns
// the plain code. The code is only ever exposed by the handler in DEV mode
// (for automated testing) — never in production.
func (s *Service) RequestOTP(ctx context.Context, phone string) (string, error) {
	code, err := generateCode()
	if err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	if err := s.repo.SaveOTP(ctx, phone, string(hash), time.Now().Add(otpTTL)); err != nil {
		return "", err
	}
	if err := s.sms.SendSMS(ctx, phone, fmt.Sprintf("Your Planetalk code is %s", code)); err != nil {
		return "", err
	}
	return code, nil
}

// VerifyOTP checks the code; on success creates/finds the user and returns tokens.
func (s *Service) VerifyOTP(ctx context.Context, phone, code string) (access, refresh string, err error) {
	otpID, codeHash, attempts, err := s.repo.LatestOTP(ctx, phone)
	if err != nil {
		return "", "", err
	}
	if otpID == "" {
		return "", "", ErrNoValidCode
	}
	if attempts >= maxAttempts {
		return "", "", ErrTooManyAttempts
	}

	// bcrypt.CompareHashAndPassword returns nil only when the code matches.
	if err := bcrypt.CompareHashAndPassword([]byte(codeHash), []byte(code)); err != nil {
		_ = s.repo.IncrementAttempts(ctx, otpID) // count the wrong guess
		return "", "", ErrInvalidCode
	}

	// Code is correct: clear used codes and create-or-find the user.
	_ = s.repo.DeleteOTPsForPhone(ctx, phone)
	userID, err := s.repo.UpsertUser(ctx, phone)
	if err != nil {
		return "", "", err
	}

	if access, err = s.tokens.Access(userID); err != nil {
		return "", "", err
	}
	if refresh, err = s.issueRefreshToken(ctx, userID); err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

// RefreshTokens validates a refresh token, ROTATES it (revokes the old, issues a
// new one), and returns a fresh access token + the new refresh token.
//
// If a revoked token is replayed, that's a theft signal: we revoke every token
// for that user and reject.
func (s *Service) RefreshTokens(ctx context.Context, rawRefresh string) (access, newRefresh string, err error) {
	hash := hashRefreshToken(rawRefresh)
	rt, err := s.repo.GetRefreshToken(ctx, hash)
	if err != nil {
		return "", "", err
	}
	if rt == nil {
		return "", "", ErrInvalidRefresh
	}
	if rt.Revoked {
		// Reuse of an already-rotated token -> assume compromise, revoke all.
		_ = s.repo.RevokeAllForUser(ctx, rt.UserID)
		return "", "", ErrInvalidRefresh
	}
	if time.Now().After(rt.ExpiresAt) {
		return "", "", ErrInvalidRefresh
	}

	// Rotate: revoke the used token, mint a fresh one.
	if err := s.repo.RevokeRefreshToken(ctx, hash); err != nil {
		return "", "", err
	}
	if newRefresh, err = s.issueRefreshToken(ctx, rt.UserID); err != nil {
		return "", "", err
	}
	if access, err = s.tokens.Access(rt.UserID); err != nil {
		return "", "", err
	}
	return access, newRefresh, nil
}

// Logout revokes the given refresh token (best-effort; unknown tokens are a no-op).
func (s *Service) Logout(ctx context.Context, rawRefresh string) error {
	return s.repo.RevokeRefreshToken(ctx, hashRefreshToken(rawRefresh))
}

// issueRefreshToken creates a random token, stores its hash, and returns the raw.
func (s *Service) issueRefreshToken(ctx context.Context, userID string) (string, error) {
	raw, err := generateRefreshToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(s.tokens.RefreshTTL())
	if err := s.repo.SaveRefreshToken(ctx, userID, hashRefreshToken(raw), expiresAt); err != nil {
		return "", err
	}
	return raw, nil
}

// generateCode returns a cryptographically-random 6-digit string like "048213".
func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
