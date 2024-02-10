package cache

import (
	"fmt"
	"testing"
	"time"
)

type User struct {
	id int
}

func TestLRU(t *testing.T) {
	cache := NewLRUCache[User, User](2)
	var u User
	u.id = 1
	// 设置10S过期时间
	cache.Put(u, u, time.Second*10)
	//cache.Put(2, 2)
	v, _ := cache.Get(u) // 返回  u
	fmt.Println(v)
	time.Sleep(time.Second * 10)
	v, _ = cache.Get(u) // 返回  u
	fmt.Println(v)
	//cache.Put(3, 3)
	//v, ok := cache.Get(2) // 返回 -1 (未找到)
	//fmt.Println(ok)       //返回false
	//cache.Put(4, 4)       // 该操作会使得密钥 1 作废
	//v, ok = cache.Get(1)  // 返回 -1 (未找到)
	//fmt.Println(ok)       //返回false
	//v, ok = cache.Get(3)
	//fmt.Println(v) //返回3
	//v, ok = cache.Get(4)
	//fmt.Println(v) //返回4
}
