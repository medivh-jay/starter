package server

import (
	"crypto/tls"
	"fmt"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"starter/pkg/app"
	"starter/pkg/database/mgo"
	"starter/pkg/database/mongo"
	"starter/pkg/database/orm"
	"starter/pkg/elastic"
	"starter/pkg/log"
	"starter/pkg/redis"
	"starter/pkg/validator"
	"strings"
	"time"
)

type (
	application struct {
		Name          string `toml:"name"`
		Domain        string `toml:"domain"`
		Addr          string `toml:"addr"`
		PasswordToken string `toml:"password_token"`
		JwtToken      string `toml:"jwt-token"`
		CertFile      string `toml:"cert_file"`
		KeyFile       string `toml:"key_file"`
	}

	applications map[string]application
)

var (
	pidFile     = fmt.Sprintf("./%s.pid", app.Name())
	Mode        string
	After       func(engine *gin.Engine) // 在各项服务启动之后会执行的操作
	swagHandler gin.HandlerFunc
	engine      = gin.New()
	Modes       applications
)

func certInfo(module string) (string, string) {
	return Modes[module].CertFile, Modes[module].KeyFile
}

// 启动各项服务
func start() {
	log.Start()
	orm.Start()
	mongo.Start()
	mgo.Start()
	redis.Start()
	elastic.Start()

	// 加载应用配置
	_ = app.Config().Bind("application", "application", &Modes)

	// 将 gin 的验证器替换为 v9 版本
	binding.Validator = new(validator.Validator)
}

// 启动服务
func Run(service func(engine *gin.Engine)) {
	lock := createPid()
	defer lock.UnLock()

	start()
	app.Logger().WithField("log_type", "pkg.server.server").Info("server started at:", time.Now().String())
	engine.Use(logger, recovery)
	service(engine)

	if swagHandler != nil && gin.Mode() != gin.ReleaseMode {
		engine.GET("/doc/*any", swagHandler)
	}

	if After != nil {
		After(engine)
	}

	_ = gracehttp.ServeWithOptions([]*http.Server{createServer(engine)}, gracehttp.PreStartProcess(func() error {
		app.Logger().WithField("log_type", "pkg.server.server").Println("unlock pid")
		lock.UnLock()
		return nil
	}))
}

func createServer(engine *gin.Engine) *http.Server {
	server := &http.Server{
		Addr:    Modes[Mode].Addr,
		Handler: engine,
	}

	if certFile, certKey := certInfo(Mode); certFile != "" && certKey != "" {
		server.TLSConfig = &tls.Config{}
		f, _ := tls.LoadX509KeyPair(certFile, certKey)
		server.TLSConfig.Certificates = []tls.Certificate{f}
	}

	return server
}

// 对启动进程记录进程id
func createPid() *app.Flock {
	pidLock, pidLockErr := app.FLock(pidFile)
	if pidLockErr != nil {
		app.Logger().WithField("log_type", "pkg.server.server").Fatalln("createPid: lock error", pidLockErr)
	}

	err := pidLock.WriteTo(fmt.Sprintf(`%d`, os.Getpid()))
	if err != nil {
		app.Logger().WithField("log_type", "pkg.server.server").Fatalln("write error: ", err)
	}
	return pidLock
}

// 自定义的GIN日志处理中间件
// 可能在终端输出时看起来比较的难看
func logger(ctx *gin.Context) {
	start := time.Now()
	path := ctx.Request.URL.Path
	raw := ctx.Request.URL.RawQuery

	ctx.Next()

	if raw != "" {
		path = path + "?" + raw
	}

	var params = make(logrus.Fields)
	params["latency"] = time.Now().Sub(start)
	params["path"] = path
	params["method"] = ctx.Request.Method
	params["status"] = ctx.Writer.Status()
	params["body_size"] = ctx.Writer.Size()
	params["client_ip"] = ctx.ClientIP()
	params["user_agent"] = ctx.Request.UserAgent()
	params["log_type"] = "pkg.server.server"
	if !gin.IsDebugging() {
		// 在正式环境将上下文传递的变量也进行记录, 方便分析
		params["keys"] = ctx.Keys
	}
	app.Logger().WithFields(params).Info("request success, status is ", ctx.Writer.Status(), ", client ip is ", ctx.ClientIP())
}

func recovery(ctx *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			var brokenPipe bool
			if ne, ok := err.(*net.OpError); ok {
				if se, ok := ne.Err.(*os.SyscallError); ok {
					if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
						brokenPipe = true
					}
				}
			}
			stack := app.Stack(3)
			httpRequest, _ := httputil.DumpRequest(ctx.Request, false)

			if gin.IsDebugging() {
				app.Logger().WithField("log_type", "pkg.server.server").Error(string(httpRequest))
				for i := 0; i < len(stack); i++ {
					app.Logger().
						WithField("log_type", "pkg.server.server").
						WithFields(logrus.Fields{"func": stack[i]["func"], "source": stack[i]["source"]}).
						Error(fmt.Sprintf("%s:%d", stack[i]["file"], stack[i]["line"]))
				}
			} else {
				app.Logger().WithField("log_type", "pkg.server.server").
					WithField("stack", stack).WithField("request", string(httpRequest)).
					Error()
			}

			if brokenPipe {
				_ = ctx.Error(err.(error))
				ctx.Abort()
			} else {
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}
	}()
	ctx.Next()
}
