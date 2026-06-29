// Package agora mints short-lived RTC tokens so a client can join a video
// channel. SECURITY: the App Certificate is the secret that signs tokens — it
// stays here on the server and is NEVER sent to clients. Clients only ever get
// a finished token (which expires).
package agora

import (
	"errors"
	"time"

	rtctokenbuilder "github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/rtctokenbuilder2"
)

// ErrNotConfigured means no Agora credentials are set (e.g. local dev without
// an Agora account). The handler turns this into a clear 503.
var ErrNotConfigured = errors.New("agora app id/certificate not configured")

type TokenBuilder struct {
	appID   string
	appCert string
	ttl     time.Duration
}

func New(appID, appCert string, ttl time.Duration) *TokenBuilder {
	if ttl <= 0 {
		ttl = time.Hour
	}
	return &TokenBuilder{appID: appID, appCert: appCert, ttl: ttl}
}

// Enabled reports whether real Agora credentials are present.
func (b *TokenBuilder) Enabled() bool { return b.appID != "" && b.appCert != "" }

func (b *TokenBuilder) AppID() string   { return b.appID }
func (b *TokenBuilder) TTLSeconds() int { return int(b.ttl.Seconds()) }

// BuildRTCToken returns a publisher token for the given channel + uid.
func (b *TokenBuilder) BuildRTCToken(channel string, uid uint32) (string, error) {
	if !b.Enabled() {
		return "", ErrNotConfigured
	}
	expire := uint32(b.ttl.Seconds())
	// tokenExpire and privilegeExpire both use the same TTL.
	return rtctokenbuilder.BuildTokenWithUid(
		b.appID, b.appCert, channel, uid,
		rtctokenbuilder.RolePublisher, expire, expire,
	)
}
