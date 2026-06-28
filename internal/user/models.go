// Package user handles profiles, interests, and friends ("V-friends").
// This file defines the data shapes returned by the API.
package user

import "time"

// Interest is one item from the fixed interests list (e.g. "Music").
type Interest struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Profile is a user's full profile. Pointer fields (*string) can be null,
// meaning "not set yet" — that's how we tell "empty" apart from "missing".
type Profile struct {
	ID          string     `json:"id"`
	Phone       string     `json:"phone"`
	DisplayName *string    `json:"display_name"`
	Bio         *string    `json:"bio"`
	AvatarURL   *string    `json:"avatar_url"`
	Gender      *string    `json:"gender"`
	Birthdate   *string    `json:"birthdate"` // "YYYY-MM-DD"
	Country     *string    `json:"country"`
	Language    *string    `json:"language"`
	CreatedAt   time.Time  `json:"created_at"`
	LastSeen    *time.Time `json:"last_seen"`
	Interests   []Interest `json:"interests"`
}

// UserSummary is a lightweight public view (used in friend lists).
type UserSummary struct {
	ID          string  `json:"id"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Country     *string `json:"country"`
}
