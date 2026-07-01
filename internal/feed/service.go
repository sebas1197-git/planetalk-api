// The feed service: create posts, delete your own, and list a user's posts with
// the correct visibility for whoever is looking.
package feed

import (
	"context"
	"errors"
)

// ErrInvalidVisibility is returned for a visibility value we don't allow.
var ErrInvalidVisibility = errors.New("visibility must be public, friends, or private")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput is what a caller provides to make a post.
type CreateInput struct {
	MediaURL   *string
	MediaType  *string
	Caption    *string
	Visibility string
}

// Create makes a new post for the author.
func (s *Service) Create(ctx context.Context, authorID string, in CreateInput) (*Post, error) {
	if in.Visibility == "" {
		in.Visibility = "public"
	}
	if in.Visibility != "public" && in.Visibility != "friends" && in.Visibility != "private" {
		return nil, ErrInvalidVisibility
	}
	p := &Post{
		UserID:     authorID,
		MediaURL:   in.MediaURL,
		MediaType:  in.MediaType,
		Caption:    in.Caption,
		Visibility: in.Visibility,
	}
	if err := s.repo.CreatePost(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete removes a post if the caller owns it.
func (s *Service) Delete(ctx context.Context, ownerID, postID string) error {
	return s.repo.DeletePost(ctx, postID, ownerID)
}

// ListUserPosts returns a page of targetID's posts, filtered to what viewerID is
// allowed to see:
//   - own posts       -> everything
//   - accepted friend -> public + friends
//   - anyone else     -> public only
func (s *Service) ListUserPosts(ctx context.Context, viewerID, targetID string, limit, offset int) ([]Post, int, error) {
	visibilities, err := s.allowedVisibilities(ctx, viewerID, targetID)
	if err != nil {
		return nil, 0, err
	}
	posts, err := s.repo.ListByUser(ctx, targetID, visibilities, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountByUser(ctx, targetID, visibilities)
	if err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (s *Service) allowedVisibilities(ctx context.Context, viewerID, targetID string) ([]string, error) {
	if viewerID == targetID {
		return []string{"public", "friends", "private"}, nil
	}
	friends, err := s.repo.AreFriends(ctx, viewerID, targetID)
	if err != nil {
		return nil, err
	}
	if friends {
		return []string{"public", "friends"}, nil
	}
	return []string{"public"}, nil
}
