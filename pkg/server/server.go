package server

import (
	"crypto/tls"
	"fmt"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"log"
	"net/http"
	"os"
	"starter/pkg/app"
	"starter/pkg/config"
	"starter/pkg/elastic"
	starterLog "starter/pkg/log"
	"starter/pkg/mgo"
	"starter/pkg/mongo"
	"starter/pkg/orm"
	"starter/pkg/redis"
	"starter/pkg/validator"
)

var (
	pidFile     = fmt.Sprintf("./%s.pid", app.Name())
	Mode        string
	After       func(engine *gin.Engine) // 在各项服务启动之后会执行的操作
	swagHandler gin.HandlerFunc
)

func init() {
	config.Load()
}

// 启动各项服务
func start() {
	starterLog.Start()

	orm.Start()
	mongo.Start()
	mgo.Start()
	redis.Start()
	elastic.Start()

	// 将 gin 的验证器替换为 v9 版本
	binding.Validator = new(validator.Validator)
}

// 启动服务
func Run(engine *gin.Engine) {
	lock := createPid()
	defer lock.UnLock()

	start()
	if swagHandler != nil && gin.Mode() != gin.ReleaseMode {
		engine.GET("/doc/*any", swagHandler)
	}

	if After != nil {
		After(engine)
	}

	_ = gracehttp.ServeWithOptions([]*http.Server{createServer(engine)}, gracehttp.PreStartProcess(func() error {
		log.Println("unlock pid")
		lock.UnLock()
		return nil
	}))
}

func createServer(engine *gin.Engine) *http.Server {
	server := &http.Server{
		Addr:    config.ApplicationAddr(Mode),
		Handler: engine,
	}

	if certFile, certKey := config.ApplicationCertInfo(Mode); certFile != "" && certKey != "" {
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
		log.Fatalln("createPid: lock error", pidLockErr)
	}

	err := pidLock.WriteTo(fmt.Sprintf(`%d`, os.Getpid()))
	if err != nil {
		log.Fatalln("write error: ", err)
	}
	return pidLock
}
