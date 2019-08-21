package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// 全局允许跨域
func CORS(c *gin.Context) {
	if c.Request.Header.Get(`X-Requested-With`) != "" || c.Request.Header.Get(`Origin`) != "" {
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.Request.Header.Get(`Origin`))
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Sec-Fetch-Mode, Origin, Content-Type, Content-Length, Accept-Encoding, JWT,x-requested-with")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusOK)
		c.Abort()
		return
	} else {
		c.Next()
	}
}
