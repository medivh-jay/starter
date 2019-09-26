// Package managers 基本的后台 curd 操作
package managers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	mgoBson "gopkg.in/mgo.v2/bson"
	"net/http"
	"reflect"
	"starter/pkg/app"
	"starter/pkg/database"
	starterMgo "starter/pkg/database/mgo"
	starterMongo "starter/pkg/database/mongo"
	starterOrm "starter/pkg/database/orm"
	"starter/pkg/validator"
)

const (
	// Mysql 类型
	Mysql EntityTyp = iota
	// Mongo 类型
	Mongo
	// Mgo 类型
	Mgo
)

type (
	// EntityTyp 结构体类型
	EntityTyp int
	// Managers 所有 mangers 对应数据结构体 对象
	Managers struct {
		container []ManagerInterface
	}
	// Response 统一返回结构体
	Response struct {
		Data     interface{} `json:"data"`      // 数据集
		AfterID  interface{} `json:"after_id"`  // 下一页,这个id为这一页最后一条id
		BeforeID interface{} `json:"before_id"` // 上一页,这个id为这一页第一条id
		Rows     int         `json:"rows"`      // 每页条数
		Count    int         `json:"count"`     // 总数
		Message  string      `json:"message"`
		Code     int         `json:"code"`
	}
	// ManagerInterface manger 接口
	ManagerInterface interface {
		List(*gin.Context)
		Post(*gin.Context)
		Put(*gin.Context)
		Delete(*gin.Context)
		GetRoute() string
		SetRoute(route string)
		SetTableTyp(typ reflect.Type)
		GetTableTyp() reflect.Type
		GetTable() database.Table
		SetTable(table database.Table)
	}
	// MysqlManager mysql 的实现
	MysqlManager struct {
		TableTyp reflect.Type
		Route    string
		Table    database.Table
	}
	// MongoManager mongo 实现
	MongoManager struct {
		TableTyp reflect.Type
		Route    string
		Table    database.Table
	}
	// MgoManager mgo 的实现
	MgoManager struct {
		TableTyp reflect.Type
		Route    string
		Table    database.Table
	}
	// Setup 设置
	Setup interface {
		set(managerInterface ManagerInterface)
	}
	// Route 路由设置
	Route struct {
		route string
	}
)

// New 获得一个新的 Managers 管理对象
func New() *Managers {
	return &Managers{
		container: make([]ManagerInterface, 0),
	}
}

var updateOrCreate = reflect.TypeOf((*database.UpdateOrCreate)(nil)).Elem()

//var managers = make(Managers, 0)

// NewManager 返回一个新的默认管理器
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

// SetRoute 设置自定义路由
func SetRoute(r string) *Route                        { return &Route{r} }
func (r Route) set(managerInterface ManagerInterface) { managerInterface.SetRoute(r.route) }

func newItems(manager ManagerInterface) reflect.Value {
	newInstance := reflect.MakeSlice(reflect.SliceOf(manager.GetTableTyp()), 0, 0)
	items := reflect.New(newInstance.Type())
	items.Elem().Set(newInstance)
	return items
}

// GetRoute 获取路由
func (manager *MysqlManager) GetRoute() string { return manager.Route }

// GetRoute 获取路由
func (manager *MongoManager) GetRoute() string { return manager.Route }

// GetRoute 获取路由
func (manager *MgoManager) GetRoute() string { return manager.Route }

// SetRoute 设置路由
func (manager *MysqlManager) SetRoute(route string) { manager.Route = route }

// SetRoute 设置路由
func (manager *MongoManager) SetRoute(route string) { manager.Route = route }

// SetRoute 设置路由
func (manager *MgoManager) SetRoute(route string) { manager.Route = route }

// SetTableTyp 保存表结构反射类型
func (manager *MysqlManager) SetTableTyp(typ reflect.Type) { manager.TableTyp = typ }

// SetTableTyp 保存表结构反射类型
func (manager *MongoManager) SetTableTyp(typ reflect.Type) { manager.TableTyp = typ }

// SetTableTyp 保存表结构反射类型
func (manager *MgoManager) SetTableTyp(typ reflect.Type) { manager.TableTyp = typ }

// GetTableTyp 获取表结构图反射类型对象
func (manager *MysqlManager) GetTableTyp() reflect.Type { return manager.TableTyp }

