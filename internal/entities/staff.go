package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

// mongo 表, 示例
type Staff struct {
	Id        primitive.ObjectID `json:"_id" bson:"_id"`
	Username  string             `json:"username" bson:"username" form:"username" binding:"required,max=12"`
	Password  string             `json:"-" bson:"password" form:"password"`
	LoginAt   map[string]int64   `json:"login_at" bson:"login_at"`
	CreatedAt int64              `json:"created_at" bson:"created_at"`
	UpdatedAt int64              `json:"updated_at" bson:"updated_at"`
}

// mgo 驱动
type Mgo struct {
	Id        bson.ObjectId `json:"_id" bson:"_id"`
	Username  string        `json:"username" bson:"username" form:"username" binding:"max=12"`
	Password  string        `json:"-" bson:"password" form:"password"`
	Members   []string      `json:"members" bson:"members" form:"members[]"`
	Textarea  string        `json:"textarea" bson:"textarea" form:"textarea"`
	CreatedAt int64         `json:"created_at" bson:"created_at"`
	UpdatedAt int64         `json:"updated_at" bson:"updated_at"`
}

func (Staff) TableName() string {
	return "staffs"
}

func (Mgo) TableName() string {
	return "mgo"
}
