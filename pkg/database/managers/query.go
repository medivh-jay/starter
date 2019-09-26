package managers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mgoBson "gopkg.in/mgo.v2/bson"
	"reflect"
	orm2 "starter/pkg/database/orm"
	"strconv"
	"strings"
)

// MysqlQuery MySQL List 列表的 query 参数解析
type MysqlQuery struct {
	entityTyp reflect.Type
	statement string
	params    []interface{}
}

var databaseTyp = reflect.TypeOf(orm2.Database{})

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

// GetQuery 获取解析之后的SQL 和 对应参数
func (query *MysqlQuery) GetQuery(ctx *gin.Context) (statement string, params []interface{}) {
	query.statement = ""
	query.params = make([]interface{}, 0, 0)
	after, before := ctx.DefaultQuery("after_id", ""), ctx.DefaultQuery("before_id", "")
	if after != "" {
		query.statement = "id > ?"
		query.params = append(params, after)
	}
	if before != "" {
		query.statement = "id < ?"
		query.params = append(params, before)
	}

	query.createTagQuery(databaseTyp, ctx)
	query.createTagQuery(query.entityTyp, ctx)

	return query.statement, query.params
}

// Limit 设置限制条数
func (query *MysqlQuery) Limit(ctx *gin.Context) int {
	limit := ctx.DefaultQuery("limit", "10")
	num, _ := strconv.Atoi(limit)
	return num
}

// Offset 不提供 offset 操作
func (query *MysqlQuery) Offset(ctx *gin.Context) int {
	page := ctx.DefaultQuery("page", "1")
	num, _ := strconv.Atoi(page)
	return (num - 1) * (query.Limit(ctx))
}

// MongoQuery Mongo 的 List 数据 query 解析
type MongoQuery struct {
	entityTyp reflect.Type
	query     bson.M
}

// GetQuery 获取参数
func (query *MongoQuery) GetQuery(ctx *gin.Context) bson.M {

	query.query = make(bson.M)
	after, before := ctx.DefaultQuery("after_id", ""), ctx.DefaultQuery("before_id", "")
	if after != "" {
		query.query["_id"] = bson.M{"$gt": query.convertID(after)}
	}
	if before != "" {
		query.query["_id"] = bson.M{"$lt": query.convertID(before)}
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
					objectID, _ := primitive.ObjectIDFromHex(val)
					query.query[field] = objectID
				} else {
					query.query[field] = val
				}
			}
		}
	}
}

// Limit 限制条数
func (query *MongoQuery) Limit(ctx *gin.Context) int {
	limit := ctx.DefaultQuery("limit", "10")
	num, _ := strconv.Atoi(limit)
	return num
}

func (query *MongoQuery) convertID(id string) interface{} {
	fieldTyp, exists := query.entityTyp.FieldByName("ID")
	if !exists {
		return id
	}

	if fieldTyp.Type == reflect.TypeOf(primitive.ObjectID{}) {
		objectID, _ := primitive.ObjectIDFromHex(id)
		return objectID
	}

	return id
}

// Offset 不提供 offset 操作
func (query *MongoQuery) Offset(ctx *gin.Context) int {
	page := ctx.DefaultQuery("page", "1")
	num, _ := strconv.Atoi(page)
	return (num - 1) * (query.Limit(ctx))
	//return 0
}

// MgoQuery mgo 的 List query 参数解析
type MgoQuery struct {
	entityTyp reflect.Type
	query     mgoBson.M
}

// GetQuery 获取查询参数
func (query *MgoQuery) GetQuery(ctx *gin.Context) mgoBson.M {

	query.query = make(mgoBson.M)
	after, before := ctx.DefaultQuery("after_id", ""), ctx.DefaultQuery("before_id", "")
	if after != "" {
		query.query["_id"] = bson.M{"$gt": query.convertID(after)}
	}
	if before != "" {
		query.query["_id"] = bson.M{"$lt": query.convertID(before)}
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
					objectID := mgoBson.ObjectIdHex(val)
					query.query[field] = objectID
				} else {
					query.query[field] = val
				}
			}
		}
	}
}

// Limit 限制条数
func (query *MgoQuery) Limit(ctx *gin.Context) int {
	limit := ctx.DefaultQuery("limit", "10")
	num, _ := strconv.Atoi(limit)
	return num
}

func (query *MgoQuery) convertID(id string) interface{} {
	fieldTyp, exists := query.entityTyp.FieldByName("ID")
	if !exists {
		return id
	}

	if fieldTyp.Type == reflect.TypeOf(mgoBson.ObjectId("")) {
		objectID := mgoBson.ObjectIdHex(id)
		return objectID
	}

	return id
}

// Offset 不提供 offset 操作
func (query *MgoQuery) Offset(ctx *gin.Context) int {
	page := ctx.DefaultQuery("page", "1")
	num, _ := strconv.Atoi(page)
	return (num - 1) * (query.Limit(ctx))
}
