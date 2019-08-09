package controllers

import (
	"github.com/gin-gonic/gin"
	"log"
	"starter/pkg/managers"
)

type CustomOrder struct {
	managers.MysqlManager
}

func (custom *CustomOrder) List(ctx *gin.Context) {
	log.Println("called this method")
	custom.MysqlManager.List(ctx)
}
