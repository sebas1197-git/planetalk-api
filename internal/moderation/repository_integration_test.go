//go:build integration

package moderation

import (
	"context"
	"errors"
	"testing"

	"github.com/sebas1197-git/planetalk/internal/testutil"
)

func TestModerationRepository_Blocks(t *testing.T) {
	pool := testutil.PostgresDB(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	a := testutil.CreateUser(t, pool, "+15550000030")
	b := testutil.CreateUser(t, pool, "+15550000031")

	if blocked, _ := repo.IsBlocked(ctx, a, b); blocked {
		t.Error("should not be blocked initially")
	}
	if err := repo.AddBlock(ctx, a, b); err != nil {
		t.Fatalf("AddBlock: %v", err)
	}
	// Symmetric: enforced in both directions.
	if blocked, _ := repo.IsBlocked(ctx, a, b); !blocked {
		t.Error("a->b should be blocked")
	}
	if blocked, _ := repo.IsBlocked(ctx, b, a); !blocked {
		t.Error("b->a should also report blocked")
	}
	if list, _ := repo.ListBlocked(ctx, a); len(list) != 1 {
		t.Errorf("expected 1 blocked, got %d", len(list))
	}
	if err := repo.RemoveBlock(ctx, a, b); err != nil {
		t.Fatal(err)
	}
	if blocked, _ := repo.IsBlocked(ctx, a, b); blocked {
		t.Error("should be unblocked")
	}
}

func TestModerationRepository_Reports(t *testing.T) {
	pool := testutil.PostgresDB(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	a := testutil.CreateUser(t, pool, "+15550000032")
	b := testutil.CreateUser(t, pool, "+15550000033")

	// Reporting a non-existent user -> ErrUserNotFound (FK violation mapped).
	err := repo.CreateReport(ctx, a, "00000000-0000-0000-0000-000000000000", "spam", nil)
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
	// Valid report succeeds.
	if err := repo.CreateReport(ctx, a, b, "spam", nil); err != nil {
		t.Fatalf("CreateReport: %v", err)
	}
}
