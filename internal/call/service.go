// The call service mints an Agora token for a match's channel, but only for a
// user who actually belongs to that match.
package call

import (
	"context"
	"hash/crc32"

	"github.com/sebas1197-git/planetalk/internal/match"
	"github.com/sebas1197-git/planetalk/pkg/agora"
)

type Service struct {
	matches *match.Service // reused to verify participation + read the channel
	agora   *agora.TokenBuilder
}

func NewService(matches *match.Service, ag *agora.TokenBuilder) *Service {
	return &Service{matches: matches, agora: ag}
}

// Token returns a join token for the match's video channel.
// It relies on match.Service.Get, which returns ErrNotFound / ErrForbidden so
// non-participants can't get a token.
func (s *Service) Token(ctx context.Context, userID, matchID string) (*TokenResponse, error) {
	m, err := s.matches.Get(ctx, userID, matchID)
	if err != nil {
		return nil, err
	}

	uid := uidFromUser(userID)
	token, err := s.agora.BuildRTCToken(m.ChannelName, uid)
	if err != nil {
		return nil, err // includes agora.ErrNotConfigured when creds are missing
	}

	return &TokenResponse{
		AppID:            s.agora.AppID(),
		ChannelName:      m.ChannelName,
		UID:              uid,
		Token:            token,
		ExpiresInSeconds: s.agora.TTLSeconds(),
	}, nil
}

// uidFromUser turns a user's UUID into a stable non-zero uint32 (Agora uids are
// numeric). Same user -> same uid every time.
func uidFromUser(userID string) uint32 {
	h := crc32.ChecksumIEEE([]byte(userID))
	if h == 0 {
		h = 1 // 0 means "let Agora assign"; we want a stable id
	}
	return h
}
