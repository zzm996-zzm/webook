package cache

import "time"

type LocalCache[K comparable, V comparable] struct {
	lru *LRUCache[K, V]
	// 过期时间
	expiration time.Duration
}
