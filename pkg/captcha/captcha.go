package captcha

import (
	"github.com/go-redis/redis"
	"github.com/mojocn/base64Captcha"
	"starter/pkg/app"
	"sync"
	"time"
)

// 重新对该变量赋值实现自定义验证码操作
var Config = base64Captcha.ConfigDigit{
	Height:     80,
	Width:      240,
	MaxSkew:    0.01,
	DotCount:   80,
	CaptchaLen: 6,
}

var (
	store = new(customizeRdsStore)
	conf  config
)

type (
	customizeRdsStore struct {
		redisClient *redis.Client
		sync.Once
	}

	config struct {
		Addr         string `toml:"addr"`
		Password     string `toml:"password"`
		Db           int    `toml:"db"`
		PoolSize     int    `toml:"pool_size"`
		MinIdleConns int    `toml:"min_idle_conns"`
	}
)

type captcha struct {
	CaptchaId string
	captcha   base64Captcha.CaptchaInterface
}

//  idKey 自定义验证码标识id
func New(idKey string) *captcha {
	store.lazyLoad()
	captchaId, capt := base64Captcha.GenerateCaptcha(idKey, Config)
	var captcha = new(captcha)
	captcha.CaptchaId = captchaId
	captcha.captcha = capt
	return captcha
}

// 根据自定义配置生成验证码
//  config 自定义配置, 可从这里追到源码查看配置示例
func NewWithConfig(idKey string, config interface{}) *captcha {
	store.lazyLoad()
	captchaId, capt := base64Captcha.GenerateCaptcha(idKey, config)
	var captcha = new(captcha)
	captcha.CaptchaId = captchaId
	captcha.captcha = capt
	return captcha
}

func (captcha *captcha) ToBase64EncodeString() string {
	return base64Captcha.CaptchaWriteToBase64Encoding(captcha.captcha)
}

// 获取生成的验证码具体值
func (captcha *captcha) GetVerifyValue() string {
	switch captcha.captcha.(type) {
	case *base64Captcha.Audio:
		return captcha.captcha.(*base64Captcha.Audio).VerifyValue
	case *base64Captcha.CaptchaImageDigit:
		return captcha.captcha.(*base64Captcha.CaptchaImageDigit).VerifyValue
	case *base64Captcha.CaptchaImageChar:
		return captcha.captcha.(*base64Captcha.CaptchaImageChar).VerifyValue
	}
	return ""
}

func Verify(id, value string) bool {
	return base64Captcha.VerifyCaptchaAndIsClear(id, value, true)
}

func (s *customizeRdsStore) lazyLoad() {
	s.Once.Do(func() {
		_ = app.Config().Bind("application", "captcha", &conf)
		store.redisClient = redis.NewClient(&redis.Options{
			Addr:         conf.Addr,
			Password:     conf.Password,
			DB:           conf.Db,
			PoolSize:     conf.PoolSize,
			MinIdleConns: conf.MinIdleConns,
		})
		base64Captcha.SetCustomStore(s)
	})
}

func (s *customizeRdsStore) Set(id string, value string) {
	s.lazyLoad()
	err := s.redisClient.Set(id, value, time.Minute*10).Err()
	if err != nil {
		app.Logger().WithField("log_type", "pkg.captcha.captcha").Error(err)
	}
}

func (s *customizeRdsStore) Get(id string, clear bool) string {
	s.lazyLoad()
	val, err := s.redisClient.Get(id).Result()
	if err != nil {
		app.Logger().WithField("log_type", "pkg.captcha.captcha").Error(err)
		return ""
	}
	if clear {
		err := s.redisClient.Del(id).Err()
		if err != nil {
			app.Logger().WithField("log_type", "pkg.captcha.captcha").Error(err)
			return ""
		}
	}
	return val
}
