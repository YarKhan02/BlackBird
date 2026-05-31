package redis

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"
)

type Blocklist struct {
	client *redislib.Client
	prefix string
}

func NewBlocklist(redisURL string) (*Blocklist, error) {
	opts, err := redislib.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redislib.NewClient(opts)
	return &Blocklist{client: client, prefix: "bl:jti:"}, nil
}

func (b *Blocklist) Close() error {
	return b.client.Close()
}

func (b *Blocklist) Add(ctx context.Context, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return b.client.Set(ctx, b.prefix+jti, "1", ttl).Err()
}

func (b *Blocklist) Contains(ctx context.Context, jti string) (bool, error) {
	res, err := b.client.Exists(ctx, b.prefix+jti).Result()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}
