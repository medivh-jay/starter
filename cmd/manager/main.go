package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"starter/internal/entities"
	"starter/internal/manager"
	"starter/pkg/mgo"
	"starter/pkg/middlewares"
	"starter/pkg/server"
	"time"
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
	mgo.Start()
	middlewares.AuthInfo = middlewares.Auth{TableName: "staffs", Entity: entities.Staff{}, ParseId: func(s string) interface{} {
		id, _ := primitive.ObjectIDFromHex(s)
		return id
	}}

	id, _ := primitive.ObjectIDFromHex("5d4bc41a80a1cd400ae715f7")
	log.Println(middlewares.NewToken(id, 1, time.Now().Add(36000*time.Second).Unix()))
	server.Run(manager.GetEngine())
}
