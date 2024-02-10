package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
	"webook/pkg/cache"
)

var ErrKeyNotExist = errors.New("key不存在")

// LocalCodeCache 本地缓存实现
type LocalCodeCache struct {
	cache      cache.Cache[string, any]
	expiration time.Duration
}

type codeItem struct {
	code string
	// 可验证次数
	cnt int
}

func NewLocalCodeCache(c cache.Cache[string, any], expiration time.Duration) *LocalCodeCache {
	return &LocalCodeCache{
		cache:      c,
		expiration: expiration,
	}
}

func (l LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	key := l.key(biz, phone)
	// 先获取验证码
	_, ok := l.cache.Get(key)
	if !ok {
		// 获取失败，则插入验证码
		ci := &codeItem{
			code: code,
			cnt:  3,
		}
		err := l.cache.Put(key, ci, l.expiration)

		if err != nil {
			return err
		}
		return nil
	}

	//没有过期，并且过期时间大于9分钟 则直接返回错误
	now := time.Now()
	if expire, expired := l.cache.GetExpire(key); !expired && expire.Sub(now) > time.Minute*9 {
		// 不到一分钟
		return ErrCodeSendTooMany
	}

	// 重新发送code
	err := l.cache.Put(key, &codeItem{
		code: code,
		cnt:  3,
	}, l.expiration)
	if err != nil {
		return err
	}

	return nil
}

func (l LocalCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	key := l.key(biz, phone)
	val, ok := l.cache.Get(key)
	if !ok {
		// 都没发验证码
		return false, ErrKeyNotExist
	}
	itm, ok := val.(*codeItem)
	if !ok {
		// 理论上来说这是不可能的
		return false, errors.New("系统错误")
	}
	if itm.cnt <= 0 {
		return false, ErrCodeVerifyTooMany
	}
	//单纯验证失败 减少cnt
	if itm.code != inputCode {
		itm.cnt--

		return false, nil
	}

	itm.cnt = 0
	return true, nil
}

func (l LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
