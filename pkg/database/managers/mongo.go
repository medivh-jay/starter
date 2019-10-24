package managers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
	"starter/pkg/app"
	"starter/pkg/database"
	"starter/pkg/database/mongo"
	"starter/pkg/pager"
	"starter/pkg/validator"
)

var mongoObjectID = reflect.TypeOf(primitive.ObjectID{})

// MongoManager mongo 实现
type MongoManager struct {
	TableTyp  reflect.Type
	Route     string
	Table     database.Table
	ConvertID func(id string) interface{}
}

// GetRoute 获取路由
func (manager *MongoManager) GetRoute() string { return manager.Route }

// SetRoute 设置路由
func (manager *MongoManager) SetRoute(route string) { manager.Route = route }

// GetTable 获取表结构体
func (manager *MongoManager) GetTable() database.Table { return manager.Table }

// SetTable 保存表结构体信息
func (manager *MongoManager) SetTable(table database.Table) { manager.Table = table }

// GetTableTyp 获取表结构图反射类型对象
func (manager *MongoManager) GetTableTyp() reflect.Type { return manager.TableTyp }

// SetTableTyp 保存表结构反射类型
func (manager *MongoManager) SetTableTyp(typ reflect.Type) {
	manager.TableTyp = typ

	fieldTyp, exists := manager.TableTyp.FieldByName("ID")
	defaultConvert := func(id string) interface{} {
		return id
	}
	if !exists {
		manager.ConvertID = defaultConvert
	}

	if fieldTyp.Type == mongoObjectID {
		manager.ConvertID = func(id string) interface{} {
			objectID, _ := primitive.ObjectIDFromHex(id)
			return objectID
		}
	} else {
		manager.ConvertID = defaultConvert
	}
}

// List 数据列表
func (manager *MongoManager) List(ctx *gin.Context) {
	result := pager.New(ctx, pager.NewMongoDriver()).SetIndex(manager.Table.TableName()).
		SetNextStartField("ID").SetPrevStartField("ID").Find(manager.Table).Result()
	app.NewResponse(app.Success, result, "SUCCESS").End(ctx)
}

// Post 增加数据
func (manager *MongoManager) Post(ctx *gin.Context) {
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if !validate.IsValid() {
		app.NewResponse(app.Fail, validate.ErrorsInfo, "FAIL").End(ctx)
		return
	}

	if newInstance.Type().Implements(updateOrCreate) {
		newInstance.Interface().(database.UpdateOrCreate).PreOperation()
	}

	insertID := mongo.Collection(manager.GetTable()).InsertOne(newInstance.Interface())
	if insertID.InsertedID == nil {
		app.NewResponse(app.Fail, nil, "FAIL").End(ctx)
		return
	}

	_ = mongo.Collection(manager.GetTable()).Where(bson.M{"_id": insertID.InsertedID}).FindOne(newInstance.Interface())
	app.NewResponse(app.Success, newInstance.Interface(), "SUCCESS").End(ctx)
	return

}

// Put 修改数据
func (manager *MongoManager) Put(ctx *gin.Context) {
	id := ctx.PostForm("_id")
	if id == "" {
		app.NewResponse(app.Fail, nil, "OperateIdCanNotBeNull").End(ctx)
		return
	}
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if !validate.IsValid() {
		app.NewResponse(app.Fail, nil, "FAIL").End(ctx)
		return
	}

	newInstance.Elem().FieldByName("ID").Set(reflect.ValueOf(manager.ConvertID(id)))
	if newInstance.Type().Implements(updateOrCreate) {
		newInstance.Interface().(database.UpdateOrCreate).PreOperation()
	}
	result := mongo.Collection(manager.GetTable()).Where(bson.M{"_id": manager.ConvertID(id)}).UpdateOne(newInstance.Interface())
	if result.ModifiedCount == 0 {
		app.NewResponse(app.Fail, nil, "FAIL").End(ctx)
		return
	}
	app.NewResponse(app.Success, newInstance.Interface(), "SUCCESS").End(ctx)
	return

}

// Delete 删除数据
func (manager *MongoManager) Delete(ctx *gin.Context) {
	id := ctx.Query("_id")
	if id == "" {
		app.NewResponse(app.Fail, nil, "OperateIdCanNotBeNull").End(ctx)
		return
	}

	count := mongo.Collection(manager.GetTable()).Where(bson.M{"_id": manager.ConvertID(id)}).Delete()
	if count == 0 {
		app.NewResponse(app.Fail, nil, "FAIL").End(ctx)
		return
	}
	app.NewResponse(app.Success, nil, "SUCCESS").End(ctx)
}
