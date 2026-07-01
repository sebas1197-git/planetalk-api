// Package moderation handles user reports and blocks. Blocks are enforced in
// matchmaking and chat (a blocked pair can't be matched or message each other).
package moderation

// BlockedUser is the lightweight view of someone you've blocked.
type BlockedUser struct {
	ID          string  `json:"id"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}
