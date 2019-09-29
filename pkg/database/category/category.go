package category

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"starter/pkg/app"
	"starter/pkg/database"
	"starter/pkg/database/mgo"
	"starter/pkg/database/mongo"
	"starter/pkg/database/orm"
	"strings"
)

type (
	// Node 分类信息
	Node struct {
		Alias       string `json:"alias"`       // 可能是页面链接拼接需要的数据, 比如 items?category={alias} 或者 items/{alias}
		Name        string `json:"name"`        // 分类展示名字
		Pic         string `json:"pic"`         // 图标
		Badge       string `json:"badge"`       // 徽章,可能会存在的角标
		Description string `json:"description"` // 可能存在的简介
		Next        Nodes  `json:"next"`        // 子分类
	}
	// Nodes 指定一级的分类信息
	Nodes []*Node

	// Tree 所有分类原始数据(已排序)
	Tree struct {
		items    reflect.Value
		category *Category
	}

	// Category 分类实例
	Category struct {
		aliasField, nameField, picField, badgeField, levelField, descriptionField, PidField, IDField string
		tableName                                                                                    func() string
		tableTyp                                                                                     reflect.Type
	}

	// Table 表操作
	Table struct {
		category *Category
	}

	// Option 统一设置
	Option func(category *Category)
)

var (
	defaultSetting = []Option{
		SetDescriptionField("Description"), SetAliasField("Alias"), SetBadgeField("Badge"),
		SetLevelField("Level"), SetNameField("Name"), SetPicField("Pic"), SetPidField("Pid"), SetIDField("ID"),
	}
)

// SetDescriptionField 设置详情简介的字段名
func SetDescriptionField(field string) Option {
	return func(category *Category) {
		if category.descriptionField == "" {
			category.descriptionField = field
		}
	}
}

// SetPidField 设置父级标识字段名
func SetPidField(field string) Option {
	return func(category *Category) {
		if category.PidField == "" {
			category.PidField = field
		}
	}
}

// SetIDField 设置id字段名
func SetIDField(field string) Option {
	return func(category *Category) {
		if category.IDField == "" {
			category.IDField = field
		}
	}
}

// SetLevelField 设置分级字段名
func SetLevelField(field string) Option {
	return func(category *Category) {
		if category.levelField == "" {
			category.levelField = field
		}
	}
}

// SetBadgeField 设置角标徽章字段名
func SetBadgeField(field string) Option {
	return func(category *Category) {
		if category.badgeField == "" {
			category.badgeField = field
		}
	}
}

// SetPicField 设置图标字段名
func SetPicField(field string) Option {
	return func(category *Category) {
		if category.picField == "" {
			category.picField = field
		}
	}
}

// SetNameField 设置显示名字字段名
func SetNameField(field string) Option {
	return func(category *Category) {
		if category.nameField == "" {
			category.nameField = field
		}
	}
}

// SetAliasField 设置别名字段字段名
func SetAliasField(field string) Option {
	return func(category *Category) {
		if category.aliasField == "" {
			category.aliasField = field
		}
	}
}

func newItems(typ reflect.Type) reflect.Value {
	slice := reflect.MakeSlice(reflect.SliceOf(typ), 0, 0)
	items := reflect.New(slice.Type())
	items.Elem().Set(slice)
	return items
}

// New 分类对象
func New(options ...Option) *Category {
	var category = new(Category)
	// 先按用户配置对字段赋值
	for _, option := range options {
		option(category)
	}

	// 设置方法内部只会对用户没有配置的字段进行赋值
	for _, option := range defaultSetting {
		option(category)
	}

	return category
}

// Table 获得表结构, 不要传指针
func (category *Category) Table(table database.Table) *Table {
	category.tableName = table.TableName
	category.tableTyp = reflect.TypeOf(table)
	return &Table{category: category}
}

// WithMgo 使用 mgo 驱动库
func (table *Table) WithMgo() *Tree {
	collection := reflect.New(table.category.tableTyp).Interface().(database.Table)
	items := newItems(table.category.tableTyp)
	field, _ := table.category.tableTyp.FieldByName(table.category.levelField)
	bsonTag := strings.Split(field.Tag.Get("bson"), ",")
	db := mgo.Collection(collection)
	db.Where(nil).Sort(bsonTag[0]).FindMany(items.Interface())
	return &Tree{items: items.Elem(), category: table.category}
}

// WithMysql 使用 gorm 驱动库
func (table *Table) WithMysql() *Tree {
	items := newItems(table.category.tableTyp)
	field, _ := table.category.tableTyp.FieldByName(table.category.levelField)
	column := func() string {
		tags := strings.Split(field.Tag.Get("gorm"), ",")
		for _, tag := range tags {
			tagInfo := strings.Split(tag, ":")
			if len(tagInfo) == 2 && strings.ToLower(tagInfo[0]) == "column" {
				return tagInfo[1]
			}
		}
		return table.category.levelField
	}()

	orm.Slave().Table(table.category.tableName()).Order(fmt.Sprintf("%s asc", column)).Find(items.Interface())
	return &Tree{items: items.Elem(), category: table.category}
}

// WithMongo 使用 go-mongo-driver 驱动库
func (table *Table) WithMongo() *Tree {
	collection := reflect.New(table.category.tableTyp).Interface().(database.Table)
	items := newItems(table.category.tableTyp)
	field, _ := table.category.tableTyp.FieldByName(table.category.levelField)
	bsonTag := strings.Split(field.Tag.Get("bson"), ",")
	mongo.Collection(collection).Where(nil).Sort(bson.M{bsonTag[0]: 1}).FindMany(items.Interface())
	return &Tree{items: items.Elem(), category: table.category}
}

// Categories 获得分类树
func (tree *Tree) Categories() Nodes {
	dataLen := tree.items.Len()
	var nodes, tmp = make(Nodes, 0), make(map[string]*Node)

	for i := 0; i < dataLen; i++ {
		item := tree.items.Index(i)
		pid, id := item.FieldByName(tree.category.PidField).String(), item.FieldByName(tree.category.IDField).String()
		tmp[id] = newNode(item, tree.category)

		if pid == "" {
			// 如果pid为空默认为顶级分类
			nodes = append(nodes, tmp[id])
		} else {
			if _, ok := tmp[pid]; ok {
				if tmp[pid].Next == nil {
					tmp[pid].Next = make(Nodes, 0)
				}
				// 如果在pos中存在父id为pid的数据,直接将这次的数据追加到pos中该组数据中的children中
				tmp[pid].Next = append(tmp[pid].Next, tmp[id])
			} else {
				// 否则默认无父分类，直接写入tree中
				nodes = append(nodes, tmp[id])
			}
		}
	}

	return nodes
}

func mustHasValue(value reflect.Value) string {
	v := value.Interface()
	switch v.(type) {
	case string:
		return v.(string)
	}
	app.Logger().WithField("log_type", "category.category").Error("only string is supported for the time being")
	return ""
}

func newNode(item reflect.Value, category *Category) *Node {
	return &Node{
		Alias:       mustHasValue(item.FieldByName(category.aliasField)),
		Name:        mustHasValue(item.FieldByName(category.nameField)),
		Pic:         mustHasValue(item.FieldByName(category.picField)),
		Badge:       mustHasValue(item.FieldByName(category.badgeField)),
		Description: mustHasValue(item.FieldByName(category.descriptionField)),
	}
}
