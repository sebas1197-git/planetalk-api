package agora

import (
	"errors"
	"testing"
	"time"
)

func TestDisabledWhenNoCreds(t *testing.T) {
	b := New("", "", time.Hour)
	if b.Enabled() {
		t.Error("should be disabled with empty credentials")
	}
	if _, err := b.BuildRTCToken("chan", 1); !errors.Is(err, ErrNotConfigured) {
		t.Errorf("expected ErrNotConfigured, got %v", err)
	}
}

func TestBuildTokenWithCreds(t *testing.T) {
	// Dummy but well-formed credentials (32 hex chars each).
	b := New("20b7c51ff4c644ab80cf5a4e646b0537", "5cba2d0a3a9a4f0e9b6f2c1d8e7a6b5c", time.Hour)
	if !b.Enabled() {
		t.Fatal("should be enabled with credentials")
	}
	tok, err := b.BuildRTCToken("chan", 12345)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if tok == "" {
		t.Error("expected a non-empty token")
	}
}
