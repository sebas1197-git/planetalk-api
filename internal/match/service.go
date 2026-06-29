// The match service runs "Quick Match": a Redis-backed waiting queue plus the
// pairing logic that creates a match and notifies both users over WebSocket.
package match

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	goredis "github.com/redis/go-redis/v9"

	"github.com/sebas1197-git/planetalk/internal/realtime"
)

// queueKey is the Redis list of user ids currently waiting to be matched.
const queueKey = "matchmaking:queue"

// ErrForbidden is returned when a user touches a match they're not part of.
var ErrForbidden = errors.New("not a participant in this match")

type Service struct {
	repo *Repository
	rdb  *goredis.Client
	hub  *realtime.Hub
}

func NewService(repo *Repository, rdb *goredis.Client, hub *realtime.Hub) *Service {
	return &Service{repo: repo, rdb: rdb, hub: hub}
}

// Enter puts the user in the queue, OR pairs them with someone already waiting.
//
// The queue is a Redis list. LPOP is atomic, so two people entering at the same
// time can't grab the same partner twice.
func (s *Service) Enter(ctx context.Context, userID string) (*EnterResult, error) {
	// Remove any stale copy of me first (e.g. I tapped "match" twice).
	s.rdb.LRem(ctx, queueKey, 0, userID)

	// Try to take the next waiting user.
	partnerID, err := s.rdb.LPop(ctx, queueKey).Result()
	if errors.Is(err, goredis.Nil) || partnerID == "" || partnerID == userID {
		// Nobody waiting (or only me) -> join the queue and wait.
		if err := s.rdb.RPush(ctx, queueKey, userID).Err(); err != nil {
			return nil, err
		}
		return &EnterResult{Status: "waiting"}, nil
	}
	if err != nil {
		return nil, err
	}

	// We have a partner -> create the match.
	channel := generateChannel()
	matchID, err := s.repo.CreateMatch(ctx, partnerID, userID, channel)
	if err != nil {
		return nil, err
	}

	// Notify BOTH users over WebSocket, each getting the OTHER as "partner".
	s.notify(ctx, partnerID, matchID, channel, userID)  // waiting user hears about the enterer
	s.notify(ctx, userID, matchID, channel, partnerID)  // enterer hears about the waiting user

	// Also return the result directly to the caller (the enterer).
	partner, _ := s.repo.PartnerSummary(ctx, partnerID)
	return &EnterResult{
		Status:      "matched",
		MatchID:     matchID,
		ChannelName: channel,
		Partner:     partner,
	}, nil
}

// Leave removes the user from the queue (e.g. they cancelled).
func (s *Service) Leave(ctx context.Context, userID string) error {
	return s.rdb.LRem(ctx, queueKey, 0, userID).Err()
}

// Get returns a match, but only if the caller is one of its participants.
func (s *Service) Get(ctx context.Context, userID, matchID string) (*Match, error) {
	m, err := s.repo.GetMatch(ctx, matchID)
	if err != nil {
		return nil, err
	}
	if userID != m.UserA && userID != m.UserB {
		return nil, ErrForbidden
	}
	return m, nil
}

// End finishes a match and tells the other participant it ended.
func (s *Service) End(ctx context.Context, userID, matchID string) error {
	m, err := s.repo.GetMatch(ctx, matchID)
	if err != nil {
		return err
	}
	if userID != m.UserA && userID != m.UserB {
		return ErrForbidden
	}
	if err := s.repo.EndMatch(ctx, matchID); err != nil {
		return err
	}
	other := m.UserA
	if userID == m.UserA {
		other = m.UserB
	}
	_ = s.hub.SendToUser(ctx, other, realtime.Event{
		Type: "match_ended",
		Data: map[string]string{"match_id": matchID},
	})
	return nil
}

// notify sends a "match_found" event to `toUser`, describing `partnerID`.
func (s *Service) notify(ctx context.Context, toUser, matchID, channel, partnerID string) {
	partner, _ := s.repo.PartnerSummary(ctx, partnerID)
	_ = s.hub.SendToUser(ctx, toUser, realtime.Event{
		Type: "match_found",
		Data: MatchFound{MatchID: matchID, ChannelName: channel, Partner: partner},
	})
}

// generateChannel returns a unique, hard-to-guess Agora channel name.
func generateChannel() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return "ptk_" + hex.EncodeToString(b)
}
