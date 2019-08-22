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
		GetTableTyp() reflect.Type
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
	Setup interface {
		set(managerInterface ManagerInterface)
	}
	route struct {
		route string
	}
)

var managers = make(Managers, 0)

// 返回一个新的默认管理器
func (entityTyp EntityTyp) NewManager() ManagerInterface {
	switch entityTyp {
	case Mysql:
		return new(MysqlManager)
	case Mongo:
		return new(MongoManager)
	case Mgo:
		return new(MgoManager)
	}
	panic("entity type error")
}

func (r route) set(managerInterface ManagerInterface) { managerInterface.SetRoute(r.route) }
func SetRoute(r string) *route                        { return &route{r} }

func newItems(manager ManagerInterface) reflect.Value {
	newInstance := reflect.MakeSlice(reflect.SliceOf(manager.GetTableTyp()), 0, 0)
	items := reflect.New(newInstance.Type())
	items.Elem().Set(newInstance)
	return items
}

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
func (manager *MysqlManager) GetTableTyp() reflect.Type    { return manager.TableTyp }
func (manager *MongoManager) GetTableTyp() reflect.Type    { return manager.TableTyp }
func (manager *MgoManager) GetTableTyp() reflect.Type      { return manager.TableTyp }

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

func (response *Response) SetPageId(items reflect.Value) *Response {
	if items.Elem().Len() > 0 {
		response.SetAfterId(items.Elem().Index(items.Elem().Len() - 1).FieldByName("Id").Interface())
		response.SetBeforeId(items.Elem().Index(0).FieldByName("Id").Interface())
	}
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
func (response *Response) SetMessage(ctx *gin.Context, message string) *Response {
	response.Message = app.Translate(app.Lang(ctx), message)
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
func Register(entity app.Table, entityTyp EntityTyp, setups ...Setup) Managers {
	manager := entityTyp.NewManager()
	RegisterCustomManager(manager, entity, setups...)
	return managers
}

// 自定义管理器
// 可自己继承 MysqlManager 或者 MongoManager 然后重写方法实现自定义操作
func RegisterCustomManager(manager ManagerInterface, entity app.Table, setups ...Setup) Managers {
	manager.SetRoute("/" + entity.TableName())
	manager.SetTableName(entity.TableName())
	manager.SetTableTyp(reflect.TypeOf(entity))
	for _, set := range setups {
		set.set(manager)
	}
	managers = append(managers, manager)
	return managers
}

// 获取列表
func (manager *MysqlManager) List(ctx *gin.Context) {
	items := newItems(manager)

	// 查询条件
	var query = &MysqlQuery{entityTyp: manager.TableTyp}
	statement, params := query.GetQuery(ctx)

	// 区间查询内容
	parse := ParseSectionParams(ctx, Mysql)
	if statement != "" {
		statement = statement + " and " + parse.Parse().(string)
	} else {
		statement = parse.Parse().(string)
	}

	// 排序内容
	var sorts = NewSorter(Mysql).Parse(ctx).(string)

	// 查询
	orm.Master().Table(manager.TableName).Where(statement, params...).Limit(query.Limit(ctx)).Offset(query.Offset(ctx)).Order(sorts).Find(items.Interface())

	// 返回数据
	var response = NewResponse(items.Interface(), app.Success).SetPageId(items).SetRows(query.Limit(ctx))
	orm.Master().Table(manager.TableName).Where(statement, params...).Count(&response.Count)
	response.SetMessage(ctx, "SUCCESS")
	ctx.JSON(http.StatusOK, response)
}

// 增加数据
func (manager *MysqlManager) Post(ctx *gin.Context) {
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if validate.IsValid() {
		err := orm.Master().Create(newInstance.Interface()).Error
		if err != nil {
			ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL"))
			return
		}

		ctx.JSON(http.StatusOK, NewResponse(newInstance.Interface(), app.Success).SetMessage(ctx, "SUCCESS"))
		return
	}

	ctx.JSON(http.StatusOK, NewResponse(validate.ErrorsInfo, app.Fail).SetMessage(ctx, "FAIL"))
}

// 修改数据
func (manager *MysqlManager) Put(ctx *gin.Context) {
	id := ctx.PostForm("id")
	if id == "" {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "OperateIdCanNotBeNull"))
		return
	}
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if validate.IsValid() {
		err := orm.Master().Table(manager.TableName).Model(newInstance.Interface()).Where("id = ?", id).Update(newInstance.Interface()).Error
		if err != nil {
			ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL"))
			return
		}

		ctx.JSON(http.StatusOK, NewResponse(newInstance.Interface(), app.Success).SetMessage(ctx, "SUCCESS"))
		return
	}

	ctx.JSON(http.StatusOK, NewResponse(validate.ErrorsInfo, app.Fail).SetMessage(ctx, "FAIL"))
}

// 删除数据
func (manager *MysqlManager) Delete(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "OperateIdCanNotBeNull"))
		return
	}
	var newInstance = reflect.New(manager.TableTyp)
	err := orm.Master().Table(manager.TableName).Where("id = ?", id).Delete(newInstance.Interface()).Error
	if err != nil {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL").SetCount(0))
		return
	}
	ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "SUCCESS").SetCount(app.Success))
}

