package order

import (
	"github.com/gin-gonic/gin"
	"starter/internal/entities"
	"starter/pkg/app"
	"starter/pkg/database/category"
	"starter/pkg/database/mongo"
	"starter/pkg/database/orm"
	"starter/pkg/pager"
	"sync"
)

// Category 分类控制器, 示例
type Category struct {
	entity entities.Category
}

var once sync.Once

var tempData = []entities.Category{
	{ID: "1", Pid: "", Alias: "aaa", Name: "顶级分类一", Pic: "无", Badge: "无", Description: "无", Level: 1},
	{ID: "2", Pid: "", Alias: "bbb", Name: "顶级分类二", Pic: "无", Badge: "无", Description: "无", Level: 1},
	{ID: "3", Pid: "", Alias: "ccc", Name: "顶级分类三", Pic: "无", Badge: "无", Description: "无", Level: 1},
	{ID: "4", Pid: "1", Alias: "aaa-1", Name: "顶级分类一(子分类1)", Pic: "无", Badge: "无", Description: "无", Level: 2},
	{ID: "5", Pid: "1", Alias: "aaa-2", Name: "顶级分类一(子分类2)", Pic: "无", Badge: "无", Description: "无", Level: 2},
	{ID: "6", Pid: "1", Alias: "aaa-3", Name: "顶级分类一(子分类3)", Pic: "无", Badge: "无", Description: "无", Level: 2},
	{ID: "7", Pid: "2", Alias: "bbb-1", Name: "顶级分类二(子分类1)", Pic: "无", Badge: "无", Description: "无", Level: 2},
	{ID: "8", Pid: "2", Alias: "bbb-2", Name: "顶级分类二(子分类2)", Pic: "无", Badge: "无", Description: "无", Level: 2},
	{ID: "9", Pid: "2", Alias: "bbb-3", Name: "顶级分类二(子分类3)", Pic: "无", Badge: "无", Description: "无", Level: 2},
	{ID: "10", Pid: "3", Alias: "ccc-1", Name: "顶级分类三(子分类1)", Pic: "无", Badge: "无", Description: "无", Level: 2},
	{ID: "11", Pid: "3", Alias: "ccc-2", Name: "顶级分类三(子分类2)", Pic: "无", Badge: "无", Description: "无", Level: 2},
	{ID: "12", Pid: "3", Alias: "ccc-3", Name: "顶级分类三(子分类3)", Pic: "无", Badge: "无", Description: "无", Level: 2},
	{ID: "13", Pid: "5", Alias: "aaa-2-1", Name: "顶级分类一(子分类2)-(子分类1)", Pic: "无", Badge: "无", Description: "无", Level: 3},
}

// NewCategory 实例化控制器
func NewCategory() *Category {
	orm.Master().AutoMigrate(entities.Category{})
	once.Do(func() {
		for _, data := range tempData {
			app.Logger().Debug(orm.Master().Create(&data).Error)
		}
		mongo.Collection(entities.Category{}).InsertMany(&tempData)
	})
	return &Category{entity: entities.Category{}}
}

// Mgo 使用 mgo
func (c *Category) Mgo(ctx *gin.Context) {
	app.NewResponse(app.Success, category.New().Table(c.entity).WithMgo().Categories()).End(ctx)
}

// Mongo 使用 mongo
func (c *Category) Mongo(ctx *gin.Context) {
	app.NewResponse(app.Success, category.New().Table(c.entity).WithMongo().Categories()).End(ctx)
}

// Mysql 使用 gorm
func (c *Category) Mysql(ctx *gin.Context) {
	app.NewResponse(app.Success, category.New().Table(c.entity).WithMysql().Categories()).End(ctx)
}

// ListMongo 分页功能
func (c *Category) ListMongo(ctx *gin.Context) {
	app.NewResponse(app.Success, pager.New(ctx, pager.NewMongoDriver()).SetIndex(c.entity.TableName()).Find(c.entity).Result()).End(ctx)
}

// ListMgo 分页功能
func (c *Category) ListMgo(ctx *gin.Context) {
	app.NewResponse(app.Success, pager.New(ctx, pager.NewMgoDriver()).SetIndex(c.entity.TableName()).Find(c.entity).Result()).End(ctx)
}

// ListMysql 分页功能
func (c *Category) ListMysql(ctx *gin.Context) {
	app.NewResponse(app.Success, pager.New(ctx, pager.NewGormDriver()).Where(pager.Where{"level": 2}).SetIndex(c.entity.TableName()).Find(c.entity).Result()).End(ctx)
}
