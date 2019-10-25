package pager

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"reflect"
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

// NewGormDriver gorm 分页实现
func NewGormDriver() *Gorm {
	return &Gorm{
		conn: orm.Slave(),
	}
}

// Where 构建查询条件
func (orm *Gorm) Where(kv Where) {
	for k, v := range kv {
		orm.query = fmt.Sprintf("AND %s = ? ", k)
		orm.args = append(orm.args, v)
	}
}

// Section 范围查询条件
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

// Limit 每页数量
func (orm *Gorm) Limit(limit int) {
	orm.limit = limit
}

// Skip 跳过数量
func (orm *Gorm) Skip(skip int) {
	orm.skip = skip
}

// Index table 表名
func (orm *Gorm) Index(index string) {
	orm.index = index
}

// Sort 排序
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

// Find 从数据库查询数据
func (orm *Gorm) Find(data interface{}) {
	orm.query = strings.TrimPrefix(orm.query, "AND")
	orm.conn.Table(orm.index).Where(orm.query, orm.args...).Limit(orm.limit).Offset(orm.skip).Order(orm.sorts).Find(data)
}

// SetTyp sql 对数据查询的类型不敏感
func (orm *Gorm) SetTyp(typ reflect.Type) {
	return
}

// Count 计算指定查询条件的总数量
func (orm *Gorm) Count() int {
	var count int
	orm.conn.Table(orm.index).Where(orm.query, orm.args...).Count(&count)
	return count
}
