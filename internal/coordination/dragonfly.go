package coordination

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Dragonfly struct {
	client *redis.Client
}

func OpenDragonfly(addr string) *Dragonfly {
	return &Dragonfly{client: redis.NewClient(&redis.Options{Addr: addr})}
}

func (d *Dragonfly) Close() error {
	if d == nil || d.client == nil {
		return nil
	}
	return d.client.Close()
}

func (d *Dragonfly) Ping(ctx context.Context) error {
	if err := d.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping dragonfly: %w", err)
	}
	return nil
}

func (d *Dragonfly) AcquireLease(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	ok, err := d.client.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("acquire dragonfly lease: %w", err)
	}
	return ok, nil
}

func (d *Dragonfly) ReleaseLease(ctx context.Context, key string) error {
	if err := d.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("release dragonfly lease: %w", err)
	}
	return nil
}
