package cache

import (
	"context"
	"fmt"
	"testing"
	"time"
	"webook/pkg/cache"
)

func TestCodeCache(t *testing.T) {
	lru := cache.NewLRUCache[string, any](10)
	local := NewLocalCodeCache(lru, time.Minute*2)
	//发生验证码
	err := local.Set(context.Background(), "login", "17674123135", "123456")
	if err != nil {
		return
	}

	// 第一次验证码失败
	verify, err := local.Verify(context.Background(), "login", "17674123135", "123451")
	if err != nil {
		return
	}
	// 获得验证次数
	key := local.key("login", "17674123135")
	v, _ := local.cache.Get(key)
	itm := v.(*codeItem)
	fmt.Println("验证剩余次数为：", itm.cnt)
	//// 第二次验证码失败
	//verify, err = local.Verify(context.Background(), "login", "17674123135", "123452")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//v, _ = local.cache.Get(key)
	//itm = v.(*codeItem)
	//fmt.Println("验证剩余次数为：", itm.cnt)
	//
	//// 第三次验证码失败
	//verify, err = local.Verify(context.Background(), "login", "17674123135", "123453")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//v, _ = local.cache.Get(key)
	//itm = v.(*codeItem)
	//fmt.Println("验证剩余次数为：", itm.cnt)

	// 第二次验证码成功
	verify, err = local.Verify(context.Background(), "login", "17674123135", "123456")
	if err != nil {
		return
	}

	v, _ = local.cache.Get(key)
	itm = v.(*codeItem)
	fmt.Println("验证剩余次数为：", itm.cnt)

	fmt.Println(verify)
}
