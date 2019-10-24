package managers

import (
	"github.com/gin-gonic/gin"
	"reflect"
	"starter/pkg/app"
	"starter/pkg/database"
	"starter/pkg/database/orm"
	"starter/pkg/pager"
	"starter/pkg/validator"
)

// GormManager mysql 的实现
type GormManager struct {
	TableTyp reflect.Type
	Route    string
	Table    database.Table
}

// GetRoute 获取路由
func (manager *GormManager) GetRoute() string { return manager.Route }

// SetRoute 设置路由
func (manager *GormManager) SetRoute(route string) { manager.Route = route }

// SetTableTyp 保存表结构反射类型
func (manager *GormManager) SetTableTyp(typ reflect.Type) { manager.TableTyp = typ }

// GetTableTyp 获取表结构图反射类型对象
func (manager *GormManager) GetTableTyp() reflect.Type { return manager.TableTyp }

// GetTable 获取表结构体
func (manager *GormManager) GetTable() database.Table { return manager.Table }

// SetTable 保存表结构体信息
func (manager *GormManager) SetTable(table database.Table) { manager.Table = table }

// List 获取列表
func (manager *GormManager) List(ctx *gin.Context) {
	result := pager.New(ctx, pager.NewGormDriver()).SetIndex(manager.Table.TableName()).
		SetNextStartField("ID").SetPrevStartField("ID").Find(manager.Table).Result()
	app.NewResponse(app.Success, result, "SUCCESS").End(ctx)
}

// Post 增加数据
func (manager *GormManager) Post(ctx *gin.Context) {
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if !validate.IsValid() {
		app.NewResponse(app.Fail, validate.ErrorsInfo, "FAIL").End(ctx)
		return
	}

	if newInstance.Type().Implements(updateOrCreate) {
		newInstance.Interface().(database.UpdateOrCreate).PreOperation()
	}

	err := orm.Master().Create(newInstance.Interface()).Error
	if err != nil {
		app.NewResponse(app.Fail, nil, "FAIL").End(ctx)
		return
	}

	app.NewResponse(app.Success, newInstance.Interface(), "SUCCESS").End(ctx)

}

// Put 修改数据
func (manager *GormManager) Put(ctx *gin.Context) {
	id := ctx.PostForm("id")
	if id == "" {
		app.NewResponse(app.Fail, nil, "OperateIdCanNotBeNull").End(ctx)
		return
	}

	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if !validate.IsValid() {
		app.NewResponse(app.Fail, validate.ErrorsInfo, "FAIL").End(ctx)
		return
	}

	if newInstance.Type().Implements(updateOrCreate) {
		newInstance.Interface().(database.UpdateOrCreate).PreOperation()
	}
	err := orm.Master().Table(manager.GetTable().TableName()).
		Model(newInstance.Interface()).Where("id = ?", id).Update(newInstance.Interface()).Error
	if err != nil {
		app.NewResponse(app.Fail, nil, "FAIL").End(ctx)
		return
	}

	app.NewResponse(app.Success, newInstance.Interface(), "SUCCESS").End(ctx)
	return
}

// Delete 删除数据
func (manager *GormManager) Delete(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		app.NewResponse(app.Fail, nil, "OperateIdCanNotBeNull").End(ctx)
		return
	}
	var newInstance = reflect.New(manager.TableTyp)
	err := orm.Master().Table(manager.GetTable().TableName()).Where("id = ?", id).Delete(newInstance.Interface()).Error
	if err != nil {
		app.NewResponse(app.Fail, nil, "FAIL").End(ctx)
		return
	}
	app.NewResponse(app.Success, nil, "SUCCESS").End(ctx)
}
