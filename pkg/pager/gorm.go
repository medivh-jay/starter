package pager

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"starter/pkg/database/orm"
	"strings"
)

// Gorm 查询
type Gorm struct {
	conn  *gorm.DB
	query string
	args  []interface{}
	limit int
	skip  int
	index string
	sorts string
}

func NewGormDriver() *Gorm {
	return &Gorm{
		conn: orm.Slave(),
	}
}

func (orm *Gorm) Where(kv Where) {
	for k, v := range kv {
		orm.query = fmt.Sprintf("AND %s = ? ", k)
		orm.args = append(orm.args, v)
	}
}

func (orm *Gorm) Section(section Section) {
	for k, v := range section {
		if val, ok := v[Gte]; ok {
			orm.query = fmt.Sprintf("AND %s >= ? ", k)
			orm.args = append(orm.args, val)
		}
		if val, ok := v[Lte]; ok {
			orm.query = fmt.Sprintf("AND %s <= ? ", k)
			orm.args = append(orm.args, val)
		}
	}
}

func (orm *Gorm) Limit(limit int) {
	orm.limit = limit
}

func (orm *Gorm) Skip(skip int) {
	orm.skip = skip
}

func (orm *Gorm) Index(index string) {
	orm.index = index
}

func (orm *Gorm) Sort(kv map[string]Sort) {
	for k, v := range kv {
		if v == Asc {
			orm.sorts += fmt.Sprintf("AND %s ASC ", k)
		} else {
			orm.sorts += fmt.Sprintf("AND %s DESC ", k)
		}
	}
	orm.sorts = strings.TrimPrefix(orm.sorts, "AND")
}

func (orm *Gorm) Find(data interface{}) {
	orm.query = strings.TrimPrefix(orm.query, "AND")
	orm.conn.Table(orm.index).Where(orm.query, orm.args...).Limit(orm.limit).Offset(orm.skip).Order(orm.sorts).Find(data)
}

func (orm *Gorm) Count() int {
	var count int
	orm.conn.Table(orm.index).Where(orm.query, orm.args...).Count(&count)
	return count
}
