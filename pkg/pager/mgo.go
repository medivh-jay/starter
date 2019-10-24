package pager

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"starter/pkg/app"
	db "starter/pkg/database/mgo"
)

var mgoObjectIDTyp = reflect.TypeOf(bson.ObjectId(""))

// Mgo 查询
type Mgo struct {
	where Where
	limit int
	skip  int
	index string
	sorts []string
	// 字段的类型转换操作
	FieldConvert map[string]func(str interface{}) interface{}
}

// NewMgoDriver mgo 驱动支持
func NewMgoDriver() *Mgo {
	return &Mgo{FieldConvert: make(map[string]func(str interface{}) interface{})}
}

// Where 写入查询条件
func (mgo *Mgo) Where(kv Where) {
	if mgo.where == nil {
		mgo.where = make(Where)
	}
	mgo.where = mergeWhere(mgo.where, kv)
	for k, v := range mgo.where {
		if convert, ok := mgo.FieldConvert[k]; ok {
			mgo.where[k] = convert(v)
		}
	}
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
}

// SetTyp 记录结构体类型
//  从 bson tag 获取数据库字段值
func (mgo *Mgo) SetTyp(typ reflect.Type) {
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("bson")
		switch field.Type.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			mgo.FieldConvert[tag] = StringToInt
		case reflect.Float32:
			mgo.FieldConvert[tag] = StringToFloat32
		case reflect.Float64:
			mgo.FieldConvert[tag] = StringToFloat64
		case reflect.Bool:
			mgo.FieldConvert[tag] = StringToBool
		default:
			if field.Type == mgoObjectIDTyp {
				mgo.FieldConvert[tag] = func(str interface{}) interface{} {
					return bson.ObjectIdHex(str.(string))
				}
			} else {
				mgo.FieldConvert[tag] = func(str interface{}) interface{} {
					return str
				}
			}
		}
	}
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
