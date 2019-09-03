package redis

import (
	"github.com/go-redis/redis"
	"starter/pkg/config"
	"time"
)

var Client *redis.Client

func Start() {
	conf := config.Config.Redis
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
