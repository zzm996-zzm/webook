package cache

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type User struct {
	Id   int
	Name string
	Age  int
}

func (user *User) AddAge() {
	user.Age++
}

func TestLRU(t *testing.T) {

	testCases := []struct {
		name     string
		cap      int
		testFunc func(t *testing.T, cap int) (User, error)
		wantUser User
		wantErr  error
	}{
		{
			name: "Put",
			testFunc: func(t *testing.T, cap int) (User, error) {
				// 创建User
				user := User{
					Id:   1,
					Name: "zzm",
					Age:  18,
				}
				// 测试LRU Get方法正常
				lru := NewLRUCache[string, User](cap)
				err := lru.Put("user1", user, -1)
				return User{}, err
			},
			cap:      10,
			wantUser: User{},
			wantErr:  nil,
		},
		{
			name: "Get",
			testFunc: func(t *testing.T, cap int) (User, error) {
				// 创建User
				user := User{
					Id:   1,
					Name: "zzm",
					Age:  18,
				}
				// 测试LRU Get方法正常
				lru := NewLRUCache[string, User](cap)
				err := lru.Put("user1", user, -1)
				// 这种情况不应该出现err
				assert.NoError(t, err)

				u, _ := lru.Get("user1")

				return u, nil
			},
			cap: 10,
			wantUser: User{
				Id:   1,
				Name: "zzm",
				Age:  18,
			},
			wantErr: nil,
		},
		{
			name: "测试过期时间",
			testFunc: func(t *testing.T, cap int) (User, error) {
				// 创建User
				user := User{
					Id:   1,
					Name: "zzm",
					Age:  18,
				}
				// 测试LRU Get方法正常
				lru := NewLRUCache[string, User](cap)
				err := lru.Put("user1", user, time.Millisecond*100)
				// 这种情况不应该出现err
				assert.NoError(t, err)

				// 睡眠一下
				time.Sleep(time.Millisecond * 110)

				u, ok := lru.Get("user1")

				if !ok {
					return User{}, errors.New("数据不存在")
				}
				return u, nil
			},
			cap:      10,
			wantUser: User{},
			wantErr:  errors.New("数据不存在"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建容量为10的LRU缓存
			user, err := tc.testFunc(t, tc.cap)
			assert.Equal(t, tc.wantUser, user)
			assert.Equal(t, tc.wantErr, err)
		})
	}

	//cache := NewLRUCache[User, User](2)
	//var u User
	//u.id = 1
	//// 设置10S过期时间
	//cache.Put(u, u, time.Second*10)
	////cache.Put(2, 2)
	//v, _ := cache.Get(u) // 返回  u
	//fmt.Println(v)
	//time.Sleep(time.Second * 10)cache := NewLRUCache[User, User](2)
	//	//var u User
	//	//u.id = 1
	//	//// 设置10S过期时间
	//	//cache.Put(u, u, time.Second*10)
	//	////cache.Put(2, 2)
	//	//v, _ := cache.Get(u) // 返回  u
	//	//fmt.Println(v)
	//	//time.Sleep(time.Second * 10)
	//	//v, _ = cache.Get(u) // 返回  u
	//	//fmt.Println(v)
	//v, _ = cache.Get(u) // 返回  u
	//fmt.Println(v)
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
