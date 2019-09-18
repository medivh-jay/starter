package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"starter/internal/entities"
	"starter/internal/manager"
	"starter/pkg/app"
	"starter/pkg/middlewares"
	"starter/pkg/mongo"
	"starter/pkg/password"
	"starter/pkg/permission"
	"starter/pkg/queue"
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

	queue.NewConsumer().SetTopics("test").Do(func(consumer *kafka.Consumer, message *kafka.Message) {
		var d map[string]interface{}
		_ = jsoniter.Unmarshal(message.Value, &d)
		fmt.Println(d)
	})

	_ = queue.NewProducer().Send("test", map[string]int{"a": 1, "b": 2})

	// 在mongo连接上之后再操作
	server.After = func(engine *gin.Engine) {

		// 这里的操作可以自行完成, 写在这里只是为了方便开发
		// 生成root用户
		var staff = &entities.Staff{Username: "root", Password: password.Hash("123456")}
		if mongo.Collection(staff).Where(bson.M{"username": "root"}).Count() == 0 {
			staffResult := mongo.Collection(staff).InsertOne(staff)

			// 生成总权限
			var perm = &permission.Permission{Name: "所有权限", Path: "*", Method: "*"}
			if mongo.Collection(perm).Where(bson.M{"path": "*", "method": "*"}).Count() == 0 {
				permissionResult := mongo.Collection(perm).InsertOne(perm)

				// 生成超级管理员角色
				var role = &permission.Role{Name: "超级管理员", Permissions: []string{permissionResult.InsertedID.(primitive.ObjectID).Hex()}}
				if mongo.Collection(role).Where(bson.M{"permission": permissionResult.InsertedID.(primitive.ObjectID).Hex()}).Count() == 0 {
					roleResult := mongo.Collection(role).InsertOne(role)

					var binding = &permission.Binding{UserId: staffResult.InsertedID.(primitive.ObjectID).Hex(), RoleId: roleResult.InsertedID.(primitive.ObjectID).Hex()}
					app.Logger().Debug(mongo.Collection(binding).InsertOne(binding))
				}
			}
		}

		//email.StartEmailSender()
		//email.Send(email.NewSender("服务启动成功", "服务启动成功!", "xxx@qq.com"))
		//email.Send(email.NewHtmlSender("subject", email.ParseHtml("HTML模板路径", gin.H{"name":"value"}), ""))
		// page.NewMgo().
	}

	server.Run(manager.GetEngine)
}
