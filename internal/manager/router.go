package manager

import (
	"github.com/gin-gonic/gin"
	"starter/internal/entities"
	"starter/internal/manager/controllers"
	"starter/pkg/managers"
	"starter/pkg/middlewares"
	"starter/pkg/sessions"
)

var engine = gin.Default()

func GetEngine() *gin.Engine {
	engine.Use(middlewares.VerifyAuth)
	sessions.Start(engine)
	engine.Use(middlewares.CsrfToken)
	managers.NewManager("/staff", "staffs", entities.Staff{}, managers.Mongo)
	managers.NewManager("/mgo", "mgo", entities.Mgo{}, managers.Mgo)
	managers.NewCustomManager(&controllers.CustomOrder{}, "/order", "orders", entities.Order{})
	managers.Start(engine)

	return engine
}
