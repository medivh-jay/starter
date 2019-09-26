package mgo

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"starter/pkg/app"
	"starter/pkg/database"
	"starter/pkg/unique"
	"strings"
	"time"
)

// CollectionInfo 集合包含的连接信息和查询等操作信息
type CollectionInfo struct {
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
	URL             string `toml:"url"`
	Database        string `toml:"database"`
	MaxConnIdleTime int    `toml:"max_conn_idle_time"`
	MaxPoolSize     int    `toml:"max_pool_size"`
	Username        string `toml:"username"`
	Password        string `toml:"password"`
}

var db *mgo.Session
var conf config

// Start 连接到mongo
func Start() {
	_ = app.Config().Bind("application", "mongo", &conf)
	dialInfo := &mgo.DialInfo{
		Addrs:     []string{strings.ReplaceAll(conf.URL, "mongodb://", "")},
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

// GetDB 获得一个 mgo 的 session, 获取使用之后需要显式调用 Close 关闭
func GetDB() *mgo.Session {
	return db.Copy()
}

// Collection 得到一个mongo操作对象
// 请显式调用 Close 方法释放session
func Collection(table database.Table) *CollectionInfo {
	clone := db.Copy()
	session := clone.DB(conf.Database)
	return &CollectionInfo{
		Database: clone,
		Session:  session,
		Table:    session.C(table.TableName()),
		filter:   make(bson.M),
	}

}

// Where where 条件查询
func (collection *CollectionInfo) Where(m bson.M) *CollectionInfo {
	collection.filter = m
	return collection
}

// Close 关闭session
func (collection *CollectionInfo) Close() {
	collection.Database.Close()
}

// Limit 限制条数
func (collection *CollectionInfo) Limit(n int) *CollectionInfo {
	collection.limit = n
	return collection
}

// Skip 跳过条数
func (collection *CollectionInfo) Skip(n int) *CollectionInfo {
	collection.skip = n
	return collection
}

// Sort 排序 bson.M{"created_at":-1}
func (collection *CollectionInfo) Sort(sorts ...string) *CollectionInfo {
	collection.sort = sorts
	return collection
}

// Fields 指定查询字段
func (collection *CollectionInfo) Fields(fields bson.M) *CollectionInfo {
	collection.fields = fields
	return collection
}

// InsertOne 写入单条数据
func (collection *CollectionInfo) InsertOne(document interface{}) (interface{}, error) {
	data := BeforeCreate(document)
	err := collection.Table.Insert(data)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
	}
	return data, err
}

// InsertMany 写入多条数据
func (collection *CollectionInfo) InsertMany(documents interface{}) interface{} {
	var data []interface{}
	data = BeforeCreate(documents).([]interface{})
	err := collection.Table.Insert(data)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
	}
	return data
}

// UpdateOrInsert 存在更新,不存在写入, documents 里边的文档需要有 _id 的存在
func (collection *CollectionInfo) UpdateOrInsert(document interface{}) *mgo.ChangeInfo {
	result, err := collection.Table.Upsert(collection.filter, document)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
	}
	return result
}

// UpdateOne 更新一条
func (collection *CollectionInfo) UpdateOne(document interface{}) bool {
	err := collection.Table.Update(collection.filter, bson.M{"$set": BeforeUpdate(document)})
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
	}
	return err == nil
}

// UpdateMany 更新多条
func (collection *CollectionInfo) UpdateMany(document interface{}) *mgo.ChangeInfo {
	result, err := collection.Table.UpdateAll(collection.filter, bson.M{"$set": BeforeUpdate(document)})
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
	}
	return result
}

// FindOne 查询一条数据
func (collection *CollectionInfo) FindOne(document interface{}) error {
	err := collection.Table.Find(collection.filter).Select(collection.fields).One(document)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
		return err
	}
	return nil
}

// FindMany 查询多条数据
func (collection *CollectionInfo) FindMany(documents interface{}) {
	err := collection.Table.Find(collection.filter).Skip(collection.skip).Limit(collection.limit).Sort(collection.sort...).Select(collection.fields).All(documents)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
	}
}

// Delete 删除数据,并返回删除成功的数量
func (collection *CollectionInfo) Delete() bool {
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

// Count 获取指定条件下总条数
func (collection *CollectionInfo) Count() int64 {
	count, err := collection.Table.Find(collection.filter).Count()
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mgo.mgo").Error(err)
		return 0
	}
	return int64(count)
}

// BeforeCreate 在创建文档之前进行id和时间等赋值操作
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
		if val.FieldByName("ID").Type() == reflect.TypeOf(bson.ObjectId("")) {
			data["_id"] = primitive.NewObjectID()
		}

		if val.FieldByName("ID").Kind() == reflect.String && val.FieldByName("ID").Interface() == "" {
			data["_id"] = primitive.NewObjectID().Hex()
		}

		if IsIntn(val.FieldByName("ID").Kind()) && val.FieldByName("ID").Interface() == 0 {
			data["_id"] = unique.ID()
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

// IsIntn 类型是否是数字
func IsIntn(p reflect.Kind) bool {
	return p == reflect.Int || p == reflect.Int64 || p == reflect.Uint64 || p == reflect.Uint32
}

// BeforeUpdate 在更新前的操作
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
