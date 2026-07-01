//go:build integration

package user

import (
	"context"
	"testing"

	"github.com/sebas1197-git/planetalk/internal/testutil"
)

func TestUserRepository_ProfileAndInterests(t *testing.T) {
	pool := testutil.PostgresDB(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	uid := testutil.CreateUser(t, pool, "+15550000010")

	name, lang := "Alice", "es"
	if err := repo.UpdateProfile(ctx, uid, UpdateProfileInput{DisplayName: &name, Language: &lang}); err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	p, err := repo.GetProfile(ctx, uid)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if p.DisplayName == nil || *p.DisplayName != "Alice" {
		t.Errorf("display_name not updated: %+v", p.DisplayName)
	}

	// Seeded interests exist; assign two.
	all, err := repo.ListAllInterests(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) < 3 {
		t.Fatalf("expected seeded interests, got %d", len(all))
	}
	if err := repo.SetUserInterests(ctx, uid, []int{all[0].ID, all[1].ID}); err != nil {
		t.Fatalf("SetUserInterests: %v", err)
	}
	got, _ := repo.GetUserInterests(ctx, uid)
	if len(got) != 2 {
		t.Errorf("expected 2 interests, got %d", len(got))
	}

	// GetProfile of an unknown id -> ErrNotFound.
	if _, err := repo.GetProfile(ctx, "00000000-0000-0000-0000-000000000000"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUserRepository_Friends(t *testing.T) {
	pool := testutil.PostgresDB(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	a := testutil.CreateUser(t, pool, "+15550000011")
	b := testutil.CreateUser(t, pool, "+15550000012")

	if err := repo.SendFriendRequest(ctx, a, b); err != nil {
		t.Fatalf("SendFriendRequest: %v", err)
	}
	reqs, _ := repo.ListIncomingRequests(ctx, b)
	if len(reqs) != 1 {
		t.Fatalf("expected 1 incoming request, got %d", len(reqs))
	}
	if err := repo.AcceptFriendRequest(ctx, b, a); err != nil {
		t.Fatalf("AcceptFriendRequest: %v", err)
	}
	if friends, _ := repo.ListFriends(ctx, a); len(friends) != 1 {
		t.Errorf("expected 1 friend for A, got %d", len(friends))
	}
	if friends, _ := repo.ListFriends(ctx, b); len(friends) != 1 {
		t.Errorf("expected 1 friend for B, got %d", len(friends))
	}
}
