package order

import (
	"github.com/gin-gonic/gin"
	"starter/internal/entities"
	"starter/pkg/app"
	"starter/pkg/database/orm"
	"starter/pkg/validator"
)

type ListToVisitor struct {
	Id string `form:"id" binding:"required,max=32"`
}

// @Summary 订单
// @Tags 订单列表
// @Produce  json
// @Param    id        query    string     true      "订单id"
// @Success  0         {object}  entities.Order
// @failure  404
// @Router  /order [get]
func List(ctx *gin.Context) {
	var listToVisitor ListToVisitor
	if validate := validator.Bind(ctx, &listToVisitor); !validate.IsValid() {
		app.NewResponse(app.Fail, validate.ErrorsInfo).End(ctx)
		return
	}

	var order entities.Order
	result := orm.Slave().Where("id = ?", listToVisitor.Id).Find(&order)
	if result.RowsAffected > 0 {
		app.NewResponse(app.Success, order).End(ctx)
		return
	} else {
		app.NewResponse(app.NotFound, nil).End(ctx)
		return
	}

}
