// Package redis owns the connection to Redis.
//
// Redis is our fast in-memory store. We use it for:
//   - presence  (who is online right now)
//   - pub/sub   (fan out realtime events across API instances)
//   - later: the matchmaking queue
//
// Note: we alias the go-redis import as `goredis` so it doesn't clash with this
// package's own name ("redis").
package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

// Connect parses the REDIS_URL, opens a client, and pings to confirm it works.
func Connect(ctx context.Context, url string) (*goredis.Client, error) {
	opt, err := goredis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := goredis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}
