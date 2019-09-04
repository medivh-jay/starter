package redis

import (
	"github.com/go-redis/redis"
	"starter/pkg/app"
	"time"
)

type config struct {
	Addr         string `toml:"addr"`
	Password     string `toml:"password"`
	Db           int    `toml:"db"`
	PoolSize     int    `toml:"pool_size"`
	MinIdleConns int    `toml:"min_idle_conns"`
}

var (
	Client *redis.Client
	conf   config
)

func Start() {
	_ = app.Config().Bind("application", "redis", &conf)
	Client = redis.NewClient(&redis.Options{
		Addr:         conf.Addr,
		Password:     conf.Password,
		DB:           conf.Db,
		PoolSize:     conf.PoolSize,
		MinIdleConns: conf.MinIdleConns,
	})
}

// 获取指定key的值,如果值不存在,就执行f方法将返回值存入redis
func Get(key string, expiration time.Duration, f func() string) string {
	cmd := Client.Get(key)
	var val string
	result, _ := cmd.Result()
	if len(result) == 0 {
		Client.Set(key, f(), expiration)
		return val
	}
	return result
}
