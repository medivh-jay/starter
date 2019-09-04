package middlewares

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"starter/pkg/app"
	"starter/pkg/sessions"
	"sync"
	"time"
)

type config struct {
	Domain string `toml:"domain"`
}

var (
	conf config
	once sync.Once
)

var key = "csrf_token"

func generateToken() string {
	length := 12
	rand.Seed(time.Now().UnixNano())
	var token = make([]byte, length, length)
	for i := 0; i < length; i++ {
		token[i] = byte(rand.Intn(127))
	}

	return base64.URLEncoding.EncodeToString(token)
}

func CsrfToken(ctx *gin.Context) {
	once.Do(func() {
		_ = app.Config().Bind("application", "sessions", &conf)
	})
	switch ctx.Request.Method {
	case http.MethodGet:
		sessions.Del(ctx, key)
		token := generateToken()
		sessions.Set(ctx, key, token)
		ctx.SetCookie(key, token, 3600, "/", conf.Domain, false, false)
	default:
		token, err := ctx.Cookie(key)
		if err != nil || token == "" {
			app.NewResponse(app.Fail, nil, "CsrfTokenError").End(ctx)
			ctx.Abort()
			return
		}
		if token != sessions.Get(ctx, key) {
			app.NewResponse(app.Fail, nil, "CsrfTokenError").End(ctx)
			ctx.Abort()
			return
		}
		sessions.Del(ctx, key)
		ctx.Next()
	}
}
