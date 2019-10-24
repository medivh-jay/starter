package managers

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"starter/pkg/app"
	"starter/pkg/database"
	"starter/pkg/database/mgo"
	"starter/pkg/pager"
	"starter/pkg/validator"
)

var mgoObjectID = reflect.TypeOf(bson.ObjectId(""))

// MgoManager mgo 的实现
type MgoManager struct {
	TableTyp  reflect.Type
	Route     string
	Table     database.Table
	ConvertID func(id string) interface{}
}

// GetRoute 获取路由
func (manager *MgoManager) GetRoute() string { return manager.Route }

// SetRoute 设置路由
func (manager *MgoManager) SetRoute(route string) { manager.Route = route }

// GetTableTyp 获取表结构图反射类型对象
func (manager *MgoManager) GetTableTyp() reflect.Type { return manager.TableTyp }

// GetTable 获取表结构体
func (manager *MgoManager) GetTable() database.Table { return manager.Table }

// SetTable 保存表结构体信息
func (manager *MgoManager) SetTable(table database.Table) { manager.Table = table }

// SetTableTyp 保存表结构反射类型
func (manager *MgoManager) SetTableTyp(typ reflect.Type) {
	manager.TableTyp = typ

	fieldTyp, exists := manager.TableTyp.FieldByName("ID")
	defaultConvert := func(id string) interface{} {
		return id
	}
	if !exists {
		manager.ConvertID = defaultConvert
	}

	if fieldTyp.Type == mgoObjectID {
		manager.ConvertID = func(id string) interface{} {
			objectID := bson.ObjectIdHex(id)
			return objectID
		}
	} else {
		manager.ConvertID = defaultConvert
	}
}

// List 数据列表
func (manager *MgoManager) List(ctx *gin.Context) {
	result := pager.New(ctx, pager.NewMgoDriver()).SetIndex(manager.Table.TableName()).
		SetNextStartField("ID").SetPrevStartField("ID").Find(manager.Table).Result()
	app.NewResponse(app.Success, result, "SUCCESS").End(ctx)
}

// Post 增加数据
func (manager *MgoManager) Post(ctx *gin.Context) {
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	collection := mgo.Collection(manager.GetTable())
	defer collection.Close()

	if !validate.IsValid() {
		app.NewResponse(app.Fail, nil, "FAIL").End(ctx)
		return
	}

	if newInstance.Type().Implements(updateOrCreate) {
		newInstance.Interface().(database.UpdateOrCreate).PreOperation()
	}
	insert, err := collection.InsertOne(newInstance.Interface())
	if err != nil {
		app.NewResponse(app.Fail, nil, "FAIL").End(ctx)
		return
	}

	app.NewResponse(app.Success, insert, "SUCCESS").End(ctx)
	return
}

// Put 修改数据
func (manager *MgoManager) Put(ctx *gin.Context) {
	id := ctx.PostForm("_id")
	if id == "" {
		app.NewResponse(app.Fail, nil, "OperateIdCanNotBeNull").End(ctx)
		return
	}

	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	collection := mgo.Collection(manager.GetTable())
	defer collection.Close()

	if !validate.IsValid() {
		app.NewResponse(app.Fail, nil, "FAIL").End(ctx)
		return
	}

	newInstance.Elem().FieldByName("ID").Set(reflect.ValueOf(manager.ConvertID(id)))
	if newInstance.Type().Implements(updateOrCreate) {
		newInstance.Interface().(database.UpdateOrCreate).PreOperation()
	}
	result := collection.Where(bson.M{"_id": manager.ConvertID(id)}).UpdateOne(newInstance.Interface())
	if !result {
		app.NewResponse(app.Fail, nil, "FAIL").End(ctx)
		return
	}

	app.NewResponse(app.Success, newInstance.Interface(), "SUCCESS").End(ctx)
	return

}

// Delete 删除数据
func (manager *MgoManager) Delete(ctx *gin.Context) {
	id := ctx.Query("_id")
	if id == "" {
		app.NewResponse(app.Fail, nil, "OperateIdCanNotBeNull").End(ctx)
		return
	}

	collection := mgo.Collection(manager.GetTable())
	defer collection.Close()
	result := collection.Where(bson.M{"_id": manager.ConvertID(id)}).Delete()
	if !result {
		app.NewResponse(app.Fail, nil, "FAIL").End(ctx)
		return
	}
	app.NewResponse(app.Success, nil, "SUCCESS").End(ctx)
}
