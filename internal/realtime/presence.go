// Presence tracking: who is online right now, stored in a Redis SET.
// A user is "online" while they have at least one open WebSocket connection.
package realtime

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

const onlineSetKey = "planetalk:online"

func markOnline(ctx context.Context, rdb *goredis.Client, userID string) error {
	return rdb.SAdd(ctx, onlineSetKey, userID).Err()
}

func markOffline(ctx context.Context, rdb *goredis.Client, userID string) error {
	return rdb.SRem(ctx, onlineSetKey, userID).Err()
}

// IsOnline reports whether a user currently has any live connection.
func IsOnline(ctx context.Context, rdb *goredis.Client, userID string) (bool, error) {
	return rdb.SIsMember(ctx, onlineSetKey, userID).Result()
}

// OnlineUsers returns all currently-online user ids.
func OnlineUsers(ctx context.Context, rdb *goredis.Client) ([]string, error) {
	return rdb.SMembers(ctx, onlineSetKey).Result()
}
