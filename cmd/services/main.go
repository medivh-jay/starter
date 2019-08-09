package main

import (
	_ "starter/cmd/services/docs"
	"starter/internal/services"
	"starter/pkg/mgo"
	"starter/pkg/server"
)

// swag 解析命令参考
//  swag init --generalInfo=router.go  --dir=../../internal/services --parseDependency=true
func main() {
	server.Mode = "services"

	mgo.Start()
	//var staff entities.Mgo
	//staff.Username = "mgo"
	//staff.Password = "123456"
	//log.Println(mgo.Collection("mgo").InsertOne(&staff))
	//log.Println(staff)
	//_ = mgo.Collection("mgo").Where(bson.M{"username": "mgo"}).FindOne(&staff)
	//log.Println(staff)
	//staff.Username = "mgo_2"
	//log.Println(mgo.Collection("mgo").Where(bson.M{"username": "mgo"}).UpdateOne(&staff))

	//log.Println(mgo.Collection("mgo").Where(bson.M{"username": "mgo_2"}).Delete())

	server.Run(services.GetEngine())
}
