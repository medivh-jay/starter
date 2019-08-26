package main

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"starter/internal/entities"
	"starter/internal/manager"
	"starter/pkg/middlewares"
	"starter/pkg/server"
)

// @Summary 订单1
// @Tags 订单列表1
// @Produce  json
// @Param    id        query    string     true      "订单id"
// @Success  0         {object}  entities.Order
// @failure  404
// @Router  /order [get]
func main() {
	server.Mode = "manager"
	middlewares.AuthEntity = entities.Staff{}

	// 在mongo连接上之后再操作
	server.After = func(engine *gin.Engine) {
		id, _ := primitive.ObjectIDFromHex("5d4bc41a80a1cd400ae715f7")
		log.Println(middlewares.NewToken(entities.Staff{}.FindByTopic(id)))
	}

	server.Run(manager.GetEngine())
}
