package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"starter/internal/entities"
	"starter/pkg/app"
	"starter/pkg/middlewares"
	"starter/pkg/rbac"
)

// CheckPermission 验证用户权限, 这是自定义的结构体，这里只是示例
func CheckPermission(context *gin.Context) {
	staff, exists := context.Get(middlewares.AuthKey)
	if !exists {
		app.NewResponse(app.PermissionDenied, nil, app.PermissionDeniedMessage).End(context, http.StatusForbidden)
		return
	}

	if rbac.HasPermission(staff.(entities.Staff).ID.Hex(), context) {
		context.Next()
		return
	}

	app.NewResponse(app.PermissionDenied, nil, app.PermissionDeniedMessage).End(context, http.StatusForbidden)
	context.Abort()
	return
}
