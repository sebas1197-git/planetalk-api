//go:build integration

package auth

import (
	"context"
	"testing"
	"time"

	"github.com/sebas1197-git/planetalk/internal/testutil"
)

func TestRepository_OTPAndUser(t *testing.T) {
	pool := testutil.PostgresDB(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	phone := "+15550000001"

	if err := repo.SaveOTP(ctx, phone, "hash123", time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("SaveOTP: %v", err)
	}
	id, hash, attempts, err := repo.LatestOTP(ctx, phone)
	if err != nil {
		t.Fatalf("LatestOTP: %v", err)
	}
	if id == "" || hash != "hash123" || attempts != 0 {
		t.Fatalf("unexpected otp: id=%q hash=%q attempts=%d", id, hash, attempts)
	}

	// Expired codes are ignored.
	if err := repo.SaveOTP(ctx, "+15550000009", "old", time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if id, _, _, _ := repo.LatestOTP(ctx, "+15550000009"); id != "" {
		t.Error("expired OTP should not be returned")
	}

	// UpsertUser is idempotent (same phone -> same id).
	uid, err := repo.UpsertUser(ctx, phone)
	if err != nil {
		t.Fatalf("UpsertUser: %v", err)
	}
	if uid2, _ := repo.UpsertUser(ctx, phone); uid != uid2 {
		t.Errorf("upsert not idempotent: %s vs %s", uid, uid2)
	}
}

func TestRepository_RefreshTokens(t *testing.T) {
	pool := testutil.PostgresDB(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	uid := testutil.CreateUser(t, pool, "+15550000002")

	const hash = "tokenhash1"
	if err := repo.SaveRefreshToken(ctx, uid, hash, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("SaveRefreshToken: %v", err)
	}

	rt, err := repo.GetRefreshToken(ctx, hash)
	if err != nil || rt == nil {
		t.Fatalf("GetRefreshToken: rt=%v err=%v", rt, err)
	}
	if rt.UserID != uid || rt.Revoked {
		t.Fatalf("unexpected token: %+v", rt)
	}

	if err := repo.RevokeRefreshToken(ctx, hash); err != nil {
		t.Fatal(err)
	}
	if rt, _ := repo.GetRefreshToken(ctx, hash); rt == nil || !rt.Revoked {
		t.Error("token should be revoked")
	}

	// Unknown token -> (nil, nil).
	if rt, err := repo.GetRefreshToken(ctx, "does-not-exist"); err != nil || rt != nil {
		t.Errorf("expected (nil,nil), got (%v,%v)", rt, err)
	}

	// RevokeAllForUser revokes everything.
	_ = repo.SaveRefreshToken(ctx, uid, "h2", time.Now().Add(time.Hour))
	if err := repo.RevokeAllForUser(ctx, uid); err != nil {
		t.Fatal(err)
	}
	if rt, _ := repo.GetRefreshToken(ctx, "h2"); rt == nil || !rt.Revoked {
		t.Error("RevokeAllForUser should revoke h2")
	}
}
