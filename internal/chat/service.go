// The chat service: send a message (translate for the recipient + deliver live)
// and list a conversation's history. Only match participants can do either.
package chat

import (
	"context"

	"github.com/sebas1197-git/planetalk/internal/match"
	"github.com/sebas1197-git/planetalk/internal/realtime"
	"github.com/sebas1197-git/planetalk/pkg/translate"
)

type Service struct {
	repo    *Repository
	matches *match.Service // verifies participation + tells us who the recipient is
	tr      translate.Translator
	hub     *realtime.Hub
}

func NewService(repo *Repository, matches *match.Service, tr translate.Translator, hub *realtime.Hub) *Service {
	return &Service{repo: repo, matches: matches, tr: tr, hub: hub}
}

// Send stores a message, translates it into the recipient's language, delivers
// it live over WebSocket, and returns the saved message.
func (s *Service) Send(ctx context.Context, senderID, matchID, body string) (*Message, error) {
	m, err := s.matches.Get(ctx, senderID, matchID) // ErrNotFound / ErrForbidden
	if err != nil {
		return nil, err
	}

	// Figure out who receives this message.
	recipient := m.UserA
	if senderID == m.UserA {
		recipient = m.UserB
	}

	msg := &Message{MatchID: matchID, SenderID: senderID, Body: body}

	// Translate into the recipient's language (if they set one).
	if lang, _ := s.repo.GetUserLanguage(ctx, recipient); lang != "" {
		if res, err := s.tr.Translate(ctx, body, lang); err == nil {
			text, dst := res.Text, lang
			msg.TranslatedBody, msg.DstLang = &text, &dst
			if res.SourceLang != "" {
				src := res.SourceLang
				msg.SrcLang = &src
			}
		}
		// If translation fails, we still deliver the original text.
	}

	if err := s.repo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	// Push the new message to the recipient in real time.
	_ = s.hub.SendToUser(ctx, recipient, realtime.Event{Type: "message", Data: msg})

	return msg, nil
}

// List returns a page of the conversation plus the total count.
func (s *Service) List(ctx context.Context, userID, matchID string, limit, offset int) ([]Message, int, error) {
	if _, err := s.matches.Get(ctx, userID, matchID); err != nil {
		return nil, 0, err
	}
	msgs, err := s.repo.ListMessages(ctx, matchID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountMessages(ctx, matchID)
	if err != nil {
		return nil, 0, err
	}
	return msgs, total, nil
}
