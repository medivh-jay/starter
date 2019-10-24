// Package managers 基本的后台 curd 操作
package managers

import (
	"github.com/gin-gonic/gin"
	"reflect"
	"starter/pkg/database"
)

const (
	// Mysql 类型
	Mysql EntityTyp = iota
	// Mongo 类型
	Mongo
	// Mgo 类型
	Mgo
)

var updateOrCreate = reflect.TypeOf((*database.UpdateOrCreate)(nil)).Elem()

type (
	// EntityTyp 结构体类型
	EntityTyp int
	// Managers 所有 mangers 对应数据结构体 对象
	Managers struct {
		container []ManagerInterface
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

// NewManager 返回一个新的默认管理器
func (entityTyp EntityTyp) NewManager() ManagerInterface {
	switch entityTyp {
	case Mysql:
		return new(GormManager)
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
