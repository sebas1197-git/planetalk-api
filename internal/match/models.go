// Package match implements "Quick Match": pairing two waiting users into a 1v1
// video session, and notifying both in real time.
package match

import "time"

// Match is a paired session between two users.
type Match struct {
	ID          string     `json:"id"`
	UserA       string     `json:"user_a"`
	UserB       string     `json:"user_b"`
	ChannelName string     `json:"channel_name"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	EndedAt     *time.Time `json:"ended_at"`
}

// Partner is the lightweight view of the OTHER person in a match.
type Partner struct {
	ID          string  `json:"id"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

// EnterResult is returned from POST /match/enter.
//   - status "waiting": you're in the queue, wait for a "match_found" WS event.
//   - status "matched": you were paired immediately; details included.
type EnterResult struct {
	Status      string   `json:"status"`
	MatchID     string   `json:"match_id,omitempty"`
	ChannelName string   `json:"channel_name,omitempty"`
	Partner     *Partner `json:"partner,omitempty"`
}

// MatchFound is the payload of the realtime "match_found" event sent to BOTH
// users when a pairing happens.
type MatchFound struct {
	MatchID     string   `json:"match_id"`
	ChannelName string   `json:"channel_name"`
	Partner     *Partner `json:"partner"`
}
