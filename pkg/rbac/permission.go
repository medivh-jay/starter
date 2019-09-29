package rbac

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"starter/pkg/database/mongo"
)

// Permission 权限列表
type Permission struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	Name      string             `json:"name" bson:"name" binding:"max=24" form:"name"` // 权限名称
	Path      string             `json:"path" bson:"path" form:"path"`                  // 资源定位路径
	Method    string             `json:"method" bson:"method" form:"method"`            // 请求方式
	CreatedAt int64              `json:"created_at" bson:"created_at"`
	UpdatedAt int64              `json:"updated_at" bson:"updated_at"`
}

// Permissions 一个权限列表
type Permissions []Permission

var permission Permission

// HasPermission 查询权限列表中是否包含指定的权限
func (permissions Permissions) HasPermission(id primitive.ObjectID, path, method string) bool {
	for _, permission := range permissions {
		// 超级管理员
		if permission.Method == "*" && permission.Path == "*" {
			return true
		}
		if permission.ID.Hex() == id.Hex() {
			return true
		}
	}
	return false
}

// TableName 表名
func (Permission) TableName() string { return "permissions" }

// GetPermissionsByIDs 根据权限ID获取权限
func (Permission) GetPermissionsByIDs(ids ...primitive.ObjectID) Permissions {
	var permissions Permissions
	mongo.Collection(permission).Where(bson.M{"_id": bson.M{"$in": ids}}).FindMany(&permissions)
	return permissions
}

// GetPermissionsByRequest 根据请求参数获取权限列表
//  uri 资源路径
//  method 请求方式
func (Permission) GetPermissionsByRequest(path string, method string) Permissions {
	var permissions Permissions
	mongo.Collection(permission).Where(bson.M{"path": path, "method": method}).FindMany(&permissions)
	return permissions
}
