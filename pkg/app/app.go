// Package app 提供全局公用依赖性极低操作
//  不要尝试写入复杂操作逻辑到这里, 可能会引起令人头疼的循环调用问题
//  其他包可以调用app, 但app不要调用其他包, 需要调用的在其他包中调用 Register 将服务注册
package app

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var services = &Services{services: make(map[string]interface{})}

// Services 服务汇总
type Services struct {
	lock     sync.Mutex
	services map[string]interface{}
}

// Response HTTP返回数据结构体, 可使用这个, 也可以自定义
type Response struct {
	Code    int         `json:"code"`    // 状态码,这个状态码是与前端和APP约定的状态码,非HTTP状态码
	Data    interface{} `json:"data"`    // 返回数据
	Message string      `json:"message"` // 自定义返回的消息内容
}

// End 在调用了这个方法之后,还是需要 return 的
func (rsp *Response) End(c *gin.Context, httpStatus ...int) {
	status := http.StatusOK
	if len(httpStatus) > 0 {
		status = httpStatus[0]
	}

	rsp.Message = Translate(c.DefaultQuery("lang", "zh-cn"), rsp.Message)
	c.JSON(status, rsp)
}

// NewResponse 接口返回统一使用这个
//  code 服务端与客户端和web端约定的自定义状态码
//  data 具体的返回数据
//  message 可不传,自定义消息
func NewResponse(code int, data interface{}, message ...string) *Response {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return &Response{Code: code, Data: data, Message: msg}
}

func (service *Services) register(name string, se interface{}) {
	service.lock.Lock()
	defer service.lock.Unlock()

	service.services[name] = se
}

func (service *Services) get(name string) interface{} {
	if val, ok := service.services[name]; ok {
		return val
	}
	return nil
}

// Root 根目录
//  返回程序运行时的运行目录
func Root() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

// Name 程序名
//  返回程序名称
func Name() string {
	stat, _ := os.Stat(os.Args[0])
	return stat.Name()
}

// Lang 获取客户端传的 lang 参数
func Lang(ctx *gin.Context) string {
	return ctx.DefaultQuery("lang", "zh-cn")
}

// Mode 获取运行模式
func Mode() string {
	return gin.Mode()
}

// Md5 md5 hash
func Md5(text string) string {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

// Logger 获取日志对象
func Logger() *logrus.Logger {
	return Get("logger").(*logrus.Logger)
}

// Register 注册其他包的服务
func Register(name string, service interface{}) interface{} {
	services.register(name, service)
	return service
}

// Get 获取其他包的服务
func Get(name string) interface{} {
	return services.get(name)
}
