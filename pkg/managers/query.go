package managers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mgoBson "gopkg.in/mgo.v2/bson"
	"reflect"
	"starter/pkg/orm"
	"strconv"
	"strings"
)

type MysqlQuery struct {
	entityTyp reflect.Type
	statement string
	params    []interface{}
}

var databaseTyp = reflect.TypeOf(orm.Database{})

func (query *MysqlQuery) createTagQuery(entityTyp reflect.Type, ctx *gin.Context) {
	for i := 0; i < entityTyp.NumField(); i++ {
		tagInfo := strings.Split(entityTyp.Field(i).Tag.Get("gorm"), ";")
		for _, v := range tagInfo {
			if strings.Index(v, "column:") == 0 {
				tag := strings.ReplaceAll(v, "column:", "")
				value := ctx.Query(tag)
				if value != "" {
					if query.statement == "" {
						query.statement = fmt.Sprintf("%s = ?", tag)
					} else {
						query.statement = query.statement + fmt.Sprintf(" and %s = ?", tag)
					}
					query.params = append(query.params, value)
				}
			}
		}
	}
}

func (query *MysqlQuery) GetQuery(ctx *gin.Context) (statement string, params []interface{}) {
	query.statement = ""
	query.params = make([]interface{}, 0, 0)
	nextId := ctx.DefaultQuery("next_id", "")
	if nextId != "" {
		query.statement = "id > ?"
		query.params = append(params, nextId)
	}

	query.createTagQuery(databaseTyp, ctx)
	query.createTagQuery(query.entityTyp, ctx)

	return query.statement, query.params
}

func (query *MysqlQuery) Limit(ctx *gin.Context) int {
	limit := ctx.DefaultQuery("limit", "10")
	num, _ := strconv.Atoi(limit)
	return num
}

// 不提供 offset 操作
func (query *MysqlQuery) Offset(ctx *gin.Context) int {
	return 0
}

type MongoQuery struct {
	entityTyp reflect.Type
	query     bson.M
}

func (query *MongoQuery) GetQuery(ctx *gin.Context) bson.M {

	query.query = make(bson.M)
	nextId := ctx.DefaultQuery("next_id", "")
	if nextId != "" {
		query.query["_id"] = bson.M{"$gt": query.convertId(nextId)}
	}

	query.createTagQuery(query.entityTyp, ctx)
	return query.query
}

func (query *MongoQuery) createTagQuery(entityTyp reflect.Type, ctx *gin.Context) {
	for i := 0; i < entityTyp.NumField(); i++ {
		field := entityTyp.Field(i).Tag.Get("bson")
		val := ctx.DefaultQuery(field, "")
		if val != "" {
			switch entityTyp.Field(i).Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				query.query[field], _ = strconv.Atoi(val)
			case reflect.Bool:
				query.query[field] = val != ""
			default:
				if entityTyp.Field(i).Type == reflect.TypeOf(primitive.ObjectID{}) {
					objectId, _ := primitive.ObjectIDFromHex(val)
					query.query[field] = objectId
				} else {
					query.query[field] = val
				}
			}
		}
	}
}

func (query *MongoQuery) Limit(ctx *gin.Context) int {
	limit := ctx.DefaultQuery("limit", "10")
	num, _ := strconv.Atoi(limit)
	return num
}

func (query *MongoQuery) convertId(id string) interface{} {
	fieldTyp, exists := query.entityTyp.FieldByName("Id")
	if !exists {
		return id
	}

	if fieldTyp.Type == reflect.TypeOf(primitive.ObjectID{}) {
		objectId, _ := primitive.ObjectIDFromHex(id)
		return objectId
	}

	return id
}

// 不提供 offset 操作
func (query *MongoQuery) Offset(ctx *gin.Context) int {
	return 0
}

type MgoQuery struct {
	entityTyp reflect.Type
	query     mgoBson.M
}

func (query *MgoQuery) GetQuery(ctx *gin.Context) mgoBson.M {

	query.query = make(mgoBson.M)
	nextId := ctx.DefaultQuery("next_id", "")
	if nextId != "" {
		query.query["_id"] = bson.M{"$gt": query.convertId(nextId)}
	}

	query.createTagQuery(query.entityTyp, ctx)
	return query.query
}

func (query *MgoQuery) createTagQuery(entityTyp reflect.Type, ctx *gin.Context) {
	for i := 0; i < entityTyp.NumField(); i++ {
		field := entityTyp.Field(i).Tag.Get("bson")
		val := ctx.DefaultQuery(field, "")
		if val != "" {
			switch entityTyp.Field(i).Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				query.query[field], _ = strconv.Atoi(val)
			case reflect.Bool:
				query.query[field] = val != ""
			default:
				if entityTyp.Field(i).Type == reflect.TypeOf(mgoBson.ObjectId("")) {
					objectId := mgoBson.ObjectIdHex(val)
					query.query[field] = objectId
				} else {
					query.query[field] = val
				}
			}
		}
	}
}

func (query *MgoQuery) Limit(ctx *gin.Context) int {
	limit := ctx.DefaultQuery("limit", "10")
	num, _ := strconv.Atoi(limit)
	return num
}

func (query *MgoQuery) convertId(id string) interface{} {
	fieldTyp, exists := query.entityTyp.FieldByName("Id")
	if !exists {
		return id
	}

	if fieldTyp.Type == reflect.TypeOf(mgoBson.ObjectId("")) {
		objectId := mgoBson.ObjectIdHex(id)
		return objectId
	}

	return id
}

// 不提供 offset 操作
func (query *MgoQuery) Offset(ctx *gin.Context) int {
	return 0
}
