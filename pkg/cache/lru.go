package cache

import (
	"errors"
	"sync"
	"time"
)

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

type node[K comparable, V comparable] struct {
	key    K
	value  V
	expire time.Time
	//是否有过期时间 true 可以过期 false 不可过期
	canExpire bool
	prev      *node[K, V]
	next      *node[K, V]
}

type LRUCache[K comparable, V comparable] struct {
	*sync.RWMutex
	c     chan node[K, V]
	head  *node[K, V]
	tail  *node[K, V]
	cache map[K]*node[K, V]
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
		res = append(res, v.value)
	}
	return res
}

func (lru *LRUCache[K, V]) Len() int64 {
	return int64(len(lru.cache))
}

func NewLRUCache[K comparable, V comparable](capacity int) *LRUCache[K, V] {
	var key K
	var val V

	head := &node[K, V]{key: key, value: val, expire: time.Time{}, next: nil, prev: nil}
	tail := &node[K, V]{key: key, value: val, expire: time.Time{}, next: nil, prev: nil}
	head.next = tail
	tail.prev = head

	return &LRUCache[K, V]{cap: capacity, size: 0, cache: make(map[K]*node[K, V]), head: head, tail: tail, RWMutex: &sync.RWMutex{}}
}

func (lru *LRUCache[K, V]) remove(node *node[K, V]) *node[K, V] {
	node.prev.next = node.next
	node.next.prev = node.prev
	node.next = nil
	node.prev = nil
	return node

}

func (lru *LRUCache[K, V]) moveToFront(node *node[K, V]) {
	lru.remove(node)
	lru.addToFront(node)
}

func (lru *LRUCache[K, V]) addToFront(node *node[K, V]) {
	// 将当前节点的Next 换成 头节点的Next（即第一个数据）
	node.next = lru.head.next
	lru.head.next.prev = node

	lru.head.next = node
	node.prev = lru.head
}

func (lru *LRUCache[K, V]) removeTail() *node[K, V] {

	tail := lru.tail.prev
	return lru.remove(tail)
}

// getExpire 设置节点的过期时间
func (n *node[K, V]) setExpire(expiration time.Duration) {
	if expiration <= 0 {
		// 不可过期
		n.canExpire = false
		return
	}

	n.canExpire = true
	n.expire = time.Now().Add(expiration)
}

// getExpire 判断节点是否过期时间
func (n *node[K, V]) expired() bool {
	// 没有过期时间，直接返回未过期
	if !n.canExpire {
		return false
	}
	now := time.Now()
	// 节点过期了
	if n.expire.Sub(now) <= 0 {
		return true
	}

	return false
}

func (lru *LRUCache[K, V]) GetExpire(key K) (time.Time, bool) {
	lru.RLock()
	defer lru.RUnlock()
	return lru.cache[key].expire, lru.cache[key].expired()
}

func (lru *LRUCache[K, V]) Get(key K) (V, bool) {
	lru.Lock()
	defer lru.Unlock()

	// lazy删除 查询的时候判断过期时间 然后实现删除
	var res V
	if res, has := lru.cache[key]; has {
		// 如果没过期了则返回
		if !res.expired() {
			lru.moveToFront(res)
			return res.value, true
		}

		// 过期了则删除节点 以及map中的数据
		lru.remove(res)
		delete(lru.cache, res.key)
	}

	return res, false
}

func (lru *LRUCache[K, V]) Put(key K, value V, expiration time.Duration) error {
	//判断是否存在，存在则moveToronto
	lru.Lock()
	defer lru.Unlock()
	if vnode, has := lru.cache[key]; has {
		if vnode != nil {
			return errors.New("node is nil")
		}
		//如果不相等则更新value即可
		if vnode.value != value {
			vnode.value = value
		}
		// 刷新过期时间
		vnode.setExpire(expiration)
		lru.moveToFront(vnode)
		return nil

	}

	//如果不存在则put,addToronto,并且设置过期时间
	node := &node[K, V]{key: key, value: value, expire: time.Now().Add(expiration)}
	node.setExpire(expiration)
	lru.cache[key] = node
	lru.addToFront(node)
	lru.size++

	//超出容量了需要删除
	if lru.size > lru.cap {
		delNode := lru.removeTail()
		delete(lru.cache, delNode.key)
		lru.size--
	}
	return nil
}
