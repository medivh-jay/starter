package rbac

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"starter/pkg/database/mongo"
)

// Binding 用户id与角色id的关系
type Binding struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	UserID    string             `json:"user_id" bson:"user_id" form:"user_id" `
	RoleID    string             `json:"role_id" bson:"role_id" form:"role_id" `
	CreatedAt int64              `json:"created_at" bson:"created_at"`
	UpdatedAt int64              `json:"updated_at" bson:"updated_at"`
}

var binding Binding

// TableName 表名
func (Binding) TableName() string { return "binding" }

// GetRoleIDs 获取指定用户的所有角色ID
func (Binding) GetRoleIDs(userID string) []primitive.ObjectID {
	var bindings []Binding
	mongo.Collection(binding).Where(bson.M{"user_id": userID}).FindMany(&bindings)

	var roles []primitive.ObjectID
	for _, bind := range bindings {
		id, _ := primitive.ObjectIDFromHex(bind.RoleID)
		roles = append(roles, id)
	}

	return roles
}
