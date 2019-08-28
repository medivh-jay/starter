package entities

import (
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"starter/pkg/middlewares"
	"starter/pkg/mongo"
	"starter/pkg/password"
	"time"
)

// mongo 表, 示例
type Staff struct {
	Id            primitive.ObjectID `json:"_id" bson:"_id"`
	Username      string             `json:"username" bson:"username" form:"username" binding:"max=12"`
	Password      string             `json:"-" bson:"password" form:"password"`
	Avatar        string             `json:"avatar" bson:"avatar" form:"avatar"`
	RelationUsers []string           `json:"relation_users" bson:"relation_users" form:"relation_users[]"`
	Pictures      []string           `json:"pictures" bson:"pictures" form:"pictures[]"`
	LoginAt       map[string]int64   `json:"login_at" bson:"login_at"`
	Description   string             `json:"description" bson:"description" form:"description"`
	Status        int                `json:"status" bson:"status" form:"status"`
	CreatedAt     int64              `json:"created_at" bson:"created_at"`
	UpdatedAt     int64              `json:"updated_at" bson:"updated_at"`
}

func (staff *Staff) PreOperation() {
	if staff.Password != "" {
		staff.Password = password.Hash(staff.Password)
	}
}

func (staff *Staff) Logged(platform string) {
	if staff.LoginAt == nil {
		staff.LoginAt = make(map[string]int64)
	}
	staff.LoginAt[platform] = time.Now().Unix()
	mongo.Collection(staff).UpdateOne(staff)
}

func (Staff) TableName() string {
	return "staffs"
}

func (staff Staff) GetTopic() interface{} {
	return staff.Id
}

func (Staff) FindByTopic(topic interface{}) middlewares.AuthInterface {
	var id primitive.ObjectID
	var err error
	var staff Staff
	if jwtId, ok := topic.(string); ok {
		id, err = primitive.ObjectIDFromHex(jwtId)
		if err != nil {
			return staff
		}
	} else {
		if jwtId, ok := topic.(primitive.ObjectID); ok {
			id = jwtId
		}
	}

	_ = mongo.Collection(staff).Where(bson.M{"_id": id}).FindOne(&staff)
	return staff
}

func (staff Staff) GetCheckData() string {
	checkData, _ := jsoniter.MarshalToString(staff.LoginAt)
	return checkData
}
func (staff Staff) Check(ctx *gin.Context, checkData string) bool {
	var loginAt map[string]int64
	_ = jsoniter.UnmarshalFromString(checkData, &loginAt)
	return staff.LoginAt[ctx.DefaultQuery("platform", "web")] == loginAt[ctx.DefaultQuery("platform", "web")]
}

func (Staff) ExpiredAt() int64 {
	return time.Now().Add(86400 * time.Second).Unix()
}
