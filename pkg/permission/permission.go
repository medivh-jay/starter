package permission

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"starter/pkg/managers"
	"starter/pkg/mongo"
)

type Role struct {
	Id          primitive.ObjectID `json:"_id" bson:"_id"`
	Name        string             `json:"name" bson:"name" form:"name" binding:"required,max=12"` // 角色名称
	Pid         string             `json:"pid" bson:"pid" form:"pid"`                              // 角色的父级角色id
	Permissions []string           `json:"permissions" bson:"permissions" form:"permissions[]"`    // 权限id列表
	CreatedAt   int64              `json:"created_at" bson:"created_at"`
	UpdatedAt   int64              `json:"updated_at" bson:"updated_at"`
}

type Permission struct {
	Id         primitive.ObjectID  `json:"_id" bson:"_id"`
	Name       string              `json:"name" bson:"name" binding:"required,max=24" form:"name"` // 权限名称
	Path       string              `json:"path" bson:"path" binding:"required" form:"path"`        // 资源定位路径
	Method     string              `json:"method" bson:"method" binding:"required" form:"method"`  // 请求方式
	Filters    []map[string]string `json:"filters" bson:"filters" form:"filters[]"`                // 指定字段允许修改的值
	NotAllowed []string            `json:"not_allowed" bson:"not_allowed" form:"not_allowed[]"`    // 直接不允许操作的字段
	CreatedAt  int64               `json:"created_at" bson:"created_at"`
	UpdatedAt  int64               `json:"updated_at" bson:"updated_at"`
}

type Binding struct {
	Id        primitive.ObjectID `json:"_id" bson:"_id"`
	UserId    string             `json:"user_id" bson:"user_id" form:"user_id" binding:"required"`
	RoleId    string             `json:"role_id" bson:"role_id" form:"role_id" binding:"required"`
	CreatedAt int64              `json:"created_at" bson:"created_at"`
	UpdatedAt int64              `json:"updated_at" bson:"updated_at"`
}

func Start() {
	managers.Register("/permission/role", "roles", Role{}, managers.Mongo)
	managers.Register("/permission/permissions", "permissions", Permission{}, managers.Mongo)
	managers.Register("/permission/binding", "binding", Binding{}, managers.Mongo)
}

func GetPermissionsForUser(id string) []Permission {
	var binding Binding
	err := mongo.Collection("binding").Where(bson.M{"user_id": id}).FindOne(&binding)
	if err != nil {
		log.Println(err)
		return nil
	}

	roleId, _ := primitive.ObjectIDFromHex(binding.RoleId)
	var role Role
	err = mongo.Collection("roles").Where(bson.M{"_id": roleId}).FindOne(&role)
	if err != nil {
		log.Println(err)
		return nil
	}

	permissionIdList := make([]primitive.ObjectID, len(role.Permissions), len(role.Permissions))
	for k, v := range role.Permissions {
		permissionIdList[k], _ = primitive.ObjectIDFromHex(v)
	}

	var permissions []Permission
	mongo.Collection("permissions").Where(bson.M{"_id": bson.M{"$in": permissionIdList}}).FindMany(&permissions)

	return permissions
}

// 是否是超级管理员
func isRoot(id string, permissions ...[]Permission) bool {
	var perm []Permission
	if len(permissions) == 0 {
		perm = GetPermissionsForUser(id)
	}
	for _, permission := range perm {
		if permission.Path == "*" && permission.Method == "*" {
			return permission.Filters == nil
		}
	}
	return false
}

func HasPermission(id string, ctx *gin.Context) bool {
	permissions := GetPermissionsForUser(id)
	// root 用户拥有一切权限
	if isRoot(id, permissions) {
		return true
	}
	path, method := ctx.Request.URL.Path, ctx.Request.Method

	for _, permission := range permissions {
		// 请求方式与已有权限匹配并且请求路径也匹配,无限制字段值,拥有权限
		if permission.Method == method && permission.Path == path {
			if permission.Filters == nil {
				return true
			}

			// 对限定值和字段进行判断
			// DELETE 不用配置以下判断, 直接拥有或者不拥有DELETE权限即可
			switch method {
			case http.MethodGet:
				for _, filters := range permission.Filters {
					for key, val := range filters {
						// 如果限定值对应key传入了并且该值不等于限定值,无权限
						if ctx.Query(key) != "" && ctx.Query(key) != val {
							return false
						}
					}
				}

				for _, field := range permission.NotAllowed {
					// 如果存在不允许操作的字段但是被传入了, 无权限
					if ctx.Query(field) != "" {
						return false
					}
				}
				return true

			case http.MethodPost, http.MethodPut:
				validFields := make(map[string][]string)
				for _, filters := range permission.Filters {
					for key, val := range filters {
						validFields[key] = append(validFields[key], val)
					}
				}

				if filterValid(ctx, validFields) && notAllowValid(ctx, permission.NotAllowed) {
					return true
				}
			}
		}
	}

	return false
}

func filterValid(ctx *gin.Context, fields map[string][]string) bool {
	var result = true
	for field, values := range fields {
		val := ctx.PostForm(field)
		var valValid bool
		for _, allowValue := range values {
			valValid = valValid || val == allowValue
		}
		result = result && valValid
	}
	return result
}

func notAllowValid(ctx *gin.Context, fields []string) bool {
	for _, field := range fields {
		// 如果存在不允许操作的字段但是被传入了, 无权限
		if ctx.PostForm(field) != "" {
			return false
		}
	}
	return true
}
