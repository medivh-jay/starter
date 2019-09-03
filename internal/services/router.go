package services

import (
	"github.com/gin-gonic/gin"
	"starter/internal/services/controllers/order"
	"starter/pkg/app"
	"starter/pkg/middlewares"
	"starter/pkg/permission"
)

// @title starter
// @version 1.0
// @host golang-project.com
func GetEngine(engine *gin.Engine) {
	engine.Use(middlewares.CORS)
	engine.GET("/order", order.List)
	permission.Start(engine)

	engine.Any("/permission/test", func(context *gin.Context) {
		app.Logger().Debug(permission.HasPermission("10001", context))
	})
}
