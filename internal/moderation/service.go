// The moderation service: report a user, block/unblock, list blocks, and the
// IsBlocked check that matchmaking + chat use to enforce blocks.
package moderation

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrSelf         = errors.New("you cannot do that to yourself")
	ErrReasonEmpty  = errors.New("a reason is required")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Report files a report against another user.
func (s *Service) Report(ctx context.Context, reporterID, reportedID, reason string, context_ *string) error {
	if reporterID == reportedID {
		return ErrSelf
	}
	if strings.TrimSpace(reason) == "" {
		return ErrReasonEmpty
	}
	return s.repo.CreateReport(ctx, reporterID, reportedID, reason, context_)
}

// Block blocks another user.
func (s *Service) Block(ctx context.Context, blockerID, blockedID string) error {
	if blockerID == blockedID {
		return ErrSelf
	}
	return s.repo.AddBlock(ctx, blockerID, blockedID)
}

// Unblock removes a block.
func (s *Service) Unblock(ctx context.Context, blockerID, blockedID string) error {
	return s.repo.RemoveBlock(ctx, blockerID, blockedID)
}

// ListBlocked returns who the user has blocked.
func (s *Service) ListBlocked(ctx context.Context, blockerID string) ([]BlockedUser, error) {
	return s.repo.ListBlocked(ctx, blockerID)
}

// IsBlocked satisfies the block-checker interfaces used by match and chat.
func (s *Service) IsBlocked(ctx context.Context, a, b string) (bool, error) {
	return s.repo.IsBlocked(ctx, a, b)
}
