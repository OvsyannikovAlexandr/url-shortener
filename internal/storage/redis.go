package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(addr string) *RedisStorage {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisStorage{client: rdb}
}

func (s *RedisStorage) SaveURL(ctx context.Context, key, url string, ttl time.Duration) error {
	return s.client.Set(ctx, key, url, ttl).Err()
}

func (s *RedisStorage) GetURL(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

func (s *RedisStorage) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *RedisStorage) IncrementClicks(ctx context.Context, key string) error {
	return s.client.Incr(ctx, "clicks:"+key).Err()
}

func (s *RedisStorage) GetClicks(ctx context.Context, key string) (int64, error) {
	return s.client.Get(ctx, "clicks:"+key).Int64()
}
