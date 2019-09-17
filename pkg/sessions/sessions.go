package sessions

import (
	"github.com/gin-contrib/sessions"
	redisSession "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"starter/pkg/app"
	"starter/pkg/log"
	"strconv"
)

type config struct {
	Key          string `toml:"key"`
	Name         string `toml:"name"`
	Domain       string `toml:"domain"`
	Addr         string `toml:"addr"`
	Password     string `toml:"password"`
	Db           int    `toml:"db"`
	PoolSize     int    `toml:"pool_size"`
	MinIdleConns int    `toml:"min_idle_conns"`
}

var conf config

func Start(engine *gin.Engine) {
	_ = app.Config().Bind("application", "sessions", &conf)
	store, err := redisSession.NewStoreWithDB(conf.PoolSize, "tcp", conf.Addr, conf.Password, strconv.Itoa(conf.Db), []byte(conf.Key))
	if err != nil {
		log.Logger.Error(err)
	} else {
		store.Options(sessions.Options{MaxAge: 3600, Path: "/", Domain: conf.Domain, HttpOnly: true})
		engine.Use(sessions.Sessions(conf.Name, store))
	}
}

func Get(c *gin.Context, key string) string {
	sess := sessions.Default(c)
	val := sess.Get(key)
	if val != nil {
		return val.(string)
	}
	return ""
}

func Set(c *gin.Context, key, val string) {
	sess := sessions.Default(c)
	sess.Set(key, val)
	_ = sess.Save()
}

func Del(c *gin.Context, key string) {
	sess := sessions.Default(c)
	sess.Delete(key)
	_ = sess.Save()
}
