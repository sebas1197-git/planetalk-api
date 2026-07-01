//go:build integration

package match

import (
	"context"
	"errors"
	"testing"

	"github.com/sebas1197-git/planetalk/internal/testutil"
)

func TestMatchRepository(t *testing.T) {
	pool := testutil.PostgresDB(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	a := testutil.CreateUser(t, pool, "+15550000040")
	b := testutil.CreateUser(t, pool, "+15550000041")

	id, err := repo.CreateMatch(ctx, a, b, "ptk_test")
	if err != nil {
		t.Fatalf("CreateMatch: %v", err)
	}

	m, err := repo.GetMatch(ctx, id)
	if err != nil {
		t.Fatalf("GetMatch: %v", err)
	}
	if m.Status != "active" || m.ChannelName != "ptk_test" {
		t.Fatalf("unexpected match: %+v", m)
	}

	// PartnerSummary returns a lightweight profile.
	if p, err := repo.PartnerSummary(ctx, a); err != nil || p.ID != a {
		t.Fatalf("PartnerSummary: %+v err=%v", p, err)
	}

	if err := repo.EndMatch(ctx, id); err != nil {
		t.Fatalf("EndMatch: %v", err)
	}
	if m, _ := repo.GetMatch(ctx, id); m.Status != "ended" {
		t.Error("match should be ended")
	}

	// Unknown match -> ErrNotFound.
	if _, err := repo.GetMatch(ctx, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
