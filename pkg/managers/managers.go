// 基本的后台 curd 操作
package managers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	mgoBson "gopkg.in/mgo.v2/bson"
	"net/http"
	"reflect"
	"starter/pkg/app"
	"starter/pkg/mgo"
	"starter/pkg/mongo"
	"starter/pkg/orm"
	"starter/pkg/validator"
)

const (
	Mysql EntityTyp = iota
	Mongo
	Mgo
)

type (
	EntityTyp int
	Managers  []ManagerInterface
	Response  struct {
		Data     interface{} `json:"data"`      // 数据集
		AfterId  interface{} `json:"after_id"`  // 下一页,这个id为这一页最后一条id
		BeforeId interface{} `json:"before_id"` // 上一页,这个id为这一页第一条id
		Rows     int         `json:"rows"`      // 每页条数
		Count    int         `json:"count"`     // 总数
		Message  string      `json:"message"`
		Code     int         `json:"code"`
	}
	ManagerInterface interface {
		List(*gin.Context)
		Post(*gin.Context)
		Put(*gin.Context)
		Delete(*gin.Context)
		GetRoute() string
		SetRoute(route string)
		SetTableName(table string)
		SetTableTyp(typ reflect.Type)
	}
	MysqlManager struct {
		TableName string
		TableTyp  reflect.Type
		Route     string
	}
	MongoManager struct {
		TableName string
		TableTyp  reflect.Type
		Route     string
	}
	MgoManager struct {
		TableName string
		TableTyp  reflect.Type
		Route     string
	}
)

var managers = make(Managers, 0)

func (manager *MysqlManager) GetRoute() string             { return manager.Route }
func (manager *MongoManager) GetRoute() string             { return manager.Route }
func (manager *MgoManager) GetRoute() string               { return manager.Route }
func (manager *MysqlManager) SetRoute(route string)        { manager.Route = route }
func (manager *MongoManager) SetRoute(route string)        { manager.Route = route }
func (manager *MgoManager) SetRoute(route string)          { manager.Route = route }
func (manager *MysqlManager) SetTableName(table string)    { manager.TableName = table }
func (manager *MongoManager) SetTableName(table string)    { manager.TableName = table }
func (manager *MgoManager) SetTableName(table string)      { manager.TableName = table }
func (manager *MysqlManager) SetTableTyp(typ reflect.Type) { manager.TableTyp = typ }
func (manager *MongoManager) SetTableTyp(typ reflect.Type) { manager.TableTyp = typ }
func (manager *MgoManager) SetTableTyp(typ reflect.Type)   { manager.TableTyp = typ }

func NewResponse(data interface{}, code int) *Response {
	var response = &Response{}
	response.Data = data
	response.AfterId = ""
	response.BeforeId = ""
	response.Rows = 0
	response.Count = 0
	response.Message = ""
	response.Code = code
	return response
}

func (response *Response) SetAfterId(nextId interface{}) *Response {
	response.AfterId = nextId
	return response
}
func (response *Response) SetBeforeId(nextId interface{}) *Response {
	response.BeforeId = nextId
	return response
}
func (response *Response) SetRows(rows int) *Response   { response.Rows = rows; return response }
func (response *Response) SetCount(count int) *Response { response.Count = count; return response }
func (response *Response) SetMessage(message string) *Response {
	response.Message = message
	return response
}

func Start(engine *gin.Engine) {
	for _, manager := range managers {
		route := manager.GetRoute()
		manage := manager
		engine.GET(route+"/list", func(context *gin.Context) { manage.List(context) })
		engine.POST(route, func(context *gin.Context) { manage.Post(context) })
		engine.PUT(route, func(context *gin.Context) { manage.Put(context) })
		engine.DELETE(route, func(context *gin.Context) { manage.Delete(context) })
	}
}

// 实例化一个新的默认管理器
func Register(route, table string, entity interface{}, entityTyp EntityTyp) Managers {
	switch entityTyp {
	case Mysql:
		newManager := &MysqlManager{TableName: table, TableTyp: reflect.TypeOf(entity), Route: route}
		managers = append(managers, newManager)
	case Mongo:
		newManager := &MongoManager{TableName: table, TableTyp: reflect.TypeOf(entity), Route: route}
		managers = append(managers, newManager)
	case Mgo:
		newManager := &MgoManager{TableName: table, TableTyp: reflect.TypeOf(entity), Route: route}
		managers = append(managers, newManager)
	}
	return managers
}

