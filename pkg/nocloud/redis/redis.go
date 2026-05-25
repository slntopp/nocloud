package redisdb

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

// Client is an interface wrapper for [redis.Client]
type Client interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	MGet(ctx context.Context, keys ...string) *redis.SliceCmd
	HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd
	Keys(ctx context.Context, pattern string) *redis.StringSliceCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	MSet(ctx context.Context, values ...interface{}) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Options() *redis.Options
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
	Ping(ctx context.Context) *redis.StatusCmd
}

type client struct {
	*redis.Client
}

func NewRedisClient(c *redis.Client) Client {
	return &client{
		Client: c,
	}
}

func RedisUnderlying(c Client) (*redis.Client, bool) {
	switch v := c.(type) {
	case *client:
		if v != nil && v.Client != nil {
			return v.Client, true
		}
	case *redis.Client:
		return v, true
	}
	return nil, false
}