func (manager *MongoManager) List(ctx *gin.Context) {
	var query = MongoQuery{entityTyp: manager.TableTyp}
	statement := query.GetQuery(ctx)
	items := newItems(manager)

	parse := ParseSectionParams(ctx, Mongo)
	statement = mergeMongo(statement, parse.Parse().(bson.M))

	sorts := NewSorter(Mongo).Parse(ctx).(bson.M)
	mongo.Collection(manager.TableName).Where(statement).Limit(int64(query.Limit(ctx))).Skip(int64(query.Offset(ctx))).Sort(sorts).FindMany(items.Interface())

	var response = NewResponse(items.Interface(), app.Success).SetRows(query.Limit(ctx)).
		SetCount(int(mongo.Collection(manager.TableName).Where(statement).Count())).SetPageId(items).SetMessage(ctx, "SUCCESS")

	ctx.JSON(http.StatusOK, response)
}

func (manager *MongoManager) Post(ctx *gin.Context) {
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if validate.IsValid() {
		insertId := mongo.Collection(manager.TableName).InsertOne(newInstance.Interface())
		if insertId.InsertedID == nil {
			ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL"))
			return
		}
		_ = mongo.Collection(manager.TableName).Where(bson.M{"_id": insertId.InsertedID}).FindOne(newInstance.Interface())
		ctx.JSON(http.StatusOK, NewResponse(newInstance.Interface(), app.Success).SetMessage(ctx, "SUCCESS"))
		return
	}
	ctx.JSON(http.StatusOK, NewResponse(validate.ErrorsInfo, app.Fail).SetMessage(ctx, "FAIL"))
}

func (manager *MongoManager) Put(ctx *gin.Context) {
	id := ctx.PostForm("_id")
	if id == "" {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "OperateIdCanNotBeNull"))
		return
	}
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if validate.IsValid() {
		var query = &MongoQuery{entityTyp: manager.TableTyp}
		newInstance.Elem().FieldByName("Id").Set(reflect.ValueOf(query.convertId(id)))
		result := mongo.Collection(manager.TableName).Where(bson.M{"_id": query.convertId(id)}).UpdateOne(newInstance.Interface())
		if result.ModifiedCount == 0 {
			ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL"))
			return
		}
		ctx.JSON(http.StatusOK, NewResponse(newInstance.Interface(), app.Success).SetMessage(ctx, "SUCCESS"))
		return
	}
	ctx.JSON(http.StatusOK, NewResponse(validate.ErrorsInfo, app.Fail).SetMessage(ctx, "FAIL"))
}

func (manager *MongoManager) Delete(ctx *gin.Context) {
	id := ctx.Query("_id")
	if id == "" {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "OperateIdCanNotBeNull"))
		return
	}
	var query = &MongoQuery{entityTyp: manager.TableTyp}
	count := mongo.Collection(manager.TableName).Where(bson.M{"_id": query.convertId(id)}).Delete()
	if count == 0 {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL").SetCount(int(count)))
		return
	}
	ctx.JSON(http.StatusOK, NewResponse(nil, app.Success).SetMessage(ctx, "SUCCESS").SetCount(int(count)))
}

func (manager *MgoManager) List(ctx *gin.Context) {
	var query = MgoQuery{entityTyp: manager.TableTyp}
	statement := query.GetQuery(ctx)
	items := newItems(manager)
	collection := mgo.Collection(manager.TableName)
	defer collection.Close()

	parse := ParseSectionParams(ctx, Mgo)
	statement = mergeMgo(statement, parse.Parse().(mgoBson.M))
	var sorts = NewSorter(Mgo).Parse(ctx).([]string)

	collection.Where(statement).Limit(query.Limit(ctx)).Skip(query.Offset(ctx)).Sort(sorts...).FindMany(items.Interface())
	var response = NewResponse(items.Interface(), app.Success).SetRows(query.Limit(ctx)).
		SetCount(int(collection.Where(statement).Count())).SetPageId(items)

	ctx.JSON(http.StatusOK, response.SetMessage(ctx, "SUCCESS"))
}

func (manager *MgoManager) Post(ctx *gin.Context) {
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	collection := mgo.Collection(manager.TableName)
	defer collection.Close()
	if validate.IsValid() {
		insert, err := collection.InsertOne(newInstance.Interface())
		if err != nil {
			ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL"))
			return
		}

		ctx.JSON(http.StatusOK, NewResponse(insert, app.Success).SetMessage(ctx, "SUCCESS"))
		return
	}
	ctx.JSON(http.StatusOK, NewResponse(validate.ErrorsInfo, app.Fail).SetMessage(ctx, "FAIL"))
}

func (manager *MgoManager) Put(ctx *gin.Context) {
	id := ctx.PostForm("_id")
	if id == "" {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "OperateIdCanNotBeNull"))
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
			ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL"))
			return
		}
		ctx.JSON(http.StatusOK, NewResponse(newInstance.Interface(), app.Success).SetMessage(ctx, "SUCCESS"))
		return
	}
	ctx.JSON(http.StatusOK, NewResponse(validate.ErrorsInfo, app.Fail).SetMessage(ctx, "FAIL"))
}

func (manager *MgoManager) Delete(ctx *gin.Context) {
	id := ctx.Query("_id")
	if id == "" {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "OperateIdCanNotBeNull"))
		return
	}
	collection := mgo.Collection(manager.TableName)
	defer collection.Close()
	var query = &MgoQuery{entityTyp: manager.TableTyp}
	result := collection.Where(mgoBson.M{"_id": query.convertId(id)}).Delete()
	if !result {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL").SetCount(1))
		return
	}
	ctx.JSON(http.StatusOK, NewResponse(nil, app.Success).SetMessage(ctx, "SUCCESS").SetCount(1))
}
