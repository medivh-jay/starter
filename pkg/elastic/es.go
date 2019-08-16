package elastic

import (
	"context"
	"github.com/olivere/elastic/v7"
	"log"
	"starter/pkg/config"
)

var ES *elastic.Client
var err error

func Start() {
	ES, err = elastic.NewClientFromConfig(config.ElasticSearchConfig())
	if err != nil {
		log.Fatalln(err)
	}
}

// 向es写入数据
func Insert(index string, body interface{}) *elastic.IndexResponse {
	rs, err := ES.Index().Index(index).BodyJson(body).Do(context.Background())
	if err != nil {
		log.Println("es write error: ", err)
	}

	return rs
}

func InsertString(index string, body string) *elastic.IndexResponse {
	rs, err := ES.Index().Index(index).BodyString(body).Do(context.Background())
	if err != nil {
		log.Println("es write error: ", err)
	}

	return rs
}
