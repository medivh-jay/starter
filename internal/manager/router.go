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
	engine.Use(middlewares.CORS)
	engine.Use(middlewares.VerifyAuth)
	sessions.Start(engine)
	engine.Use(middlewares.CsrfToken)

	managers.Register(entities.Staff{}, managers.Mongo)
	managers.Register(entities.Mgo{}, managers.Mgo)
	managers.RegisterCustomManager(&controllers.CustomOrder{}, entities.Order{})

	managers.Start(engine)

	return engine
}
