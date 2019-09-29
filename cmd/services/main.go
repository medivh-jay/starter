package main

import (
	_ "starter/cmd/services/docs"
	"starter/internal/services"
	"starter/pkg/server"
)

// swag 解析命令参考
//  swag init --generalInfo=router.go  --dir=../../internal/services --parseDependency=true
func main() {
	server.Mode = "services"
	server.Run(services.GetEngine)
}
