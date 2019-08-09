//+build doc

package server

import (
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

func init() {
	swagHandler = ginSwagger.WrapHandler(swaggerFiles.Handler)
}
