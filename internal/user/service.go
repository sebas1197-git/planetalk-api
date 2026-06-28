// This file is the service: the business rules for the user module. It validates
// input and coordinates the repository. Handlers stay thin.
package user

import (
	"context"
	"errors"
)

// ErrCannotFriendSelf is returned when a user tries to friend themselves.
var ErrCannotFriendSelf = errors.New("you cannot add yourself")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ---- Profiles ----

func (s *Service) GetProfile(ctx context.Context, id string) (*Profile, error) {
	return s.repo.GetProfile(ctx, id)
}

// UpdateProfile applies the change, then returns the fresh profile.
func (s *Service) UpdateProfile(ctx context.Context, id string, in UpdateProfileInput) (*Profile, error) {
	if err := s.repo.UpdateProfile(ctx, id, in); err != nil {
		return nil, err
	}
	return s.repo.GetProfile(ctx, id)
}

// ---- Interests ----

func (s *Service) ListInterests(ctx context.Context) ([]Interest, error) {
	return s.repo.ListAllInterests(ctx)
}

func (s *Service) SetMyInterests(ctx context.Context, userID string, interestIDs []int) ([]Interest, error) {
	if err := s.repo.SetUserInterests(ctx, userID, interestIDs); err != nil {
		return nil, err
	}
	return s.repo.GetUserInterests(ctx, userID)
}

// ---- Friends ----

func (s *Service) SendFriendRequest(ctx context.Context, me, target string) error {
	if me == target {
		return ErrCannotFriendSelf
	}
	return s.repo.SendFriendRequest(ctx, me, target)
}

func (s *Service) AcceptFriendRequest(ctx context.Context, me, requester string) error {
	return s.repo.AcceptFriendRequest(ctx, me, requester)
}

func (s *Service) ListFriends(ctx context.Context, me string) ([]UserSummary, error) {
	return s.repo.ListFriends(ctx, me)
}

func (s *Service) ListFriendRequests(ctx context.Context, me string) ([]UserSummary, error) {
	return s.repo.ListIncomingRequests(ctx, me)
}

func (s *Service) RemoveFriend(ctx context.Context, me, other string) error {
	return s.repo.RemoveFriend(ctx, me, other)
}
