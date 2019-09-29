package entities

import (
	"go.mongodb.org/mongo-driver/bson"
	"starter/pkg/database/mongo"
)

// Category 商品分类, 分类示例
type Category struct {
	ID          string `gorm:"primary_key;column:id;" bson:"_id"`
	CreatedAt   int    `gorm:"column:created_at;index:created_at" bson:"created_at"`
	UpdatedAt   int    `gorm:"column:updated_at;index:updated_at" bson:"updated_at"`
	Alias       string `gorm:"column:alias" bson:"alias"`
	Name        string `gorm:"column:name" bson:"name"`
	Pic         string `gorm:"column:pic" bson:"pic"`
	Badge       string `gorm:"column:badge" bson:"badge"`
	Description string `gorm:"column:description" bson:"description"`
	Pid         string `gorm:"column:pid" bson:"pid"`
	Level       int    `gorm:"column:level" bson:"level"`
}

// TableName 表名
func (Category) TableName() string {
	return "categories"
}

// PreOperation 通过managers新增或修改数据时, 计算该数据所属第几级分类
func (category *Category) PreOperation() {
	if category.Pid != "" {
		var parent Category
		// 这里只是使用了 mongo 作为示例, 这个结构体虽然同是写了 gorm 和 mongo 的tag, 但是实际一个表只会在一个数据库下进行操作, 所以实际开发按自己需求写
		_ = mongo.Collection(category).Where(bson.M{"_id": category.Pid}).FindOne(&parent)
		category.Level = parent.Level + 1
	} else {
		category.Level = 1
	}
}
