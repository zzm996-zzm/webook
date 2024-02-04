//go:build k8s

package config

var Config = config{
	DB: DBconfig{
		DNS: "root:root@tcp(webook-mysql:3308)/webook",
	},
	Redis: RedisConfig{Addr: "webook-redis:6379"},
}
