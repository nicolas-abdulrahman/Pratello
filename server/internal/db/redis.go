package db

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisIdempotency caches completed operations. Callers must retain a durable
// database check and only mark a key after committing the operation.
type RedisIdempotency struct{ client *redis.Client }

func OpenRedis(rawURL string) (*RedisIdempotency, error) {
	opts, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, err
	}
	opts.ContextTimeoutEnabled = true
	return &RedisIdempotency{redis.NewClient(opts)}, nil
}

func (r *RedisIdempotency) Close() error { return r.client.Close() }

func (r *RedisIdempotency) Completed(ctx context.Context, key string) bool {
	ctx, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()
	n, err := r.client.Exists(ctx, key).Result()
	return err == nil && n == 1
}

func (r *RedisIdempotency) MarkCompleted(ctx context.Context, key string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 150*time.Millisecond)
	defer cancel()
	_ = r.client.SetNX(ctx, key, "1", 24*time.Hour).Err()
}
