package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
}

type InteractiveRedisCache struct {
	client redis.Cmdable
}

func (i InteractiveRedisCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	//TODO implement me
	panic("implement me")
}

func NewInteractiveRedisCache(client redis.Cmdable) InteractiveCache {
	return &InteractiveRedisCache{client: client}
}
