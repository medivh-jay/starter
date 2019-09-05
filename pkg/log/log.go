package log

import (
	"fmt"
	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/medivh-jay/lfshook"
	"github.com/olivere/elastic/v7"
	esconfig "github.com/olivere/elastic/v7/config"
	"github.com/sirupsen/logrus"
	"os"
	"starter/pkg/app"
	"time"
)

var (
	Terminal = logrus.New()
	Logger   = Terminal
	config   logConfig
)

type logConfig struct {
	Es    bool   `toml:"es"`
	Index string `toml:"index"`
	Dir   string `toml:"dir"`
}

type elasticConfig struct {
	URL         string `toml:"url"`
	Index       string `toml:"index"`
	Username    string `toml:"username"`
	Password    string `toml:"password"`
	Shards      int    `toml:"shards"`
	Replicas    int    `toml:"replicas"`
	Sniff       bool   `toml:"sniff"`
	HealthCheck bool   `toml:"health"`
	InfoLog     string `toml:"info_log"`
	ErrorLog    string `toml:"error_log"`
	TraceLog    string `toml:"trace_log"`
}

func startEsLog() {
	var elasticConfig elasticConfig
	_ = app.Config().Bind("application", "elasticsearch", &elasticConfig)
	elasticConfig.Index = config.Index
	var retryTimes = 0

retry:
	client, err := elastic.NewClientFromConfig(&esconfig.Config{
		URL:         elasticConfig.URL,
		Index:       elasticConfig.Index,
		Username:    elasticConfig.Username,
		Password:    elasticConfig.Password,
		Shards:      elasticConfig.Shards,
		Replicas:    elasticConfig.Replicas,
		Sniff:       &elasticConfig.Sniff,
		Healthcheck: &elasticConfig.HealthCheck,
		Infolog:     elasticConfig.InfoLog,
		Errorlog:    elasticConfig.ErrorLog,
		Tracelog:    elasticConfig.TraceLog,
	})
	if err != nil {
		retryTimes++
		Logger.Info("try to reconnect: elasticsearch")
		if retryTimes < 3 {
			time.Sleep(time.Duration(retryTimes) * 5 * time.Second)
			goto retry
		}
		Logger.Fatalln(err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "development"
	}

	hook, err := NewBulkProcessorElasticHookWithFunc(client, hostname, logrus.DebugLevel, func() string {
		return fmt.Sprintf("%s-%s", config.Index, time.Now().Format("2006-01-02"))
	})
	if err != nil {
		Terminal.Debug(err)
	}
	es := logrus.New()
	es.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	es.Hooks.Add(hook)
	// 这个操作太影响性能,release不启用
	es.SetReportCaller(gin.Mode() != gin.ReleaseMode)
	es.SetNoLock()
	Logger = es
}

func Start() {
	_ = app.Config().Bind("application", "log", &config)
	Terminal.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	// 这个操作太影响性能,release不启用
	Terminal.SetReportCaller(gin.Mode() != gin.ReleaseMode)

	if config.Es {
		startEsLog()
	} else {
		infoWriter, _ := rotatelogs.New(config.Dir+"/info_%Y%m%d.log",
			rotatelogs.WithMaxAge(time.Duration(86400)*time.Second), rotatelogs.WithRotationTime(time.Duration(86400)*time.Second))
		errorWriter, _ := rotatelogs.New(config.Dir+"/error_%Y%m%d.log",
			rotatelogs.WithMaxAge(time.Duration(86400)*time.Second), rotatelogs.WithRotationTime(time.Duration(86400)*time.Second))
		debugWriter, _ := rotatelogs.New(config.Dir+"/debug_%Y%m%d.log",
			rotatelogs.WithMaxAge(time.Duration(86400)*time.Second), rotatelogs.WithRotationTime(time.Duration(86400)*time.Second))
		warnWriter, _ := rotatelogs.New(config.Dir+"/warn_%Y%m%d.log",
			rotatelogs.WithMaxAge(time.Duration(86400)*time.Second), rotatelogs.WithRotationTime(time.Duration(86400)*time.Second))

		Terminal.AddHook(lfshook.NewHook(
			lfshook.WriterMap{
				logrus.InfoLevel:  infoWriter,
				logrus.ErrorLevel: errorWriter,
				logrus.DebugLevel: debugWriter,
				logrus.WarnLevel:  warnWriter,
			}, &logrus.JSONFormatter{}))
	}

	app.Register("logger", Logger)
}
