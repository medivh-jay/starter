//  三方 mgo 太长时间不维护了
//  官方 mongo 驱动很不友好
//  所以这里稍微对常用方法做了处理,可以直接调用这里的方法进行一些常规操作
//  复杂的操作,调用这里的 Collection 之后可获取里边的 Database 属性 和 Table 属性操作
//  这里的添加和修改操作将会自动补全 created_at updated_at 和 _id
package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"starter/pkg/app"
	"starter/pkg/database"
	"starter/pkg/unique"
	"time"
)

var (
	client *mongo.Client
	conf   config
)

type (
	collection struct {
		Database *mongo.Database
		Table    *mongo.Collection
		filter   bson.M
		limit    int64
		skip     int64
		sort     bson.M
		fields   bson.M
	}

	config struct {
		Url             string `toml:"url"`
		Database        string `toml:"database"`
		MaxConnIdleTime int    `toml:"max_conn_idle_time"`
		MaxPoolSize     int    `toml:"max_pool_size"`
		Username        string `toml:"username"`
		Password        string `toml:"password"`
	}
)

// 启动 mongo
func Start() {
	_ = app.Config().Bind("application", "mongo", &conf)
	var err error
	mongoOptions := options.Client()
	mongoOptions.SetMaxConnIdleTime(time.Duration(conf.MaxConnIdleTime) * time.Second)
	mongoOptions.SetMaxPoolSize(uint16(conf.MaxPoolSize))
	if conf.Username != "" && conf.Password != "" {
		mongoOptions.SetAuth(options.Credential{Username: conf.Username, Password: conf.Password})
	}

	client, err = mongo.NewClient(mongoOptions.ApplyURI(conf.Url))
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mongo.mongo").Error(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mongo.mongo").Error(err)
	}
}

// 得到一个mongo操作对象
func Collection(table database.Table) *collection {
	db := client.Database(conf.Database)
	return &collection{
		Database: db,
		Table:    db.Collection(table.TableName()),
		filter:   make(bson.M),
	}
}

// 条件查询, bson.M{"field": "value"}
func (collection *collection) Where(m bson.M) *collection {
	collection.filter = m
	return collection
}

// 限制条数
func (collection *collection) Limit(n int64) *collection {
	collection.limit = n
	return collection
}

// 跳过条数
func (collection *collection) Skip(n int64) *collection {
	collection.skip = n
	return collection
}

// 排序 bson.M{"created_at":-1}
func (collection *collection) Sort(sorts bson.M) *collection {
	collection.sort = sorts
	return collection
}

// 指定查询字段
func (collection *collection) Fields(fields bson.M) *collection {
	collection.fields = fields
	return collection
}

// 写入单条数据
func (collection *collection) InsertOne(document interface{}) *mongo.InsertOneResult {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.Table.InsertOne(ctx, BeforeCreate(document))
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mongo.mongo").Error(err)
	}
	return result
}

// 写入多条数据
func (collection *collection) InsertMany(documents interface{}) *mongo.InsertManyResult {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	var data []interface{}
	data = BeforeCreate(documents).([]interface{})
	result, err := collection.Table.InsertMany(ctx, data)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mongo.mongo").Error(err)
	}
	return result
}

// 存在更新,不存在写入, documents 里边的文档需要有 _id 的存在
func (collection *collection) UpdateOrInsert(documents []interface{}) *mongo.UpdateResult {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	var upsert = true
	result, err := collection.Table.UpdateMany(ctx, documents, &options.UpdateOptions{Upsert: &upsert})
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mongo.mongo").Error(err)
	}
	return result
}

//
func (collection *collection) UpdateOne(document interface{}) *mongo.UpdateResult {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.Table.UpdateOne(ctx, collection.filter, bson.M{"$set": BeforeUpdate(document)})
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mongo.mongo").Error(err)
	}
	return result
}

//
func (collection *collection) UpdateMany(document interface{}) *mongo.UpdateResult {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.Table.UpdateMany(ctx, collection.filter, bson.M{"$set": BeforeUpdate(document)})
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mongo.mongo").Error(err)
	}
	return result
}

// 查询一条数据
func (collection *collection) FindOne(document interface{}) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result := collection.Table.FindOne(ctx, collection.filter, &options.FindOneOptions{
		Skip:       &collection.skip,
		Sort:       collection.sort,
		Projection: collection.fields,
	})
	err := result.Decode(document)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mongo.mongo").Error(err)
		return err
	}
	return nil
}

// 查询多条数据
func (collection *collection) FindMany(documents interface{}) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.Table.Find(ctx, collection.filter, &options.FindOptions{
		Skip:       &collection.skip,
		Limit:      &collection.limit,
		Sort:       collection.sort,
		Projection: collection.fields,
	})
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mongo.mongo").Error(err)
	}
	defer result.Close(ctx)

	val := reflect.ValueOf(documents)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		app.Logger().WithField("log_type", "pkg.mongo.mongo").Error("result argument must be a slice address")
	}

	slice := reflect.MakeSlice(val.Elem().Type(), 0, 0)

	itemTyp := val.Elem().Type().Elem()
	for result.Next(ctx) {

		item := reflect.New(itemTyp)
		err := result.Decode(item.Interface())
		if err != nil {
			app.Logger().WithField("log_type", "pkg.mongo.mongo").Error(err)
			break
		}

		slice = reflect.Append(slice, reflect.Indirect(item))
	}
	val.Elem().Set(slice)
}

// 删除数据,并返回删除成功的数量
func (collection *collection) Delete() int64 {
	if collection.filter == nil || len(collection.filter) == 0 {
		app.Logger().WithField("log_type", "pkg.mongo.mongo").Error("you can't delete all documents, it's very dangerous")
		return 0
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.Table.DeleteMany(ctx, collection.filter)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mongo.mongo").Error(err)
	}
	return result.DeletedCount
}

func (collection *collection) Count() int64 {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.Table.CountDocuments(ctx, collection.filter)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.mongo.mongo").Error(err)
		return 0
	}
	return result
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
		if val.FieldByName("Id").Type() == reflect.TypeOf(primitive.ObjectID{}) {
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
				val.SetMapIndex(reflect.ValueOf("_id"), reflect.ValueOf(primitive.NewObjectID()))
			}
			val.SetMapIndex(reflect.ValueOf("created_at"), reflect.ValueOf(time.Now().Unix()))
			val.SetMapIndex(reflect.ValueOf("updated_at"), reflect.ValueOf(time.Now().Unix()))
		}
		return val.Interface()
	}
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

func IsIntn(p reflect.Kind) bool {
	return p == reflect.Int || p == reflect.Int64 || p == reflect.Uint64 || p == reflect.Uint32
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
