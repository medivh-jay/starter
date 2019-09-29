// Package rbac 权限控制
package rbac

import (
	"github.com/gin-gonic/gin"
	"starter/pkg/database/managers"
)

// Inject 对managers注入该权限验证模块
func Inject(router gin.IRoutes) {
	managers.New().Register(Role{}, managers.Mongo).Register(Permission{}, managers.Mongo).Register(Binding{}, managers.Mongo).Start(router)
}

// HasPermission 判断指定用户是否有当前访问资源的权限
func HasPermission(userID string, ctx *gin.Context) bool {
	bindings := binding.GetRoleIDs(userID)
	if len(bindings) == 0 {
		return false
	}
	permissionIDs := role.GetRoles(bindings...).GetPermissionIDs()
	if len(permissionIDs) == 0 {
		return false
	}
	permissions := permission.GetPermissionsByIDs(permissionIDs...)

	for _, id := range permissionIDs {
		if permissions.HasPermission(id, ctx.Request.URL.Path, ctx.Request.Method) {
			return true
		}
	}

	return false
}
