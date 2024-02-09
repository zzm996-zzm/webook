package cache

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

type Node struct {
	Key   int
	Value int
	Prev  *Node
	Next  *Node
}

type LRUCache struct {
	Head  *Node
	Tail  *Node
	cache map[int]*Node
	size  int
	cap   int
}

func NewLRUCache(capacity int) LRUCache {
	head := &Node{-1, -1, nil, nil}
	tail := &Node{-1, -1, nil, nil}
	head.Next = tail
	tail.Prev = head

	return LRUCache{cap: capacity, size: 0, cache: make(map[int]*Node), Head: head, Tail: tail}
}

func remove(node *Node) *Node {
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
	node.Next = nil
	node.Prev = nil
	return node

}

func (lru *LRUCache) moveToFront(node *Node) {
	remove(node)
	lru.addToFront(node)
}

func (lru *LRUCache) addToFront(node *Node) {
	// 将当前节点的Next 换成 头节点的Next（即第一个数据）
	node.Next = lru.Head.Next
	lru.Head.Next.Prev = node

	lru.Head.Next = node
	node.Prev = lru.Head
}

func (lru *LRUCache) removeTail() *Node {
	tail := lru.Tail.Prev
	return remove(tail)
}

func (lru *LRUCache) Get(key int) int {
	if res, has := lru.cache[key]; has {
		lru.moveToFront(res)
		return res.Value
	}

	return -1
}

func (lru *LRUCache) Put(key int, value int) {
	//判断是否存在，存在则moveToronto
	if vnode, has := lru.cache[key]; has {
		if vnode.Value != value {
			vnode.Value = value //修改
		}
		lru.moveToFront(vnode)
		return
	}

	//如果不存在则put,addToronto
	node := &Node{Key: key, Value: value}
	lru.cache[key] = node
	lru.addToFront(node)
	lru.size++

	//超出容量了需要删除
	if lru.size > lru.cap {
		delNode := lru.removeTail()
		delete(lru.cache, delNode.Key)
		lru.size--
	}

}
