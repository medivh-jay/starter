package redis

import (
	"github.com/go-redis/redis"
	"starter/pkg/config"
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
