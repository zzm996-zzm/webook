package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/internal/domain"
)

type UserCache interface {
	Get(ctx context.Context, uid int64) (domain.User, error)
	Set(ctx context.Context, du domain.User) error
}

type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd: cmd,
		// 15分钟过期user数据
		expiration: time.Minute * 15,
	}
}

func (cache *RedisUserCache) key(uid int64) string {
	return fmt.Sprintf("user-info-%d", uid)
}

func (cache *RedisUserCache) Get(ctx context.Context, uid int64) (domain.User, error) {
	key := cache.key(uid)
	data, err := cache.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (cache *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	data, err := json.Marshal(&u)
	if err != nil {
		return err
	}
	key := cache.key(u.Id)

	return cache.cmd.Set(ctx, key, data, cache.expiration).Err()
}
