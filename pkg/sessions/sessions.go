package sessions

import (
	"github.com/gin-contrib/sessions"
	redisSession "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"starter/pkg/app"
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

// Inject 启动session服务, 在自定义的路由代码中调用, 传入 *gin.Engine 对象
func Inject(engine *gin.Engine) gin.IRoutes {
	_ = app.Config().Bind("application", "sessions", &conf)
	store, err := redisSession.NewStoreWithDB(conf.PoolSize, "tcp", conf.Addr, conf.Password, strconv.Itoa(conf.Db), []byte(conf.Key))
	if err != nil {
		app.Logger().WithField("log_type", "pkg.sessions.sessions").Error(err)
		return engine
	}

	store.Options(sessions.Options{MaxAge: 3600, Path: "/", Domain: conf.Domain, HttpOnly: true})
	return engine.Use(sessions.Sessions(conf.Name, store))
}

// Get 获取指定session
func Get(c *gin.Context, key string) string {
	sess := sessions.Default(c)
	val := sess.Get(key)
	if val != nil {
		return val.(string)
	}
	return ""
}

// Set 设置session
func Set(c *gin.Context, key, val string) {
	sess := sessions.Default(c)
	sess.Set(key, val)
	_ = sess.Save()
}

// Del 删除指定session
func Del(c *gin.Context, key string) {
	sess := sessions.Default(c)
	sess.Delete(key)
	_ = sess.Save()
}
