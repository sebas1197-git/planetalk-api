// Package config loads all settings from environment variables into one struct,
// so the rest of the app never reads os.Getenv directly.
//
// Pattern to follow:
//   - Define a Config struct with one field per setting.
//   - Load() reads the env (use os.Getenv, or a helper like caarlos0/env),
//     applies defaults, validates required values, and returns the struct.
//   - Call Load() once in main() and pass Config (or sub-parts) down.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds every runtime setting. Group fields to match the .env.example file.
type Config struct {
	// Server
	AppEnv   string // "development" | "production"
	HTTPPort string // e.g. "8080"

	// Datastores
	DatabaseURL string
	RedisURL    string

	// JWT
	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	// Twilio (OTP) — empty values => use the console stub in dev
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFrom       string

	// Google Translate — empty => use the stub
	GoogleCredentialsPath  string
	GoogleTranslateProject string

	// Agora (video)
	AgoraAppID    string
	AgoraAppCert  string
	AgoraTokenTTL time.Duration
}

// Load reads configuration from the environment and returns it.
func Load() (*Config, error) {
	// Load a .env file if one exists (handy in dev). Ignore the error: in
	// production there's usually no .env file, real en vars are set instead.

	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:   getenv("APP_ENV", "development"),
		HTTPPort: getenv("HTTP_PORT", "8080"),

		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),

		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTAccessTTL:  getDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL: getDuration("JWT_REFRESH_TTL", 720*time.Hour),

		TwilioAccountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
		TwilioFrom:       os.Getenv("TWILIO_FROM_NUMBER"),

		GoogleCredentialsPath:  os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		GoogleTranslateProject: os.Getenv("GOOGLE_TRANSLATE_PROJECT_ID"),

		AgoraAppID:    os.Getenv("AGORA_APP_ID"),
		AgoraAppCert:  os.Getenv("AGORA_APP_CERTIFICATE"),
		AgoraTokenTTL: getDuration("AGORA_TOKEN_TTL", time.Hour),
	}

	// Fail fast if a value we can't run without missing

	var missing []string
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.RedisURL == "" {
		missing = append(missing, "REDIS_URL")
	}
	if cfg.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

// getenv returns the env var, or a fallback if it's empty.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getDuration parses a duration like "15m" or "720h"; falls back if unset/invalid.
func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
