package log

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/medivh-jay/lfshook"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"os"
	"starter/pkg/config"
	"time"
)

var (
	Terminal = logrus.New()
	Logger   = Terminal
)

func startEsLog() {
	conf := config.Config.Logs
	elasticConfig := config.ElasticSearchConfig()
	elasticConfig.Index = conf.Index
	client, err := elastic.NewClientFromConfig(elasticConfig)
	if err != nil {
		Logger.Fatalln(err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "development"
	}

	hook, err := NewAsyncElasticHook(client, hostname, logrus.DebugLevel, conf.Index)
	if err != nil {
		Terminal.Panic(err)
	}
	es := logrus.New()
	es.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	es.Hooks.Add(hook)
	// 这个操作太影响性能
	es.SetReportCaller(false)
	es.SetNoLock()
	Logger = es
}

func Start() {
	Terminal.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	// 这个操作太影响性能
	Terminal.SetReportCaller(false)
	conf := config.Config.Logs

	if conf.Es {
		startEsLog()
	} else {
		infoWriter, _ := rotatelogs.New(conf.Dir+"/info_%Y%m%d.log",
			rotatelogs.WithMaxAge(time.Duration(86400)*time.Second), rotatelogs.WithRotationTime(time.Duration(86400)*time.Second))
		errorWriter, _ := rotatelogs.New(conf.Dir+"/error_%Y%m%d.log",
			rotatelogs.WithMaxAge(time.Duration(86400)*time.Second), rotatelogs.WithRotationTime(time.Duration(86400)*time.Second))
		debugWriter, _ := rotatelogs.New(conf.Dir+"/debug_%Y%m%d.log",
			rotatelogs.WithMaxAge(time.Duration(86400)*time.Second), rotatelogs.WithRotationTime(time.Duration(86400)*time.Second))
		warnWriter, _ := rotatelogs.New(conf.Dir+"/warn_%Y%m%d.log",
			rotatelogs.WithMaxAge(time.Duration(86400)*time.Second), rotatelogs.WithRotationTime(time.Duration(86400)*time.Second))

		Terminal.AddHook(lfshook.NewHook(
			lfshook.WriterMap{
				logrus.InfoLevel:  infoWriter,
				logrus.ErrorLevel: errorWriter,
				logrus.DebugLevel: debugWriter,
				logrus.WarnLevel:  warnWriter,
			}, &logrus.JSONFormatter{}))
	}
}
