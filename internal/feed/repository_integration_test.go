//go:build integration

package feed

import (
	"context"
	"testing"

	"github.com/sebas1197-git/planetalk/internal/testutil"
)

func strptr(s string) *string { return &s }

func TestFeedRepository_Visibility(t *testing.T) {
	pool := testutil.PostgresDB(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	a := testutil.CreateUser(t, pool, "+15550000020")

	for _, vis := range []string{"public", "friends", "private"} {
		if err := repo.CreatePost(ctx, &Post{UserID: a, Visibility: vis, Caption: strptr(vis)}); err != nil {
			t.Fatalf("CreatePost(%s): %v", vis, err)
		}
	}

	pub, _ := repo.ListByUser(ctx, a, []string{"public"}, 10, 0)
	if len(pub) != 1 {
		t.Errorf("public-only: expected 1, got %d", len(pub))
	}
	friend, _ := repo.ListByUser(ctx, a, []string{"public", "friends"}, 10, 0)
	if len(friend) != 2 {
		t.Errorf("public+friends: expected 2, got %d", len(friend))
	}
	all, _ := repo.ListByUser(ctx, a, []string{"public", "friends", "private"}, 10, 0)
	if len(all) != 3 {
		t.Errorf("owner: expected 3, got %d", len(all))
	}
	if n, _ := repo.CountByUser(ctx, a, []string{"public"}); n != 1 {
		t.Errorf("CountByUser public: expected 1, got %d", n)
	}
}

func TestFeedRepository_Delete(t *testing.T) {
	pool := testutil.PostgresDB(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	a := testutil.CreateUser(t, pool, "+15550000021")
	b := testutil.CreateUser(t, pool, "+15550000022")

	p := &Post{UserID: a, Visibility: "public", Caption: strptr("hi")}
	if err := repo.CreatePost(ctx, p); err != nil {
		t.Fatal(err)
	}

	// B can't delete A's post.
	if err := repo.DeletePost(ctx, p.ID, b); err != ErrNotFound {
		t.Errorf("non-owner delete: expected ErrNotFound, got %v", err)
	}
	// A can.
	if err := repo.DeletePost(ctx, p.ID, a); err != nil {
		t.Errorf("owner delete: %v", err)
	}
}
