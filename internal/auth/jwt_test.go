package auth

import (
	"testing"
	"time"
)

func TestTokenGenerateAndParse(t *testing.T) {
	m := NewTokenManager("test-secret", time.Minute, time.Hour)
	tok, err := m.Access("user-123")
	if err != nil {
		t.Fatalf("access: %v", err)
	}
	claims, err := m.Parse(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want user-123", claims.UserID)
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	signer := NewTokenManager("secret-a", time.Minute, time.Hour)
	verifier := NewTokenManager("secret-b", time.Minute, time.Hour)
	tok, _ := signer.Access("u")
	if _, err := verifier.Parse(tok); err == nil {
		t.Error("expected error parsing a token signed with a different secret")
	}
}

func TestParseRejectsExpired(t *testing.T) {
	m := NewTokenManager("s", -time.Minute, time.Hour) // negative TTL => already expired
	tok, _ := m.Access("u")
	if _, err := m.Parse(tok); err == nil {
		t.Error("expected error parsing an expired token")
	}
}

func TestRefreshTTL(t *testing.T) {
	m := NewTokenManager("s", time.Minute, 42*time.Hour)
	if got := m.RefreshTTL(); got != 42*time.Hour {
		t.Errorf("RefreshTTL = %v, want 42h", got)
	}
}
