package pager

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
	db "starter/pkg/database/mongo"
)

var mongoObjectIDTyp = reflect.TypeOf(primitive.ObjectID{})

// Mongo 查询
type Mongo struct {
	where Where
	limit int
	skip  int
	index string
	sorts map[string]interface{}
	// 字段的类型转换操作
	FieldConvert map[string]func(str interface{}) interface{}
}

// NewMongoDriver MongoDB官方驱动支持
func NewMongoDriver() *Mongo {
	return &Mongo{FieldConvert: make(map[string]func(str interface{}) interface{})}
}

// Where 写入查询条件
func (mongo *Mongo) Where(kv Where) {
	if mongo.where == nil {
		mongo.where = make(Where)
	}
	mongo.where = mergeWhere(mongo.where, kv)
	for k, v := range mongo.where {
		if convert, ok := mongo.FieldConvert[k]; ok {
			mongo.where[k] = convert(v)
		}
	}
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

// SetTyp 记录结构体类型
func (mongo *Mongo) SetTyp(typ reflect.Type) {
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("bson")
		switch field.Type.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			mongo.FieldConvert[tag] = StringToInt
		case reflect.Float32:
			mongo.FieldConvert[tag] = StringToFloat32
		case reflect.Float64:
			mongo.FieldConvert[tag] = StringToFloat64
		case reflect.Bool:
			mongo.FieldConvert[tag] = StringToBool
		default:
			if field.Type == mongoObjectIDTyp {
				mongo.FieldConvert[tag] = func(str interface{}) interface{} {
					objectID, _ := primitive.ObjectIDFromHex(str.(string))
					return objectID
				}
			} else {
				mongo.FieldConvert[tag] = func(str interface{}) interface{} {
					return str
				}
			}
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
