package controllers

import (
	"github.com/gin-gonic/gin"
	"starter/pkg/app"
	"starter/pkg/managers"
)

type CustomOrder struct {
	managers.MysqlManager
}

func (custom *CustomOrder) List(ctx *gin.Context) {
	app.Logger().Println("called this method")
	custom.MysqlManager.List(ctx)
}
