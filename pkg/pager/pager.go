// Package pager 分页工具
//  目前可使用具体实现
//  var driver pager.Driver
//  driver = NewMongoDriver()
//  driver = NewGormDriver()
//  driver = NewMgoDriver()
//
//  pager.New(ctx, driver).SetIndex(c.entity.TableName()).Find(c.entity).Result()
package pager

import (
	"github.com/gin-gonic/gin"
	"reflect"
	"strconv"
	"strings"
)

// Where 查询条件
type Where map[string]interface{}

// Sort 排序
type Sort int

const (
	// Desc 降序
	Desc Sort = -1
	// Asc 升序
	Asc = 1
)

// Driver 查询驱动, Pagination会在真正查询的时候传入参数
type Driver interface {
	// 查询条件
	Where(kv Where)
	// 范围查询条件
	Section(section Section)
	// 每页数量
	Limit(limit int)
	// 跳过行
	Skip(skip int)
	// Index 索引名称,可以是表名,或者比如 es 的 index 名称等,标识具体资源集合的东西
	Index(index string)
	// 排序
	Sort(kv map[string]Sort)
	// 查询具体操作
	Find(data interface{})
	// 如果具体实现代码里边需要分析Find传入的data结构, 这个方法会被调用并传入data反射类型
	SetTyp(typ reflect.Type)
	// 计算总数
	Count() int
}

// Result 返回结构体
type Result struct {
	// 数据列表
	Data interface{} `json:"data" xml:"data"`
	// 下一页id，其实就是这一页最后一个id
	NextID interface{} `json:"next_id" xml:"next_id"`
	// 上一页id, 就是这一页的第一个id
	PrevID interface{} `json:"prev_id" xml:"prev_id"`
	// 当前筛选条件下的数据总数量
	Count int `json:"count" xml:"count"`
	// 每页显示条数
	Rows int `json:"rows" xml:"rows"`
}

// Pagination 分页工具
type Pagination struct {
	ctx            *gin.Context
	defaultWhere   Where
	defaultLimit   int
	index          string
	driver         Driver
	dataTyp        reflect.Type
	nextStartField string
	prevStartField string
	result         *Result
}

// New 新的分页工具实例
func New(ctx *gin.Context, driver Driver) *Pagination {
	return &Pagination{ctx: ctx, driver: driver, defaultLimit: 12}
}

// Result 返回数据
func (pagination *Pagination) Result() *Result {
	return pagination.result
}

// SetNextStartField 设置获取下一页开始id的字段名
func (pagination *Pagination) SetNextStartField(field string) *Pagination {
	pagination.nextStartField = field
	return pagination
}

// SetPrevStartField 设置获取上一页开始id的字段名
func (pagination *Pagination) SetPrevStartField(field string) *Pagination {
	pagination.prevStartField = field
	return pagination
}

// SetIndex 设置集合或表名，有的驱动可能需要这个, 比如 elastic search
func (pagination *Pagination) SetIndex(index string) *Pagination {
	pagination.index = index
	return pagination
}

// Where 传入默认查询参数, 该传入参数将会在这个实例中一直被使用
func (pagination *Pagination) Where(kv Where) *Pagination {
	pagination.defaultWhere = kv
	return pagination
}

// Limit 默认每页条数, 如果页面未传 rows 就使用默认的
func (pagination *Pagination) Limit(limit int) *Pagination {
	pagination.defaultLimit = limit
	return pagination
}

// Find 查询数据
//  structure 不需要传类似 []struct 这样的类型, 直接传入 struct 就行, 非指针
func (pagination *Pagination) Find(structure interface{}) *Pagination {
	limit := ParsingLimit(pagination.ctx, pagination.defaultLimit)
	pagination.dataTyp = reflect.TypeOf(structure)
	pagination.driver.SetTyp(pagination.dataTyp)
	pagination.driver.Limit(limit)
	pagination.driver.Sort(ParseSorts(pagination.ctx))
	pagination.driver.Index(pagination.index)
	pagination.driver.Skip(limit * ParseSkip(pagination.ctx))
	pagination.driver.Section(ParseSection(pagination.ctx))
	pagination.driver.Where(mergeWhere(pagination.defaultWhere, ParsingQuery(pagination.ctx)))

	data := newSlice(pagination.dataTyp)
	pagination.driver.Find(data.Interface())
	pagination.result = &Result{
		Data:  data.Interface(),
		Count: pagination.driver.Count(),
		Rows:  limit,
	}
	if data.Elem().Len() > 0 && (pagination.nextStartField != "" && pagination.prevStartField != "") {
		pagination.result.NextID = data.Elem().Index(data.Elem().Len() - 1).FieldByName(pagination.nextStartField).Interface()
		pagination.result.PrevID = data.Elem().Index(0).FieldByName(pagination.prevStartField).Interface()
	}

	return pagination
}

// ParseSkip 跳过的行数
func ParseSkip(ctx *gin.Context) int {
	page := ctx.DefaultQuery("page", "1")
	p, _ := strconv.Atoi(page)
	return p - 1
}

// ParseSorts 解析排序字段, 传入规则为  sorts=-filed1,+field2,field3
//  "-"号标识降序
//  "+"或者无符号标识升序
func ParseSorts(ctx *gin.Context) map[string]Sort {
	var sortMap = make(map[string]Sort)
	query, exists := ctx.GetQuery("sorts")
	if !exists {
		return sortMap
	}
	sorts := strings.Split(query, ",")
	for _, sort := range sorts {
		if strings.HasPrefix(sort, "-") {
			sortMap[strings.TrimPrefix(sort, "-")] = Desc
		} else {
			sortMap[strings.TrimPrefix(sort, "+")] = Asc
		}
	}
	return sortMap
}

// ParsingLimit 解析每页显示的条数
func ParsingLimit(ctx *gin.Context, defaultLimit int) int {
	val := ctx.DefaultQuery("rows", strconv.Itoa(defaultLimit))
	limit, _ := strconv.Atoi(val)
	return limit
}

// ParsingQuery 解析请求中的query参数
func ParsingQuery(ctx *gin.Context) Where {
	where := make(Where)
	query := ctx.Request.URL.Query()
	for key, val := range query {
		if len(val) == 1 {
			if val[0] != "" {
				where[key] = val[0]
			}
		}
		if len(val) > 1 {
			where[key] = val
		}
	}
	return where
}

func mergeWhere(defaultWhere, where Where) Where {
	for k, v := range defaultWhere {
		if k == "rows" || k == "sorts" || k == "page" || k == "section" {
			continue
		}
		where[k] = v
	}
	delete(where, "rows")
	delete(where, "sorts")
	delete(where, "page")
	delete(where, "section")
	return where
}

func newSlice(typ reflect.Type) reflect.Value {
	newInstance := reflect.MakeSlice(reflect.SliceOf(typ), 0, 0)
	items := reflect.New(newInstance.Type())
	items.Elem().Set(newInstance)
	return items
}
