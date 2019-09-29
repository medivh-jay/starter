package rbac

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"starter/pkg/database/mongo"
)

// Role 角色
type Role struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	Name        string             `json:"name" bson:"name" form:"name" binding:"max=12"`       // 角色名称
	Pid         string             `json:"pid" bson:"pid" form:"pid"`                           // 角色的父级角色id
	Permissions []string           `json:"permissions" bson:"permissions" form:"permissions[]"` // 权限id列表
	CreatedAt   int64              `json:"created_at" bson:"created_at"`
	UpdatedAt   int64              `json:"updated_at" bson:"updated_at"`
}

// Roles 角色列表
type Roles []Role

var role Role

// GetPermissionIDs 获取角色列表中包含的所有权限id
func (roles Roles) GetPermissionIDs() []primitive.ObjectID {
	var ids []primitive.ObjectID
	for _, role := range roles {
		for _, permission := range role.Permissions {
			id, _ := primitive.ObjectIDFromHex(permission)
			ids = append(ids, id)
		}
	}
	return ids
}

// TableName 表名
func (Role) TableName() string { return "roles" }

// GetRoles 获取权限
func (Role) GetRoles(id ...primitive.ObjectID) Roles {
	var roles Roles
	mongo.Collection(role).Where(bson.M{"_id": bson.M{"$in": id}}).FindMany(&roles)
	return roles
}
