package pager

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"starter/pkg/app"
	db "starter/pkg/database/mgo"
)

// Mgo 查询
type Mgo struct {
	where Where
	limit int
	skip  int
	index string
	sorts []string
}

// NewMgoDriver mgo 驱动支持
func NewMgoDriver() *Mgo {
	return &Mgo{}
}

// Where 写入查询条件
func (mgo *Mgo) Where(kv Where) {
	if mgo.where == nil {
		mgo.where = make(Where)
	}
	mgo.where = mergeWhere(mgo.where, kv)
}

// Section 写入区间查询条件
func (mgo *Mgo) Section(section Section) {
	if mgo.where == nil {
		mgo.where = make(Where)
	}
	for k, v := range section {
		mgo.where[string(k)] = make(bson.M)
		if val, ok := v[Gte]; ok {
			mgo.where[string(k)].(bson.M)["$gte"] = val
		}
		if val, ok := v[Lte]; ok {
			mgo.where[string(k)].(bson.M)["$lte"] = val
		}
	}
}

// Limit 写入获取条数
func (mgo *Mgo) Limit(limit int) {
	mgo.limit = limit
}

// Skip 写入跳过条数
func (mgo *Mgo) Skip(skip int) {
	mgo.skip = skip
}

// Index 写入集合名字
func (mgo *Mgo) Index(index string) {
	mgo.index = index
}

// Sort 写入排序字段
func (mgo *Mgo) Sort(kv map[string]Sort) {
	app.Logger().Debug(kv)
	mgo.sorts = make([]string, 0, 0)
	for k, v := range kv {
		if v == Desc {
			mgo.sorts = append(mgo.sorts, fmt.Sprintf("-%s", k))
		} else {
			mgo.sorts = append(mgo.sorts, k)
		}
	}
	app.Logger().Debug(mgo.sorts)
}

// Find 查询数据
func (mgo *Mgo) Find(data interface{}) {
	database := db.Database()
	defer database.Close()
	database.SetTable(mgo.index).Where(bson.M(mgo.where)).Limit(mgo.limit).Skip(mgo.skip).Sort(mgo.sorts...).FindMany(data)
}

// Count 查询条数
func (mgo *Mgo) Count() int {
	database := db.Database()
	defer database.Close()
	return int(database.SetTable(mgo.index).Where(bson.M(mgo.where)).Count())
}
