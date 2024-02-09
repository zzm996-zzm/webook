package cache

// Cache TODO: 改造泛型
type Cache interface {
	Put(key int, value int)
	Get(key int) int
}
