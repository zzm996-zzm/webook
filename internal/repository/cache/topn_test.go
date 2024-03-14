package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	_ "github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"testing"
)

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: viper.GetString("redis.addr"),
	})
}

type Article struct {
	Id      int64
	Title   string
	LikeCnt int64
	Content string
}

func TestGetCache(t *testing.T) {
	client := InitRedis()
	// 先添加几个数据
	client.ZAdd(context.Background(), "alike", redis.Z{
		Score:  100,
		Member: "article:1",
	}, redis.Z{
		Score:  300,
		Member: "article:2",
	}, redis.Z{
		Score:  200,
		Member: "article:3",
	})

	client.HSet(context.Background(), "article:1", map[string]interface{}{
		"id":      1,
		"title":   "我的标题1",
		"content": "我的内容1",
		"likeCnt": 100,
	})

	client.HSet(context.Background(), "article:2", map[string]interface{}{
		"id":      2,
		"title":   "我的标题2",
		"content": "我的内容2",
		"likeCnt": 300,
	})

	client.HSet(context.Background(), "article:3", map[string]interface{}{
		"id":      3,
		"title":   "我的标题3",
		"content": "我的内容3",
		"likeCnt": 200,
	})

	topN := NewInteractiveTopN[Article](client, nil, 3)
	cache, err := topN.GetCache(context.Background(), "alike")
	if err != nil {
		return
	}

	t.Log(cache)
}
