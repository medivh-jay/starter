package entities

import "gopkg.in/mgo.v2/bson"

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

func (Mgo) TableName() string {
	return "mgo"
}
