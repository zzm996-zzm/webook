package config

type config struct {
	DB    DBconfig
	Redis RedisConfig
}

type DBconfig struct {
	DNS string
}

type RedisConfig struct {
	Addr string
}