// GetTableTyp 获取表结构图反射类型对象
func (manager *MongoManager) GetTableTyp() reflect.Type { return manager.TableTyp }

// GetTableTyp 获取表结构图反射类型对象
func (manager *MgoManager) GetTableTyp() reflect.Type { return manager.TableTyp }

// GetTable 获取表结构体
func (manager *MysqlManager) GetTable() database.Table { return manager.Table }

// GetTable 获取表结构体
func (manager *MongoManager) GetTable() database.Table { return manager.Table }

// GetTable 获取表结构体
func (manager *MgoManager) GetTable() database.Table { return manager.Table }

// SetTable 保存表结构体信息
func (manager *MysqlManager) SetTable(table database.Table) { manager.Table = table }

// SetTable 保存表结构体信息
func (manager *MongoManager) SetTable(table database.Table) { manager.Table = table }

// SetTable 保存表结构体信息
func (manager *MgoManager) SetTable(table database.Table) { manager.Table = table }

// NewResponse 实例化 Response
func NewResponse(data interface{}, code int) *Response {
	var response = &Response{}
	response.Data = data
	response.AfterID = ""
	response.BeforeID = ""
	response.Rows = 0
	response.Count = 0
	response.Message = ""
	response.Code = code
	return response
}

// SetPageID 设置返回结构体中的上一页下一页的起始id
func (response *Response) SetPageID(items reflect.Value) *Response {
	if items.Elem().Len() > 0 {
		response.SetAfterID(items.Elem().Index(items.Elem().Len() - 1).FieldByName("ID").Interface())
		response.SetBeforeID(items.Elem().Index(0).FieldByName("ID").Interface())
	}
	return response
}

// SetAfterID 设置返回结构体中的上一页下一页的起始id
func (response *Response) SetAfterID(nextID interface{}) *Response {
	response.AfterID = nextID
	return response
}

// SetBeforeID 设置返回结构体中的上一页下一页的起始id
func (response *Response) SetBeforeID(nextID interface{}) *Response {
	response.BeforeID = nextID
	return response
}

// SetRows 设置每页显示条数
func (response *Response) SetRows(rows int) *Response { response.Rows = rows; return response }

// SetCount 设置总条数
func (response *Response) SetCount(count int) *Response { response.Count = count; return response }

// SetMessage 设置返回消息
func (response *Response) SetMessage(ctx *gin.Context, message string) *Response {
	response.Message = app.Translate(app.Lang(ctx), message)
	return response
}

// Start 启动指定的
func (managers *Managers) Start(router gin.IRoutes) {
	for _, manager := range managers.container {
		route := manager.GetRoute()
		manage := manager
		router.GET(route+"/list", manage.List)
		router.POST(route, manage.Post)
		router.PUT(route, manage.Put)
		router.DELETE(route, manage.Delete)
	}
}

// Register 注册一个管理器
func (managers *Managers) Register(entity database.Table, entityTyp EntityTyp, setups ...Setup) *Managers {
	manager := entityTyp.NewManager()
	managers.RegisterCustomManager(manager, entity, setups...)
	return managers
}

// RegisterCustomManager 自定义管理器
// 可自己继承 MysqlManager 或者 MongoManager 然后重写方法实现自定义操作
func (managers *Managers) RegisterCustomManager(manager ManagerInterface, entity database.Table, setups ...Setup) *Managers {
	manager.SetRoute("/" + entity.TableName())
	manager.SetTableTyp(reflect.TypeOf(entity))
	manager.SetTable(entity)
	for _, set := range setups {
		set.set(manager)
	}
	managers.container = append(managers.container, manager)
	return managers
}

// List 获取列表
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
	starterOrm.Master().Table(manager.GetTable().TableName()).
		Where(statement, params...).Limit(query.Limit(ctx)).Offset(query.Offset(ctx)).Order(sorts).Find(items.Interface())

	// 返回数据
	var response = NewResponse(items.Interface(), app.Success).SetPageID(items).SetRows(query.Limit(ctx))
	starterOrm.Master().Table(manager.GetTable().TableName()).Where(statement, params...).Count(&response.Count)
	response.SetMessage(ctx, "SUCCESS")
	ctx.JSON(http.StatusOK, response)
}

