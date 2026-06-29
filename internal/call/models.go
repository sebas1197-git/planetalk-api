// Package call issues video tokens for a match's channel.
package call

// TokenResponse is what the client needs to join the Agora video channel.
type TokenResponse struct {
	AppID            string `json:"app_id"`
	ChannelName      string `json:"channel_name"`
	UID              uint32 `json:"uid"`   // the numeric id this user joins with
	Token            string `json:"token"` // short-lived, signed by the server
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}
