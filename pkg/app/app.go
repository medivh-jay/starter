// 提供全局公用依赖性极低操作
//  不要尝试写入复杂操作逻辑到这里, 可能会引起令人头疼的循环调用问题
package app

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"starter/pkg/log"
)

// 数据表结构体必须实现此接口
type Table interface {
	TableName() string
}

type response struct {
	Code    int         `json:"code"`    // 状态码,这个状态码是与前端和APP约定的状态码,非HTTP状态码
	Data    interface{} `json:"data"`    // 返回数据
	Message string      `json:"message"` // 自定义返回的消息内容
}

// 在调用了这个方法之后,还是需要 return 的
func (rsp *response) End(c *gin.Context, httpStatus ...int) {
	status := http.StatusOK
	if len(httpStatus) > 0 {
		status = httpStatus[0]
	}
	rsp.Message = Translate(c.DefaultQuery("lang", "zh-cn"), rsp.Message)
	c.JSON(status, rsp)
}

// 接口返回统一使用这个
//  code 服务端与客户端和web端约定的自定义状态码
//  data 具体的返回数据
//  message 可不传,自定义消息
func NewResponse(code int, data interface{}, message ...string) *response {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return &response{Code: code, Data: data, Message: msg}
}

// Root
//  返回程序运行时的运行目录
func Root() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

// Name
//  返回程序名称
func Name() string {
	stat, _ := os.Stat(os.Args[0])
	return stat.Name()
}

func Lang(ctx *gin.Context) string {
	return ctx.DefaultQuery("lang", "zh-cn")
}

// 获取运行模式
func Mode() string {
	return gin.Mode()
}

// md5
func Md5(text string) string {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

// 获取log对象, 避免全包各处导入造成的循环调用
func Logger() *logrus.Logger {
	return log.Logger
}
