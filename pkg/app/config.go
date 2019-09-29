package app

import (
	"errors"
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

// Configuration 应用配置
type Configuration struct {
	configs map[string]map[string]interface{}
	once    sync.Once
}

var (
	config = new(Configuration).singleLoad()
	json   = jsoniter.Config{EscapeHTML: true, TagKey: "toml"}.Froze()
	// ErrNodeNotExists 配置节点不存在
	ErrNodeNotExists = errors.New("node not exists")
)

// Config 得到config对象
func Config() *Configuration {
	return config
}

func (conf *Configuration) copy(node string, value map[string]interface{}) {
	for key, val := range value {
		if conf.configs[node] == nil {
			conf.configs[node] = make(map[string]interface{})
		}
		conf.configs[node][key] = val
	}
}

func (conf *Configuration) walk(path string, info os.FileInfo, err error) error {
	if err == nil {
		if !info.IsDir() {
			if !strings.HasSuffix(path, ".toml") {
				return nil
			}
			var err error
			var config map[string]interface{}
			_, err = toml.DecodeFile(path, &config)
			if err != nil {
				// 配置读失败了
				log.Fatal(err)
			}
			conf.copy(strings.TrimSuffix(info.Name(), ".toml"), config)
		} else {
			return filepath.Walk(info.Name(), conf.walk)
		}
	}
	return nil
}

func (conf *Configuration) singleLoad() *Configuration {
	conf.once.Do(func() {
		conf.configs = make(map[string]map[string]interface{})
		path, _ := filepath.Abs(fmt.Sprintf("./configs/%s/%s/", gin.Mode(), Name()))
		_ = filepath.Walk(path, conf.walk)
	})

	return conf
}

// Bind 将配置绑定到传入对象
//  node 其实是配置文件的文件名
//  key 是配置文件中的顶层key
//  具体可查看该方法的其他包的使用
func (conf *Configuration) Bind(node, key string, obj interface{}) error {
	nodeVal, ok := conf.configs[node]
	if !ok {
		return nil
	}

	var objVal interface{}

	if key != "" {
		objVal, ok = nodeVal[key]
		if !ok {
			return ErrNodeNotExists
		}
	} else {
		objVal = nodeVal
	}

	return conf.assignment(objVal, obj)
}

func (conf *Configuration) assignment(val, obj interface{}) error {
	data, _ := json.Marshal(val)
	return json.Unmarshal(data, obj)
}
