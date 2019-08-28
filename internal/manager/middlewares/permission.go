package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"starter/internal/entities"
	"starter/pkg/app"
	"starter/pkg/middlewares"
	"starter/pkg/permission"
)

func CheckPermission(context *gin.Context) {
	staff, exists := context.Get(middlewares.AuthKey)
	if !exists {
		app.NewResponse(app.PermissionDenied, nil, app.PermissionDeniedMessage).End(context, http.StatusForbidden)
		return
	}

	if permission.HasPermission(staff.(entities.Staff).Id.Hex(), context) {
		context.Next()
		return
	}

	app.NewResponse(app.PermissionDenied, nil, app.PermissionDeniedMessage).End(context, http.StatusForbidden)
	context.Abort()
	return
}
