package cache

type LocalCache[K comparable, V comparable] struct {
	lru *LRUCache[K, V]
}
