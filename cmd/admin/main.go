package main

import (
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"html/template"
	"net/http"
	"starter/pkg/server"
	"strings"
)

func main() {
	server.Mode = "admin"

	// pages.New(&Table).List(ctx)
	server.Run(func(engine *gin.Engine) {
		engine.LoadHTMLFiles()

		// 注册自定义函数
		engine.SetFuncMap(template.FuncMap{
			"map": func(json string) gin.H {
				var out gin.H
				_ = jsoniter.UnmarshalFromString(json, &out)
				return out
			},
		})

		// 加载模板
		engine.LoadHTMLGlob("web/admin/tmpl/*/*")

		// 统一访问
		engine.GET("/*filepath", func(context *gin.Context) {
			if strings.HasPrefix(context.Request.URL.Path, "/static") || context.Request.URL.Path == "/favicon.ico" {
				context.File("web/admin/" + context.Request.URL.Path)
				return
			}

			context.HTML(http.StatusOK, context.Request.URL.Path, nil)
		})
	})
}
