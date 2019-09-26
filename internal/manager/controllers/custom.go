package controllers

import (
	"github.com/gin-gonic/gin"
	"starter/pkg/app"
	"starter/pkg/database/managers"
)

// CustomOrder 继承 mangers 的 MysqlManager 并实现自定义的 List 方法
type CustomOrder struct {
	managers.MysqlManager
}

// List 自定义 List 方法
func (custom *CustomOrder) List(ctx *gin.Context) {
	app.Logger().Println("called this method")
	custom.MysqlManager.List(ctx)
}
