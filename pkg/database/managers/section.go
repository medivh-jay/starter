package managers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	mgoBson "gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
)

type (
	// ParseInterface List 接口区间查询解析
	ParseInterface interface {
		// 解析为数据库查询语句, mongo 返回 bson.M ， 例如 { "key":{"$lte": 123, "$get": 111} }
		// MySQL 返回 SQL 字符串 例如 key >= 111 and key <= 123
		Parse(engine EntityTyp) interface{}
		// 设置要进行查询的字段
		SetKey(key string)
		// 获取字段
		GetKey() string
		// 设置区间开始值
		SetBegin(begin int)
		// 设置区间结束值
		SetEnd(end int)
	}

	// Sections 数据库字段区间查询语句生成和解析
	Sections struct {
		queries []ParseInterface
		Engine  EntityTyp
	}

	// GteQuery 大于等于
	GteQuery struct {
		Key   string // 区间查询 key
		Begin int    // 区间查询开始值
	}

	// LteQuery 小于等于
	LteQuery struct {
		Key string // 区间查询 key
		End int    // 区间查询结束值
	}

	// GteLteQuery 大于等于和小于等于
	GteLteQuery struct {
		Key   string // 区间查询 key
		Begin int    // 区间查询开始值
		End   int    // 区间查询结束值
	}
)

// SetKey 设置查询的字段
func (gteQuery *GteQuery) SetKey(key string) { gteQuery.Key = key }

// GetKey 获取查询的字段
func (gteQuery *GteQuery) GetKey() string { return gteQuery.Key }

// SetBegin 设置区间开始值
func (gteQuery *GteQuery) SetBegin(begin int) { gteQuery.Begin = begin }

// SetEnd 设置区间结束值
func (gteQuery *GteQuery) SetEnd(end int) {}

// SetKey 设置查询的字段
func (lteQuery *LteQuery) SetKey(key string) { lteQuery.Key = key }

// GetKey 获取查询的字段
func (lteQuery *LteQuery) GetKey() string { return lteQuery.Key }

// SetBegin 设置区间开始值
func (lteQuery *LteQuery) SetBegin(begin int) {}

// SetEnd 设置区间结束值
func (lteQuery *LteQuery) SetEnd(end int) { lteQuery.End = end }

// SetKey 设置查询的字段
func (gteLteQuery *GteLteQuery) SetKey(key string) { gteLteQuery.Key = key }

// GetKey 获取查询的字段
func (gteLteQuery *GteLteQuery) GetKey() string { return gteLteQuery.Key }

// SetBegin 设置区间开始值
func (gteLteQuery *GteLteQuery) SetBegin(begin int) { gteLteQuery.Begin = begin }

// SetEnd 设置区间结束值
func (gteLteQuery *GteLteQuery) SetEnd(end int) { gteLteQuery.End = end }

// Parse 转为数据库查询语句
func (sections Sections) Parse() interface{} {
	switch sections.Engine {
	case Mongo:
		query := bson.M{}
		for _, item := range sections.queries {
			parse := item.Parse(sections.Engine)
			query[item.GetKey()] = parse
		}
		return query
	case Mgo:
		query := mgoBson.M{}
		for _, item := range sections.queries {
			parse := item.Parse(sections.Engine)
			query[item.GetKey()] = parse
		}
		return query
	case Mysql:
		query := ""
		for _, item := range sections.queries {
			parse := item.Parse(sections.Engine).(string)
			if query == "" {
				query = parse
			} else {
				query += " and " + parse
			}
		}
		return query
	}

	return nil
}

// Parse 转为指定数据库驱动的查询语句
func (gteQuery *GteQuery) Parse(engine EntityTyp) interface{} {
	switch engine {
	case Mysql:
		return fmt.Sprintf("%s >= %d", gteQuery.Key, gteQuery.Begin)
	case Mongo:
		return bson.M{"$gte": gteQuery.Begin}
	case Mgo:
		return mgoBson.M{"$gte": gteQuery.Begin}
	}
	return nil
}

// Parse 转为指定数据库驱动的查询语句
func (lteQuery *LteQuery) Parse(engine EntityTyp) interface{} {
	switch engine {
	case Mysql:
		return fmt.Sprintf("%s <= %d", lteQuery.Key, lteQuery.End)
	case Mongo:
		return bson.M{"$lte": lteQuery.End}
	case Mgo:
		return mgoBson.M{"$lte": lteQuery.End}
	}
	return nil
}

// Parse 转为指定数据库驱动的查询语句
func (gteLteQuery *GteLteQuery) Parse(engine EntityTyp) interface{} {
	switch engine {
	case Mysql:
		return fmt.Sprintf("%s >= %d and %s <= %d", gteLteQuery.Key, gteLteQuery.Begin, gteLteQuery.Key, gteLteQuery.End)
	case Mongo:
		return bson.M{"$gte": gteLteQuery.Begin, "$lte": gteLteQuery.End}
	case Mgo:
		return mgoBson.M{"$gte": gteLteQuery.Begin, "$lte": gteLteQuery.End}
	}
	return nil
}

// ParseSectionParams 从URL中取出section，并解析为Sections结构体
// 支持多个字段
//  格式: 大于等于 section=key:value
//  格式: 小于等于 section=-key:value
//  格式: 大于等于和小于等于 section=key:value1,value2  value1和value2的值不用进行大小排序，程序将始终取两值最小值作为大于等于的值，最大值作为小于等于的值
//  例子: ?section=created_at:1,100&section=updated_at:50&section=-age:60 表示区间筛选条件为 created_at 大于1 小于 100 并且 updated_at 大于 50 并且 age 小于 60
//
//  注意: 值类型必须是整型
func ParseSectionParams(c *gin.Context, typ EntityTyp) Sections {
	sections := c.Request.URL.Query()["section"]
	var sq = Sections{queries: make([]ParseInterface, len(sections))}
	sq.Engine = typ
	i := 0

	for _, section := range sections {
		sectionKey, sectionValue := strings.Split(section, ":")[0], strings.Split(section, ":")[1]

		switch true {

		case strings.Index(sectionKey, "-") == 0:
			sq.queries[i] = &LteQuery{}
			sq.queries[i].SetKey(string([]rune(sectionKey)[1:]))
			end, err := strconv.Atoi(sectionValue)
			if err != nil {
				end = 0
			}
			sq.queries[i].SetEnd(end)

		default:
			values := strings.Split(sectionValue, ",")
			if len(values) == 2 {
				sq.queries[i] = &GteLteQuery{}
				sq.queries[i].SetKey(sectionKey)
				if values[0] > values[1] {
					values[0], values[1] = values[1], values[0]
				}
				begin, beginErr := strconv.Atoi(values[0])
				if beginErr != nil {
					begin = 0
				}
				end, endErr := strconv.Atoi(values[1])
				if endErr != nil {
					end = 0
				}
				sq.queries[i].SetBegin(begin)
				sq.queries[i].SetEnd(end)
			} else {
				sq.queries[i] = &GteQuery{}
				sq.queries[i].SetKey(sectionKey)
				begin, err := strconv.Atoi(values[0])
				if err != nil {
					begin = 0
				}
				sq.queries[i].SetBegin(begin)
			}
		}

		i++
	}

	return sq
}

func mergeMgo(query mgoBson.M, section map[string]interface{}) mgoBson.M {
	for k, v := range section {
		query[k] = v
	}

	return query
}

func mergeMongo(query bson.M, section map[string]interface{}) bson.M {
	for k, v := range section {
		query[k] = v
	}

	return query
}
