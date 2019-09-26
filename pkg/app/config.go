package app

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
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
			var err error
			var config map[string]interface{}
			_, err = toml.DecodeFile(path, &config)
			if err != nil {
				Logger().WithField("log_type", "pkg.app.config").Error(err)
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

func (conf *Configuration) singleLoad() *Configuration {
	conf.once.Do(func() {
		conf.configs = make(map[string]map[string]interface{})
		path, _ := filepath.Abs(fmt.Sprintf("./configs/%s/", gin.Mode()))
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
			return nil
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
