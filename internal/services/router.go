package services

import (
	"github.com/gin-gonic/gin"
	"log"
	"starter/internal/services/controllers/order"
	"starter/pkg/managers"
	"starter/pkg/middlewares"
	"starter/pkg/permission"
)

var engine = gin.Default()

// @title starter
// @version 1.0
// @host golang-project.com
func GetEngine() *gin.Engine {
	engine.Use(middlewares.CORS)
	engine.GET("/order", order.List)
	permission.Start()
	managers.Start(engine)

	engine.Any("/permission/test", func(context *gin.Context) {
		log.Println(permission.HasPermission("10001", context))
	})

	return engine
}
