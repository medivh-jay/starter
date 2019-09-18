// url(?||&)sorts=-name,age,+level
package managers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"starter/pkg/app"
	"strings"
)

type Sort interface {
	Parse(ctx *gin.Context) interface{}
}

func NewSorter(driver EntityTyp) Sort {
	switch driver {
	case Mysql:
		return new(mysqlSort)
	case Mongo:
		return new(mongoSort)
	case Mgo:
		return new(mgoSort)
	default:
		app.Logger().WithField("log_type", "pkg.managers.sort").Panic("driver not found")
		return nil
	}
}

type (
	mysqlSort struct {
	}
	mongoSort struct {
	}
	mgoSort struct {
	}
)

var sortKey = "sorts"

func parse(ctx *gin.Context) map[string]int {
	sortVal := ctx.Query(sortKey)
	if sortVal == "" {
		return nil
	}

	sorts := strings.Split(sortVal, ",")

	var sortFields = make(map[string]int)
	for _, field := range sorts {
		switch {
		case strings.HasPrefix(field, "-"):
			sortFields[strings.TrimPrefix(field, "-")] = -1
		case strings.HasPrefix(field, "+"):
			sortFields[strings.TrimPrefix(field, "+")] = 1
		default:
			sortFields[field] = 1
		}
	}

	return sortFields
}

func (mysql *mysqlSort) Parse(ctx *gin.Context) interface{} {
	sorts := parse(ctx)
	var order string
	for field, sort := range sorts {
		if sort == -1 {
			order = order + field + " DESC,"
		} else {
			order = order + field + " ASC,"
		}
	}

	return strings.TrimSuffix(order, ",")
}

func (mysql *mongoSort) Parse(ctx *gin.Context) interface{} {
	var order = make(bson.M)
	sorts := parse(ctx)
	for field, sort := range sorts {
		order[field] = sort
	}

	return order
}

func (mysql *mgoSort) Parse(ctx *gin.Context) interface{} {
	var sorts = parse(ctx)
	var order = make([]string, 0, 0)
	for field, sort := range sorts {
		if sort == -1 {
			order = append(order, fmt.Sprintf("-%s", field))
		} else {
			order = append(order, field)
		}
	}

	return order
}
