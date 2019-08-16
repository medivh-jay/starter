package main

import (
	"github.com/gin-gonic/gin"
	"log"
	_ "starter/cmd/services/docs"
	"starter/internal/services"
	"starter/pkg/permission"
	"starter/pkg/server"
)

// swag 解析命令参考
//  swag init --generalInfo=router.go  --dir=../../internal/services --parseDependency=true
func main() {
	server.Mode = "services"
	server.After = func(engine *gin.Engine) {
		log.Println(permission.GetPermissionsForUser("10001"))
	}
	server.Run(services.GetEngine())
}
