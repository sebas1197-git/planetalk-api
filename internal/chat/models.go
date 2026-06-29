// Package chat stores messages within a match and delivers them live, with the
// message translated into the recipient's language.
package chat

import "time"

// Message is one chat message.
//   - Body is the original text the sender typed.
//   - TranslatedBody is that text in the recipient's language (nil if not translated).
type Message struct {
	ID             string    `json:"id"`
	MatchID        string    `json:"match_id"`
	SenderID       string    `json:"sender_id"`
	Body           string    `json:"body"`
	TranslatedBody *string   `json:"translated_body"`
	SrcLang        *string   `json:"src_lang"`
	DstLang        *string   `json:"dst_lang"`
	CreatedAt      time.Time `json:"created_at"`
}