// Post 增加数据
func (manager *MysqlManager) Post(ctx *gin.Context) {
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if validate.IsValid() {
		if newInstance.Type().Implements(updateOrCreate) {
			newInstance.Interface().(database.UpdateOrCreate).PreOperation()
		}
		err := starterOrm.Master().Create(newInstance.Interface()).Error
		if err != nil {
			ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL"))
			return
		}

		ctx.JSON(http.StatusOK, NewResponse(newInstance.Interface(), app.Success).SetMessage(ctx, "SUCCESS"))
		return
	}

	ctx.JSON(http.StatusOK, NewResponse(validate.ErrorsInfo, app.Fail).SetMessage(ctx, "FAIL"))
}

// Put 修改数据
func (manager *MysqlManager) Put(ctx *gin.Context) {
	id := ctx.PostForm("id")
	if id == "" {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "OperateIdCanNotBeNull"))
		return
	}
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if validate.IsValid() {
		if newInstance.Type().Implements(updateOrCreate) {
			newInstance.Interface().(database.UpdateOrCreate).PreOperation()
		}
		err := starterOrm.Master().Table(manager.GetTable().TableName()).
			Model(newInstance.Interface()).Where("id = ?", id).Update(newInstance.Interface()).Error
		if err != nil {
			ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL"))
			return
		}

		ctx.JSON(http.StatusOK, NewResponse(newInstance.Interface(), app.Success).SetMessage(ctx, "SUCCESS"))
		return
	}

	ctx.JSON(http.StatusOK, NewResponse(validate.ErrorsInfo, app.Fail).SetMessage(ctx, "FAIL"))
}

// Delete 删除数据
func (manager *MysqlManager) Delete(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "OperateIdCanNotBeNull"))
		return
	}
	var newInstance = reflect.New(manager.TableTyp)
	err := starterOrm.Master().Table(manager.GetTable().TableName()).Where("id = ?", id).Delete(newInstance.Interface()).Error
	if err != nil {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL").SetCount(0))
		return
	}
	ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "SUCCESS").SetCount(app.Success))
}

// List 数据列表
func (manager *MongoManager) List(ctx *gin.Context) {
	var query = MongoQuery{entityTyp: manager.TableTyp}
	statement := query.GetQuery(ctx)
	items := newItems(manager)

	parse := ParseSectionParams(ctx, Mongo)
	statement = mergeMongo(statement, parse.Parse().(bson.M))

	sorts := NewSorter(Mongo).Parse(ctx).(bson.M)
	starterMongo.Collection(manager.GetTable()).Where(statement).Limit(int64(query.Limit(ctx))).Skip(int64(query.Offset(ctx))).Sort(sorts).FindMany(items.Interface())

	var response = NewResponse(items.Interface(), app.Success).SetRows(query.Limit(ctx)).
		SetCount(int(starterMongo.Collection(manager.GetTable()).Where(statement).Count())).SetPageID(items).SetMessage(ctx, "SUCCESS")

	ctx.JSON(http.StatusOK, response)
}

// Post 增加数据
func (manager *MongoManager) Post(ctx *gin.Context) {
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	if validate.IsValid() {
		if newInstance.Type().Implements(updateOrCreate) {
			newInstance.Interface().(database.UpdateOrCreate).PreOperation()
		}
		insertID := starterMongo.Collection(manager.GetTable()).InsertOne(newInstance.Interface())
		if insertID.InsertedID == nil {
			ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL"))
			return
		}
		_ = starterMongo.Collection(manager.GetTable()).Where(bson.M{"_id": insertID.InsertedID}).FindOne(newInstance.Interface())
		ctx.JSON(http.StatusOK, NewResponse(newInstance.Interface(), app.Success).SetMessage(ctx, "SUCCESS"))
		return
	}
	ctx.JSON(http.StatusOK, NewResponse(validate.ErrorsInfo, app.Fail).SetMessage(ctx, "FAIL"))
}

// Put 修改数据
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
		newInstance.Elem().FieldByName("ID").Set(reflect.ValueOf(query.convertID(id)))
		if newInstance.Type().Implements(updateOrCreate) {
			newInstance.Interface().(database.UpdateOrCreate).PreOperation()
		}
		result := starterMongo.Collection(manager.GetTable()).Where(bson.M{"_id": query.convertID(id)}).UpdateOne(newInstance.Interface())
		if result.ModifiedCount == 0 {
			ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL"))
			return
		}
		ctx.JSON(http.StatusOK, NewResponse(newInstance.Interface(), app.Success).SetMessage(ctx, "SUCCESS"))
		return
	}
	ctx.JSON(http.StatusOK, NewResponse(validate.ErrorsInfo, app.Fail).SetMessage(ctx, "FAIL"))
}