// 自定义管理器
// 可自己继承 MysqlManager 或者 MongoManager 然后重写方法实现自定义操作
func RegisterCustomManager(managerInterface ManagerInterface, route, table string, entity interface{}) Managers {
	managerInterface.SetRoute(route)
	managerInterface.SetTableName(table)
	managerInterface.SetTableTyp(reflect.TypeOf(entity))
	managers = append(managers, managerInterface)
	return managers
}

// 获取列表
func (manager *MysqlManager) List(ctx *gin.Context) {
	newInstance := reflect.MakeSlice(reflect.SliceOf(manager.TableTyp), 0, 0)
	items := reflect.New(newInstance.Type())
	items.Elem().Set(newInstance)
	var query = &MysqlQuery{entityTyp: manager.TableTyp}

	statement, params := query.GetQuery(ctx)
	parse := ParseSectionParams(ctx)
	parse.Engine = Mysql
	if statement != "" {
		statement = statement + " and " + parse.Parse().(string)
	} else {
		statement = parse.Parse().(string)
	}

	orm.Master().Table(manager.TableName).Where(statement, params...).Limit(query.Limit(ctx)).Offset(query.Offset(ctx)).Find(items.Interface())

	var response = NewResponse(items.Interface(), app.Success)
	response.SetRows(query.Limit(ctx))
	orm.Master().Table(manager.TableName).Where(statement, params...).Count(&response.Count)

	if items.Elem().Len() > 0 {
		response.SetAfterId(items.Elem().Index(items.Elem().Len() - 1).FieldByName("Id").Interface())
		response.SetBeforeId(items.Elem().Index(0).FieldByName("Id").Interface())
	}

	message := app.Translate(ctx.DefaultQuery("lang", "zh-cn"), "SUCCESS")
	response.SetMessage(message)
	ctx.JSON(http.StatusOK, response)
}

// 增加数据
func (manager *MysqlManager) Post(ctx *gin.Context) {
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if validate.IsValid() {
		err := orm.Master().Create(newInstance.Interface()).Error
		if err != nil {
			var response = NewResponse(nil, app.Fail)
			message := app.Translate(ctx.DefaultQuery("lang", "zh-cn"), "FAIL")
			response.SetMessage(message).SetCount(app.Fail)
			ctx.JSON(http.StatusOK, response)
			return
		}
		var response = NewResponse(newInstance.Interface(), app.Success)
		response.SetMessage(app.Translate(ctx.DefaultQuery("lang", "zh-cn"), "SUCCESS"))
		ctx.JSON(http.StatusOK, response)
		return
	}
	var response = NewResponse(validate.ErrorsInfo, app.Fail)
	response.SetMessage(app.Translate(ctx.DefaultQuery("lang", "zh-cn"), "FAIL"))
	ctx.JSON(http.StatusOK, response)
}

// 修改数据
func (manager *MysqlManager) Put(ctx *gin.Context) {
	id := ctx.PostForm("id")
	lang := ctx.DefaultQuery("lang", "zh-cn")
	if id == "" {
		ctx.JSON(http.StatusOK,
			NewResponse(nil, app.Fail).SetMessage(app.Translate(lang, "OperateIdCanNotBeNull")))
		return
	}
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if validate.IsValid() {
		err := orm.Master().Table(manager.TableName).Where("id = ?", id).Updates(newInstance.Interface()).Error
		if err != nil {
			message := app.Translate(lang, "FAIL")
			ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(message).SetCount(app.Fail))
			return
		}
		var response = NewResponse(newInstance.Interface(), app.Success)
		response.SetMessage(app.Translate(lang, "SUCCESS"))
		ctx.JSON(http.StatusOK, response)
		return
	}
	var response = NewResponse(validate.ErrorsInfo, app.Fail)
	response.SetMessage(app.Translate(lang, "FAIL"))
	ctx.JSON(http.StatusOK, response)
}

// 删除数据
func (manager *MysqlManager) Delete(ctx *gin.Context) {
	id := ctx.Query("id")
	lang := ctx.DefaultQuery("lang", "zh-cn")
	if id == "" {
		ctx.JSON(http.StatusOK,
			NewResponse(nil, app.Fail).SetMessage(app.Translate(lang, "OperateIdCanNotBeNull")))
		return
	}
	var newInstance = reflect.New(manager.TableTyp)
	err := orm.Master().Table(manager.TableName).Where("id = ?", id).Delete(newInstance.Interface()).Error
	if err != nil {
		message := app.Translate(lang, "FAIL")
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(message).SetCount(0))
		return
	}
	message := app.Translate(lang, "SUCCESS")
	ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(message).SetCount(app.Success))
}

