package services

import (
	"github.com/gin-gonic/gin"
	"starter/internal/services/controllers/order"
	"starter/pkg/middlewares"
)

// GetEngine 路由注册主方法
// @title starter
// @version 1.0
// @host golang-project.com
func GetEngine(engine *gin.Engine) {
	engine.Use(middlewares.CORS)
	engine.GET("/order", order.List)

	var category = order.NewCategory()
	categories := engine.Group("/categories")
	categories.GET("/mgo", category.Mgo)
	categories.GET("/mongo", category.Mongo)
	categories.GET("/mysql", category.Mysql)
	categories.GET("/list_mgo", category.ListMgo)
	categories.GET("/list_mysql", category.ListMysql)
	categories.GET("/list_mongo", category.ListMongo)
}
