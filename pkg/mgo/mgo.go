package mgo

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"starter/pkg/app"
	"starter/pkg/unique"
	"strings"
	"time"
)

type collection struct {
	Database *mgo.Session
	Table    *mgo.Collection
	Session  *mgo.Database
	filter   bson.M
	limit    int
	skip     int
	sort     []string
	fields   bson.M
}

type config struct {
	Url             string `toml:"url"`
	Database        string `toml:"database"`
	MaxConnIdleTime int    `toml:"max_conn_idle_time"`
	MaxPoolSize     int    `toml:"max_pool_size"`
	Username        string `toml:"username"`
	Password        string `toml:"password"`
}

var db *mgo.Session
var conf config

func Start() {
	_ = app.Config().Bind("application", "mongo", &conf)
	dialInfo := &mgo.DialInfo{
		Addrs:     []string{strings.ReplaceAll(conf.Url, "mongodb://", "")},
		Direct:    false,
		Timeout:   time.Second * 5,
		PoolLimit: conf.MaxPoolSize,
		Username:  conf.Username,
		Password:  conf.Password,
	}
	var err error
	db, err = mgo.DialWithInfo(dialInfo)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
	} else {
		db.SetMode(mgo.Monotonic, true)
	}
}

// 得到一个mongo操作对象
// 请显式调用 Close 方法释放session
func Collection(table app.Table) *collection {
	database := db.Copy()
	session := database.DB(conf.Database)
	return &collection{
		Database: database,
		Session:  session,
		Table:    session.C(table.TableName()),
		filter:   make(bson.M),
	}

}

func (collection *collection) Where(m bson.M) *collection {
	collection.filter = m
	return collection
}

func (collection *collection) Close() {
	collection.Database.Close()
}

// 限制条数
func (collection *collection) Limit(n int) *collection {
	collection.limit = n
	return collection
}

// 跳过条数
func (collection *collection) Skip(n int) *collection {
	collection.skip = n
	return collection
}

// 排序 bson.M{"created_at":-1}
func (collection *collection) Sort(sorts ...string) *collection {
	collection.sort = sorts
	return collection
}

// 指定查询字段
func (collection *collection) Fields(fields bson.M) *collection {
	collection.fields = fields
	return collection
}

// 写入单条数据
func (collection *collection) InsertOne(document interface{}) (interface{}, error) {
	data := BeforeCreate(document)
	err := collection.Table.Insert(data)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
	}
	return data, err
}

// 写入多条数据
func (collection *collection) InsertMany(documents interface{}) interface{} {
	var data []interface{}
	data = BeforeCreate(documents).([]interface{})
	err := collection.Table.Insert(data)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
	}
	return data
}

// 存在更新,不存在写入, documents 里边的文档需要有 _id 的存在
func (collection *collection) UpdateOrInsert(document interface{}) *mgo.ChangeInfo {
	result, err := collection.Table.Upsert(collection.filter, document)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
	}
	return result
}

//
func (collection *collection) UpdateOne(document interface{}) bool {
	err := collection.Table.Update(collection.filter, bson.M{"$set": BeforeUpdate(document)})
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
	}
	return err == nil
}

//
func (collection *collection) UpdateMany(document interface{}) *mgo.ChangeInfo {
	result, err := collection.Table.UpdateAll(collection.filter, bson.M{"$set": BeforeUpdate(document)})
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
	}
	return result
}

// 查询一条数据
func (collection *collection) FindOne(document interface{}) error {
	err := collection.Table.Find(collection.filter).Select(collection.fields).One(document)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
		return err
	}
	return nil
}

// 查询多条数据
func (collection *collection) FindMany(documents interface{}) {
	err := collection.Table.Find(collection.filter).Skip(collection.skip).Limit(collection.limit).Sort(collection.sort...).Select(collection.fields).All(documents)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
	}
}

// 删除数据,并返回删除成功的数量
func (collection *collection) Delete() bool {
	if collection.filter == nil || len(collection.filter) == 0 {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error("you can't delete all documents, it's very dangerous")
		return false
	}
	err := collection.Table.Remove(collection.filter)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
	}
	return err == nil
}

func (collection *collection) Count() int64 {
	count, err := collection.Table.Find(collection.filter).Count()
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
		return 0
	}
	return int64(count)
}

func BeforeCreate(document interface{}) interface{} {
	val := reflect.ValueOf(document)
	typ := reflect.TypeOf(document)

	switch typ.Kind() {
	case reflect.Ptr:
		return BeforeCreate(val.Elem().Interface())

	case reflect.Array, reflect.Slice:
		var sliceData = make([]interface{}, val.Len(), val.Cap())
		for i := 0; i < val.Len(); i++ {
			sliceData[i] = BeforeCreate(val.Index(i).Interface()).(bson.M)
		}
		return sliceData

	case reflect.Struct:
		var data = make(bson.M)
		for i := 0; i < typ.NumField(); i++ {
			data[typ.Field(i).Tag.Get("bson")] = val.Field(i).Interface()
		}
		if val.FieldByName("Id").Type() == reflect.TypeOf(bson.ObjectId("")) {
			data["_id"] = primitive.NewObjectID()
		}

		if val.FieldByName("Id").Kind() == reflect.String && val.FieldByName("Id").Interface() == "" {
			data["_id"] = primitive.NewObjectID().Hex()
		}

		if IsIntn(val.FieldByName("Id").Kind()) && val.FieldByName("Id").Interface() == 0 {
			data["_id"] = unique.Id()
		}

		now := time.Now().Unix()
		data["created_at"] = now
		data["updated_at"] = now
		return data

	default:
		if val.Type() == reflect.TypeOf(bson.M{}) {
			if !val.MapIndex(reflect.ValueOf("_id")).IsValid() {
				val.SetMapIndex(reflect.ValueOf("_id"), reflect.ValueOf(bson.NewObjectId()))
			}
			val.SetMapIndex(reflect.ValueOf("created_at"), reflect.ValueOf(time.Now().Unix()))
			val.SetMapIndex(reflect.ValueOf("updated_at"), reflect.ValueOf(time.Now().Unix()))
		}
		return val.Interface()
	}
}

func IsIntn(p reflect.Kind) bool {
	return p == reflect.Int || p == reflect.Int64 || p == reflect.Uint64 || p == reflect.Uint32
}

func BeforeUpdate(document interface{}) interface{} {
	val := reflect.ValueOf(document)
	typ := reflect.TypeOf(document)

	switch typ.Kind() {
	case reflect.Ptr:
		return BeforeUpdate(val.Elem().Interface())

	case reflect.Array, reflect.Slice:
		var sliceData = make([]interface{}, val.Len(), val.Cap())
		for i := 0; i < val.Len(); i++ {
			sliceData[i] = BeforeUpdate(val.Index(i).Interface()).(bson.M)
		}
		return sliceData

	case reflect.Struct:
		var data = make(bson.M)
		for i := 0; i < typ.NumField(); i++ {
			if !isZero(val.Field(i)) {
				data[typ.Field(i).Tag.Get("bson")] = val.Field(i).Interface()
			}
		}
		data["updated_at"] = time.Now().Unix()
		return data

	default:
		if val.Type() == reflect.TypeOf(bson.M{}) {
			val.SetMapIndex(reflect.ValueOf("updated_at"), reflect.ValueOf(time.Now().Unix()))
		}
		return val.Interface()
	}
}

func isZero(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}
	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}
