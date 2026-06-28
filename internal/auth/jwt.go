// This file creates and verifies JWTs (JSON Web Tokens).

package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the data we put inside the token. UserID identifies the caller;
// RegisteredClaims adds standard fields like expiry (exp) and issued-at (iat).
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// TokenManager knows the secret + lifetimes and can mint/parse tokens.
type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewTokenManager builds a TokenManager from config values.
func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

// Generate creates a signed token for userID that expires after ttl.
// TODO:
//
//	claims := Claims{UserID: userID, RegisteredClaims: jwt.RegisteredClaims{
//	    ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
//	    IssuedAt:  jwt.NewNumericDate(time.Now()),
//	}}
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//	return token.SignedString(m.secret)
func (m *TokenManager) Generate(userID string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Access/Refresh are convenience wrappers around Generate with the right TTLs.
func (m *TokenManager) Access(userID string) (string, error) { return m.Generate(userID, m.accessTTL) }
func (m *TokenManager) Refresh(userID string) (string, error) {
	return m.Generate(userID, m.refreshTTL)
}

// Parse verifies a token's signature + expiry and returns its claims.
// TODO:
//
//	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
//	    return m.secret, nil   // tell the library which key to verify with
//	})
//	if err != nil || !parsed.Valid { return nil, err }
//	return parsed.Claims.(*Claims), nil
func (m *TokenManager) Parse(tokenStr string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return m.secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, err
	}
	return parsed.Claims.(*Claims), nil
}
