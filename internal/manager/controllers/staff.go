package controllers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"starter/internal/entities"
	"starter/pkg/app"
	"starter/pkg/database/mongo"
	"starter/pkg/middlewares"
	"starter/pkg/password"
	"starter/pkg/validator"
)

type LoginForm struct {
	Username string `binding:"required,max=12" form:"username"`
	Password string `binding:"required,max=128" form:"password"`
}

// 登录示例
func Login(ctx *gin.Context) {
	var loginForm LoginForm
	if err := validator.Bind(ctx, &loginForm); !err.IsValid() {
		app.NewResponse(app.Success, err.ErrorsInfo).End(ctx)
		return
	}

	var staff entities.Staff
	notFound := mongo.Collection(staff).Where(bson.M{"username": loginForm.Username}).FindOne(&staff)
	if notFound != nil {
		app.NewResponse(app.Success, nil, notFound.Error()).End(ctx)
		return
	}

	if !password.Verify(loginForm.Password, staff.Password) {
		app.NewResponse(app.Success, nil, "password error").End(ctx)
		return
	}
	staff.Logged(ctx.DefaultQuery("platform", "web"))
	token, err := middlewares.NewToken(staff)
	if err != nil {
		app.NewResponse(app.Success, nil, err.Error()).End(ctx)
		return
	}

	app.NewResponse(app.Success, gin.H{"token": token}).End(ctx)
}

func StaffInfo(ctx *gin.Context) {
	staff, exists := ctx.Get(middlewares.AuthKey)
	if exists {
		app.NewResponse(app.Success, staff).End(ctx)
		return
	}
	app.NewResponse(app.Success, nil).End(ctx)
}
