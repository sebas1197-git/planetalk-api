// Package feed handles posts — a user's "personal homepage" content, with
// per-post visibility (public / friends / private).
package feed

import "time"

type Post struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	MediaURL   *string   `json:"media_url"`
	MediaType  *string   `json:"media_type"`
	Caption    *string   `json:"caption"`
	Visibility string    `json:"visibility"`
	CreatedAt  time.Time `json:"created_at"`
}
