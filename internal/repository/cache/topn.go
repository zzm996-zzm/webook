package cache

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/pkg/cache"
)

// TopN 点赞数，阅读数，收藏数通用接口
type TopN[T comparable] interface {
	GetLocal(ctx context.Context) ([]T, error)
	GetCache(ctx context.Context, key string) ([]T, error)

	SetLocal(ctx context.Context, key string, val T, duration time.Duration) error
	SetCache()
}

type InteractiveTopN[T comparable] struct {
	client redis.Cmdable
	local  cache.Cache[string, T]
	n      int64
}

// 回写本地缓存
func (i *InteractiveTopN[T]) SetLocal(ctx context.Context, key string, val T, duration time.Duration) error {
	return i.local.Put(key, val, duration)
}

func (i *InteractiveTopN[T]) SetCache() {
	// 将数据同步到redis 包括zset 以及 hash
	// 使用lua脚本
}

func NewInteractiveTopN[T comparable](client redis.Cmdable, local cache.Cache[string, T], n int64) TopN[T] {
	return &InteractiveTopN[T]{
		client: client,
		local:  local,
		n:      n,
	}
}

// GetLocal 从本地缓存获取
func (i *InteractiveTopN[T]) GetLocal(ctx context.Context) ([]T, error) {
	// TODO:本地缓存查询可能需要排序，建议从外部传入
	return i.local.Values(), nil
}

// GetCache 从redis中获取 默认倒序返回即可
func (i *InteractiveTopN[T]) GetCache(ctx context.Context, key string) ([]T, error) {
	var res []T
	members, err := i.client.ZRevRange(ctx, key, 0, -1).Result()

	if err != nil {
		return res, err
	}

	// 再从hash中获取所有结果
	for _, key := range members {
		result, err := i.client.HGetAll(ctx, key).Result()
		if err != nil {
			// 单个查询不出来，其实不代表后续的有问题
			// 进一步判断什么问题，如果是key not exist 则continue
			return res, err
		}
		var t T
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:           &t,
			TagName:          "json",
			ErrorUnused:      true,
			WeaklyTypedInput: true,
		})

		if err != nil {
			return res, err
		}

		err = decoder.Decode(result)

		if err != nil {
			return res, err
		}

		// 加入结果集
		res = append(res, t)
	}
	return res, nil
}
