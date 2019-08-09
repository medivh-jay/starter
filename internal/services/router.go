package services

import (
	"github.com/gin-gonic/gin"
	"starter/internal/services/controllers/order"
	"starter/pkg/middlewares"
)

var engine = gin.Default()

// @title starter
// @version 1.0
// @host golang-project.com
func GetEngine() *gin.Engine {
	engine.Use(middlewares.CORS)

	engine.GET("/order", order.List)
	return engine
}
