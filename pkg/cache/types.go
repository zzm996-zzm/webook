package cache

import "time"

// Cache
type Cache[K comparable, V comparable] interface {
	Put(key K, val V, expiration time.Duration) error
	Get(key K) (V, bool)
	// Delete 删除
	// 第一个返回值是被删除的 key 对应的值
	// 第二个返回值是代表是否真的删除了
	Delete(k K) (V, bool)
	// Keys 返回所有的键
	// 注意，当你调用多次拿到的结果不一定相等
	// 取决于具体实现
	Keys() []K
	// Values 返回所有的值
	// 注意，当你调用多次拿到的结果不一定相等
	// 取决于具体实现
	Values() []V
	// 返回键值对数量
	Len() int64

	GetExpire(key string) (time.Time, bool)
}