func (manager *MongoManager) List(ctx *gin.Context) {
	var query = MongoQuery{entityTyp: manager.TableTyp}
	statement := query.GetQuery(ctx)
	newInstance := reflect.MakeSlice(reflect.SliceOf(manager.TableTyp), 0, 0)
	items := reflect.New(newInstance.Type())
	items.Elem().Set(newInstance)

	parse := ParseSectionParams(ctx)
	parse.Engine = Mongo
	statement = mergeMongo(statement, parse.Parse().(bson.M))

	mongo.Collection(manager.TableName).Where(statement).Limit(int64(query.Limit(ctx))).Skip(int64(query.Offset(ctx))).FindMany(items.Interface())

	var response = NewResponse(items.Interface(), app.Success)
	response.SetRows(query.Limit(ctx))
	response.SetCount(int(mongo.Collection(manager.TableName).Where(statement).Count()))

	if items.Elem().Len() > 0 {
		response.SetAfterId(items.Elem().Index(items.Elem().Len() - 1).FieldByName("Id").Interface())
		response.SetBeforeId(items.Elem().Index(0).FieldByName("Id").Interface())
	}

	message := app.Translate(ctx.DefaultQuery("lang", "zh-cn"), "SUCCESS")
	response.SetMessage(message)
	ctx.JSON(http.StatusOK, response)
}

func (manager *MongoManager) Post(ctx *gin.Context) {
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if validate.IsValid() {
		insertId := mongo.Collection(manager.TableName).InsertOne(newInstance.Interface())
		if insertId.InsertedID == nil {
			var response = NewResponse(nil, app.Fail)
			message := app.Translate(ctx.DefaultQuery("lang", "zh-cn"), "FAIL")
			response.SetMessage(message).SetCount(app.Fail)
			ctx.JSON(http.StatusOK, response)
			return
		}
		_ = mongo.Collection(manager.TableName).Where(bson.M{"_id": insertId.InsertedID}).FindOne(newInstance.Interface())
		var response = NewResponse(newInstance.Interface(), app.Success)
		response.SetMessage(app.Translate(ctx.DefaultQuery("lang", "zh-cn"), "SUCCESS"))
		ctx.JSON(http.StatusOK, response)
		return
	}
	var response = NewResponse(validate.ErrorsInfo, app.Fail)
	response.SetMessage(app.Translate(ctx.DefaultQuery("lang", "zh-cn"), "FAIL"))
	ctx.JSON(http.StatusOK, response)
}

func (manager *MongoManager) Put(ctx *gin.Context) {
	id := ctx.PostForm("_id")
	lang := ctx.DefaultQuery("lang", "zh-cn")
	if id == "" {
		ctx.JSON(http.StatusOK,
			NewResponse(nil, app.Fail).SetMessage(app.Translate(lang, "OperateIdCanNotBeNull")))
		return
	}
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if validate.IsValid() {
		var query = &MongoQuery{entityTyp: manager.TableTyp}
		newInstance.Elem().FieldByName("Id").Set(reflect.ValueOf(query.convertId(id)))
		result := mongo.Collection(manager.TableName).Where(bson.M{"_id": query.convertId(id)}).UpdateOne(newInstance.Interface())
		if result.ModifiedCount == 0 {
			message := app.Translate(lang, "FAIL")
			ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(message).SetCount(app.Fail))
			return
		}
		var response = NewResponse(newInstance.Interface(), app.Success)
		response.SetMessage(app.Translate(lang, "SUCCESS"))
		ctx.JSON(http.StatusOK, response)
		return
	}
	var response = NewResponse(validate.ErrorsInfo, app.Fail)
	response.SetMessage(app.Translate(lang, "FAIL"))
	ctx.JSON(http.StatusOK, response)
}

func (manager *MongoManager) Delete(ctx *gin.Context) {
	id := ctx.Query("_id")
	lang := ctx.DefaultQuery("lang", "zh-cn")
	if id == "" {
		ctx.JSON(http.StatusOK,
			NewResponse(nil, app.Fail).SetMessage(app.Translate(lang, "OperateIdCanNotBeNull")))
		return
	}
	var query = &MongoQuery{entityTyp: manager.TableTyp}
	count := mongo.Collection(manager.TableName).Where(bson.M{"_id": query.convertId(id)}).Delete()
	if count == 0 {
		message := app.Translate(lang, "FAIL")
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(message).SetCount(int(count)))
		return
	}
	message := app.Translate(lang, "SUCCESS")
	ctx.JSON(http.StatusOK, NewResponse(nil, app.Success).SetMessage(message).SetCount(int(count)))
}

