package captcha

import (
	"context"
	"errors"
	"time"

	"github.com/byteflowing/go-common/redis"
)

type Store interface {
	Save(ctx context.Context, key string, value string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, bool, error)
	Delete(ctx context.Context, key string) error
}

type RedisStore struct {
	rdb *redis.Redis
}

func NewRedisStore(rdb *redis.Redis) Store {
	return &RedisStore{rdb: rdb}
}

func (r *RedisStore) Save(ctx context.Context, key string, value string, expiration time.Duration) error {
	return r.rdb.Set(ctx, key, value, expiration).Err()
}

func (r *RedisStore) Get(ctx context.Context, key string) (string, bool, error) {
	result, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", false, nil
		}
		return "", false, err
	}
	return result, true, nil
}

func (r *RedisStore) Delete(ctx context.Context, key string) error {
	return r.rdb.Del(ctx, key).Err()
}
