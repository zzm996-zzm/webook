package cache

import "sync"

/*
设计一个本地内存需要有什么功能
-	存储，并可以读、写；
	原子操作(线程安全)，如ConcurrentHashMap
	可以设置缓存的最大限制；
	超过最大限制有对应淘汰策略，如LRU、LFU
	过期时间淘汰，如定时、懒式、定期；
	持久化
	统计监控
*/

type Node[K comparable, V comparable] struct {
	Key   K
	Value V
	Prev  *Node[K, V]
	Next  *Node[K, V]
}

type LRUCache[K comparable, V comparable] struct {
	*sync.RWMutex
	Head  *Node[K, V]
	Tail  *Node[K, V]
	cache map[K]*Node[K, V]
	size  int
	cap   int
}

func (lru *LRUCache[K, V]) Delete(k K) (V, bool) {
	//TODO implement me
	panic("implement me")
}

func (lru *LRUCache[K, V]) Keys() []K {
	lru.RLock()
	defer lru.RUnlock()

	res := make([]K, 0, len(lru.cache))
	for k, _ := range lru.cache {
		res = append(res, k)
	}

	return res
}

func (lru *LRUCache[K, V]) Values() []V {
	res := make([]V, 0, len(lru.cache))
	for _, v := range lru.cache {
		res = append(res, v.Value)
	}
	return res
}

func (lru *LRUCache[K, V]) Len() int64 {
	return int64(len(lru.cache))
}

func NewLRUCache[K comparable, V comparable](capacity int) *LRUCache[K, V] {
	var key K
	var val V
	head := &Node[K, V]{key, val, nil, nil}
	tail := &Node[K, V]{key, val, nil, nil}
	head.Next = tail
	tail.Prev = head

	return &LRUCache[K, V]{cap: capacity, size: 0, cache: make(map[K]*Node[K, V]), Head: head, Tail: tail}
}

func (lru *LRUCache[K, V]) remove(node *Node[K, V]) *Node[K, V] {
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
	node.Next = nil
	node.Prev = nil
	return node

}

func (lru *LRUCache[K, V]) moveToFront(node *Node[K, V]) {
	lru.remove(node)
	lru.addToFront(node)
}

func (lru *LRUCache[K, V]) addToFront(node *Node[K, V]) {
	// 将当前节点的Next 换成 头节点的Next（即第一个数据）
	node.Next = lru.Head.Next
	lru.Head.Next.Prev = node

	lru.Head.Next = node
	node.Prev = lru.Head
}

func (lru *LRUCache[K, V]) removeTail() *Node[K, V] {

	tail := lru.Tail.Prev
	return lru.remove(tail)
}

func (lru *LRUCache[K, V]) Get(key K) (V, bool) {
	lru.Lock()
	defer lru.Unlock()
	var res V
	if res, has := lru.cache[key]; has {
		lru.moveToFront(res)
		return res.Value, true
	}

	return res, false
}

func (lru *LRUCache[K, V]) Put(key K, value V) error {
	//判断是否存在，存在则moveToronto
	lru.Lock()
	defer lru.Unlock()
	if vnode, has := lru.cache[key]; has {
		if vnode.Value != value {
			vnode.Value = value //修改
		}
		lru.moveToFront(vnode)
		return nil
	}

	//如果不存在则put,addToronto
	node := &Node[K, V]{Key: key, Value: value}
	lru.cache[key] = node
	lru.addToFront(node)
	lru.size++

	//超出容量了需要删除
	if lru.size > lru.cap {
		delNode := lru.removeTail()
		delete(lru.cache, delNode.Key)
		lru.size--
	}
	return nil
}