func (manager *MgoManager) List(ctx *gin.Context) {
	var query = MgoQuery{entityTyp: manager.TableTyp}
	statement := query.GetQuery(ctx)
	newInstance := reflect.MakeSlice(reflect.SliceOf(manager.TableTyp), 0, 0)
	items := reflect.New(newInstance.Type())
	items.Elem().Set(newInstance)
	collection := mgo.Collection(manager.TableName)
	defer collection.Close()

	parse := ParseSectionParams(ctx)
	parse.Engine = Mgo
	statement = mergeMgo(statement, parse.Parse().(mgoBson.M))

	collection.Where(statement).Limit(query.Limit(ctx)).Skip(query.Offset(ctx)).FindMany(items.Interface())
	var response = NewResponse(items.Interface(), app.Success)
	response.SetRows(query.Limit(ctx))
	response.SetCount(int(collection.Where(statement).Count()))

	if items.Elem().Len() > 0 {
		response.SetAfterId(items.Elem().Index(items.Elem().Len() - 1).FieldByName("Id").Interface())
		response.SetBeforeId(items.Elem().Index(0).FieldByName("Id").Interface())
	}

	message := app.Translate(ctx.DefaultQuery("lang", "zh-cn"), "SUCCESS")
	response.SetMessage(message)
	ctx.JSON(http.StatusOK, response)
}

func (manager *MgoManager) Post(ctx *gin.Context) {
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	collection := mgo.Collection(manager.TableName)
	defer collection.Close()
	if validate.IsValid() {
		insert, err := collection.InsertOne(newInstance.Interface())
		if err != nil {
			var response = NewResponse(nil, app.Fail)
			message := app.Translate(ctx.DefaultQuery("lang", "zh-cn"), "FAIL")
			response.SetMessage(message).SetCount(app.Fail)
			ctx.JSON(http.StatusOK, response)
			return
		}

		var response = NewResponse(insert, app.Success)
		response.SetMessage(app.Translate(ctx.DefaultQuery("lang", "zh-cn"), "SUCCESS"))
		ctx.JSON(http.StatusOK, response)
		return
	}
	var response = NewResponse(validate.ErrorsInfo, app.Fail)
	response.SetMessage(app.Translate(ctx.DefaultQuery("lang", "zh-cn"), "FAIL"))
	ctx.JSON(http.StatusOK, response)
}

func (manager *MgoManager) Put(ctx *gin.Context) {
	id := ctx.PostForm("_id")
	lang := ctx.DefaultQuery("lang", "zh-cn")
	if id == "" {
		ctx.JSON(http.StatusOK,
			NewResponse(nil, app.Fail).SetMessage(app.Translate(lang, "OperateIdCanNotBeNull")))
		return
	}
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	collection := mgo.Collection(manager.TableName)
	defer collection.Close()

	if validate.IsValid() {
		var query = &MgoQuery{entityTyp: manager.TableTyp}
		newInstance.Elem().FieldByName("Id").Set(reflect.ValueOf(query.convertId(id)))
		result := collection.Where(mgoBson.M{"_id": query.convertId(id)}).UpdateOne(newInstance.Interface())
		if !result {
			message := app.Translate(lang, "FAIL")
			ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(message).SetCount(app.Fail))
			return
		}
		var response = NewResponse(newInstance.Interface(), app.Success)
		response.SetMessage(app.Translate(lang, "SUCCESS"))
		ctx.JSON(http.StatusOK, response)
		return
	}
	var response = NewResponse(validate.ErrorsInfo, app.Fail)
	response.SetMessage(app.Translate(lang, "FAIL"))
	ctx.JSON(http.StatusOK, response)
}

func (manager *MgoManager) Delete(ctx *gin.Context) {
	id := ctx.Query("_id")
	lang := ctx.DefaultQuery("lang", "zh-cn")
	if id == "" {
		ctx.JSON(http.StatusOK,
			NewResponse(nil, app.Fail).SetMessage(app.Translate(lang, "OperateIdCanNotBeNull")))
		return
	}
	collection := mgo.Collection(manager.TableName)
	defer collection.Close()
	var query = &MgoQuery{entityTyp: manager.TableTyp}
	result := collection.Where(mgoBson.M{"_id": query.convertId(id)}).Delete()
	if !result {
		message := app.Translate(lang, "FAIL")
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(message).SetCount(1))
		return
	}
	message := app.Translate(lang, "SUCCESS")
	ctx.JSON(http.StatusOK, NewResponse(nil, app.Success).SetMessage(message).SetCount(1))
}
