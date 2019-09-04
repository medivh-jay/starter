package app

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// 配置
type configuration struct {
	configs map[string]map[string]interface{}
	once    sync.Once
}

var config = new(configuration).singleLoad()

func Config() *configuration {
	return config
}

func (conf *configuration) copy(node string, value map[string]interface{}) {
	for key, val := range value {
		if conf.configs[node] == nil {
			conf.configs[node] = make(map[string]interface{})
		}
		conf.configs[node][key] = val
	}
}

func (conf *configuration) walk(path string, info os.FileInfo, err error) error {
	if err == nil {
		if !info.IsDir() {
			var err error
			var config map[string]interface{}
			_, err = toml.DecodeFile(path, &config)
			if err != nil {
				log.Fatalln(err)
			}
			relPath, _ := filepath.Rel(Root()+"/configs/"+gin.Mode(), strings.TrimSuffix(path, ".toml"))
			node := strings.ReplaceAll(relPath, "/", ".")
			conf.copy(node, config)
		} else {
			return filepath.Walk(info.Name(), conf.walk)
		}
	}
	return nil
}

func (conf *configuration) singleLoad() *configuration {
	conf.once.Do(func() {
		conf.configs = make(map[string]map[string]interface{})
		path, _ := filepath.Abs(fmt.Sprintf("./configs/%s/", gin.Mode()))
		_ = filepath.Walk(path, conf.walk)
	})

	return conf
}

func (conf *configuration) Bind(node, key string, obj interface{}) error {
	nodeVal, ok := conf.configs[node]
	if !ok {
		return nil
	}

	var objVal interface{}

	if key != "" {
		objVal, ok = nodeVal[key]
		if !ok {
			return nil
		}
	} else {
		objVal = nodeVal
	}

	return conf.assignment(objVal, obj)
}

func (conf *configuration) assignment(val, obj interface{}) error {
	data, _ := jsoniter.Marshal(val)
	return jsoniter.Unmarshal(data, obj)
}
