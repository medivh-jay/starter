package entities

import (
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"starter/pkg/database/mongo"
	"starter/pkg/middlewares"
	"starter/pkg/password"
	"time"
)

// Staff mongo 表, 示例
// 也包含了实现 jwt 接口的示例
type Staff struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id"`
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

// PreOperation managers 包前置操作, 密码不为空时对密码进行hash操作
func (staff *Staff) PreOperation() {
	if staff.Password != "" {
		staff.Password = password.Hash(staff.Password)
	}
}

// Logged 登陆成功的操作
func (staff *Staff) Logged(platform string) {
	if staff.LoginAt == nil {
		staff.LoginAt = make(map[string]int64)
	}
	staff.LoginAt[platform] = time.Now().Unix()
	mongo.Collection(staff).UpdateOne(staff)
}

// TableName 获取表名
func (Staff) TableName() string {
	return "staffs"
}

// GetTopic 获取id
func (staff Staff) GetTopic() interface{} {
	return staff.ID
}

// FindByTopic 根据id获取用户信息
func (Staff) FindByTopic(topic interface{}) middlewares.AuthInterface {
	var id primitive.ObjectID
	var err error
	var staff Staff
	if jwtID, ok := topic.(string); ok {
		id, err = primitive.ObjectIDFromHex(jwtID)
		if err != nil {
			return staff
		}
	} else {
		if jwtID, ok := topic.(primitive.ObjectID); ok {
			id = jwtID
		}
	}

	_ = mongo.Collection(staff).Where(bson.M{"_id": id}).FindOne(&staff)
	return staff
}

// GetCheckData 得到会被 jwt 加密的信息
func (staff Staff) GetCheckData() string {
	checkData, _ := jsoniter.MarshalToString(staff.LoginAt)
	return checkData
}

// Check 检验信息是否正确
func (staff Staff) Check(ctx *gin.Context, checkData string) bool {
	var loginAt map[string]int64
	_ = jsoniter.UnmarshalFromString(checkData, &loginAt)
	return staff.LoginAt[ctx.DefaultQuery("platform", "web")] == loginAt[ctx.DefaultQuery("platform", "web")]
}

// ExpiredAt token 过期时间
func (Staff) ExpiredAt() int64 {
	return time.Now().Add(86400 * time.Second).Unix()
}
