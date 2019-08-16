// 配置, 如果有新增配置,在这里完善配置的结构
package config

import (
	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	elastic "github.com/olivere/elastic/v7/config"
	"log"
	"starter/pkg/app"
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
	master struct {
		Addr     string `toml:"addr"`
		Username string `toml:"username"`
		Password string `toml:"password"`
		DbName   string `toml:"dbname"`
		MaxIdle  int    `toml:"max_idle"`
		MaxOpen  int    `toml:"max_open"`
	}
	slave struct {
		Addr     string `toml:"addr"`
		Username string `toml:"username"`
		Password string `toml:"password"`
		DbName   string `toml:"dbname"`
		MaxIdle  int    `toml:"max_idle"`
		MaxOpen  int    `toml:"max_open"`
	}
	database struct {
		Master master  `toml:"master"`
		Slaves []slave `toml:"slave"`
	}
	mongo struct {
		Url             string `toml:"url"`
		Database        string `toml:"database"`
		MaxConnIdleTime int    `toml:"max_conn_idle_time"`
		MaxPoolSize     int    `toml:"max_pool_size"`
		Username        string `toml:"username"`
		Password        string `toml:"password"`
	}
	redis struct {
		Addr         string `toml:"addr"`
		Password     string `toml:"password"`
		Db           int    `toml:"db"`
		PoolSize     int    `toml:"pool_size"`
		MinIdleConns int    `toml:"min_idle_conns"`
	}
	sessions struct {
		Key          string `toml:"key"`
		Name         string `toml:"name"`
		Domain       string `toml:"domain"`
		Addr         string `toml:"addr"`
		Password     string `toml:"password"`
		Db           int    `toml:"db"`
		PoolSize     int    `toml:"pool_size"`
		MinIdleConns int    `toml:"min_idle_conns"`
	}
	elasticsearch struct {
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
)

// 配置顶级结构
type config struct {
	Application   map[string]application `toml:"application"`
	Database      database               `toml:"database"`
	Mongo         mongo                  `toml:"mongo"`
	Redis         redis                  `toml:"redis"`
	Sessions      sessions               `toml:"sessions"`
	ElasticSearch elasticsearch          `toml:"elastic"`
}

// 三方支付配置信息
type (
	payment struct {
		Alipay alipay `toml:"alipay"`
		Wechat wechat `toml:"wechat"`
	}
	alipay struct {
		AppId              string `toml:"appid"`
		AlipayRsaPublicKey string `toml:"alipay_rsa_public_key"`
		RsaPrivateKey      string `toml:"rsa_private_key"`
		NotifyUrl          string `toml:"notify_url"`
		ReturnUrl          string `toml:"return_url"`
		Product            bool   `toml:"product"`
	}
	wechat struct {
		AppId     string `toml:"appid"`
		MchId     string `toml:"mch_id"`
		NotifyUrl string `toml:"notify_url"`
		SignKey   string `toml:"sign_key"`
	}
)

var Config config
var Payment payment

func Load() {
	loadConfig()
}

func configFile() string {
	if app.Mode() == gin.ReleaseMode {
		return app.Root() + "/configs/application.toml"
	}

	return app.Root() + "/configs/development.toml"
}

func loadConfig() {
	_, err := toml.DecodeFile(configFile(), &Config)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = toml.DecodeFile(app.Root()+"/configs/payments.toml", &Payment)
	if err != nil {
		log.Fatalln(err)
	}
}

func ApplicationAddr(module string) string {
	return Config.Application[module].Addr
}

func ApplicationCertInfo(module string) (string, string) {
	return Config.Application[module].CertFile, Config.Application[module].KeyFile
}

func ElasticSearchConfig() *elastic.Config {
	return &elastic.Config{
		URL:         Config.ElasticSearch.URL,
		Index:       Config.ElasticSearch.Index,
		Username:    Config.ElasticSearch.Username,
		Password:    Config.ElasticSearch.Password,
		Shards:      Config.ElasticSearch.Shards,
		Replicas:    Config.ElasticSearch.Replicas,
		Sniff:       &Config.ElasticSearch.Sniff,
		Healthcheck: &Config.ElasticSearch.HealthCheck,
		Infolog:     Config.ElasticSearch.InfoLog,
		Errorlog:    Config.ElasticSearch.ErrorLog,
		Tracelog:    Config.ElasticSearch.TraceLog,
	}
}
