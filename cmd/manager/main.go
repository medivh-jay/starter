package main

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"reflect"
	"starter/internal/entities"
	"starter/internal/manager"
	"starter/pkg/managers"
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
		root, _ := primitive.ObjectIDFromHex("5d4bc41a80a1cd400ae715f7")
		other, _ := primitive.ObjectIDFromHex("5d63b385790326f1bbd01317")

		log.Println(middlewares.NewToken(entities.Staff{}.FindByTopic(root)))
		log.Println(middlewares.NewToken(entities.Staff{}.FindByTopic(other)))
	}

	var typ = reflect.New(reflect.TypeOf(entities.Staff{}))
	log.Println(typ.Type().Implements(reflect.TypeOf((*managers.UpdateOrCreate)(nil)).Elem()))

	server.Run(manager.GetEngine())
}
