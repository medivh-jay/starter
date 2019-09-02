package elastic

import (
	"context"
	"github.com/olivere/elastic/v7"
	"starter/pkg/config"
	"starter/pkg/log"
)

var ES *elastic.Client
var err error

func Start() {
	ES, err = elastic.NewClientFromConfig(config.ElasticSearchConfig())
	if err != nil {
		log.Logger.Fatalln(err)
	}
}

// 向es写入数据
func Insert(index string, body interface{}) *elastic.IndexResponse {
	rs, err := ES.Index().Index(index).BodyJson(body).Do(context.Background())
	if err != nil {
		log.Logger.Println("es write error: ", err)
	}

	return rs
}

func InsertString(index string, body string) *elastic.IndexResponse {
	rs, err := ES.Index().Index(index).BodyString(body).Do(context.Background())
	if err != nil {
		log.Logger.Println("es write error: ", err)
	}

	return rs
}
