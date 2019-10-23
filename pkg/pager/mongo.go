package pager

import (
	"go.mongodb.org/mongo-driver/bson"
	db "starter/pkg/database/mongo"
)

// Mongo 查询
type Mongo struct {
	where Where
	limit int
	skip  int
	index string
	sorts map[string]interface{}
}

// NewMongoDriver MongoDB官方驱动支持
func NewMongoDriver() *Mongo {
	return &Mongo{}
}

// Where 写入查询条件
func (mongo *Mongo) Where(kv Where) {
	if mongo.where == nil {
		mongo.where = make(Where)
	}
	mongo.where = mergeWhere(mongo.where, kv)
}

// Section 写入区间查询条件
func (mongo *Mongo) Section(section Section) {
	if mongo.where == nil {
		mongo.where = make(Where)
	}
	for k, v := range section {
		mongo.where[string(k)] = make(bson.M)
		if val, ok := v[Gte]; ok {
			mongo.where[string(k)].(bson.M)["$gte"] = val
		}
		if val, ok := v[Lte]; ok {
			mongo.where[string(k)].(bson.M)["$lte"] = val
		}
	}
}

// Limit 写入获取条数
func (mongo *Mongo) Limit(limit int) {
	mongo.limit = limit
}

// Skip 写入跳过条数
func (mongo *Mongo) Skip(skip int) {
	mongo.skip = skip
}

// Index 写入集合名字
func (mongo *Mongo) Index(index string) {
	mongo.index = index
}

// Sort 写入排序字段
func (mongo *Mongo) Sort(kv map[string]Sort) {
	mongo.sorts = make(bson.M)
	for k, v := range kv {
		if v == Asc {
			mongo.sorts[k] = 1
		} else {
			mongo.sorts[k] = -1
		}
	}
}

// Find 查询数据
func (mongo *Mongo) Find(data interface{}) {
	db.Database().SetTable(mongo.index).Where(bson.M(mongo.where)).Limit(int64(mongo.limit)).Skip(int64(mongo.skip)).Sort(mongo.sorts).FindMany(data)
}

// Count 查询条数
func (mongo *Mongo) Count() int {
	return int(db.Database().SetTable(mongo.index).Where(bson.M(mongo.where)).Count())
}
