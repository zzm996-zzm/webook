package ioc

import "github.com/bwmarrin/snowflake"

func InitSnowFlake() *snowflake.Node {
	// 雪花算法
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}

	return node
}
