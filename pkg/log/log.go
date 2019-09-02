package log

import (
	"github.com/medivh-jay/eslogrus"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

var (
	Terminal = logrus.New()
	Logger   = Terminal
)

func StartEsLog(client *elastic.Client) {
	hook, err := eslogrus.NewAsyncElasticHook(client, "localhost", logrus.DebugLevel, "mylog")
	if err != nil {
		Terminal.Panic(err)
	}
	es := logrus.New()
	es.Hooks.Add(hook)
	Logger = es
}

func Start() {
	Terminal.SetFormatter(&logrus.TextFormatter{})
}