// Delete 删除数据
func (manager *MongoManager) Delete(ctx *gin.Context) {
	id := ctx.Query("_id")
	if id == "" {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "OperateIdCanNotBeNull"))
		return
	}
	var query = &MongoQuery{entityTyp: manager.TableTyp}
	count := starterMongo.Collection(manager.GetTable()).Where(bson.M{"_id": query.convertID(id)}).Delete()
	if count == 0 {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL").SetCount(int(count)))
		return
	}
	ctx.JSON(http.StatusOK, NewResponse(nil, app.Success).SetMessage(ctx, "SUCCESS").SetCount(int(count)))
}

// List 数据列表
func (manager *MgoManager) List(ctx *gin.Context) {
	var query = MgoQuery{entityTyp: manager.TableTyp}
	statement := query.GetQuery(ctx)
	items := newItems(manager)
	collection := starterMgo.Collection(manager.GetTable())
	defer collection.Close()

	parse := ParseSectionParams(ctx, Mgo)
	statement = mergeMgo(statement, parse.Parse().(mgoBson.M))
	var sorts = NewSorter(Mgo).Parse(ctx).([]string)

	collection.Where(statement).Limit(query.Limit(ctx)).Skip(query.Offset(ctx)).Sort(sorts...).FindMany(items.Interface())
	var response = NewResponse(items.Interface(), app.Success).SetRows(query.Limit(ctx)).
		SetCount(int(collection.Where(statement).Count())).SetPageID(items)

	ctx.JSON(http.StatusOK, response.SetMessage(ctx, "SUCCESS"))
}

// Post 增加数据
func (manager *MgoManager) Post(ctx *gin.Context) {
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	collection := starterMgo.Collection(manager.GetTable())
	defer collection.Close()
	if validate.IsValid() {
		if newInstance.Type().Implements(updateOrCreate) {
			newInstance.Interface().(database.UpdateOrCreate).PreOperation()
		}
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

// Put 修改数据
func (manager *MgoManager) Put(ctx *gin.Context) {
	id := ctx.PostForm("_id")
	if id == "" {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "OperateIdCanNotBeNull"))
		return
	}
	var newInstance = reflect.New(manager.TableTyp)
	validate := validator.Bind(ctx, newInstance.Interface())
	collection := starterMgo.Collection(manager.GetTable())
	defer collection.Close()

	if validate.IsValid() {
		var query = &MgoQuery{entityTyp: manager.TableTyp}
		newInstance.Elem().FieldByName("ID").Set(reflect.ValueOf(query.convertID(id)))
		if newInstance.Type().Implements(updateOrCreate) {
			newInstance.Interface().(database.UpdateOrCreate).PreOperation()
		}
		result := collection.Where(mgoBson.M{"_id": query.convertID(id)}).UpdateOne(newInstance.Interface())
		if !result {
			ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL"))
			return
		}
		ctx.JSON(http.StatusOK, NewResponse(newInstance.Interface(), app.Success).SetMessage(ctx, "SUCCESS"))
		return
	}
	ctx.JSON(http.StatusOK, NewResponse(validate.ErrorsInfo, app.Fail).SetMessage(ctx, "FAIL"))
}

// Delete 删除数据
func (manager *MgoManager) Delete(ctx *gin.Context) {
	id := ctx.Query("_id")
	if id == "" {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "OperateIdCanNotBeNull"))
		return
	}
	collection := starterMgo.Collection(manager.GetTable())
	defer collection.Close()
	var query = &MgoQuery{entityTyp: manager.TableTyp}
	result := collection.Where(mgoBson.M{"_id": query.convertID(id)}).Delete()
	if !result {
		ctx.JSON(http.StatusOK, NewResponse(nil, app.Fail).SetMessage(ctx, "FAIL").SetCount(1))
		return
	}
	ctx.JSON(http.StatusOK, NewResponse(nil, app.Success).SetMessage(ctx, "SUCCESS").SetCount(1))
}
