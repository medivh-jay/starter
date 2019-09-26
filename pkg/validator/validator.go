// Package validator 主要是将 gin 的默认表单验证模块替换为 validator.v9
package validator

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	translations "gopkg.in/go-playground/validator.v9/translations/en"
	"reflect"
	"starter/pkg/app"
	"sync"
)

// Validator 验证器
type Validator struct {
	once     sync.Once
	validate *validator.Validate
}

var (
	_                   binding.StructValidator = &Validator{}
	validatorMessages   map[string]map[string]string
	langMapping         = map[string]string{"zh-cn": "zh", "en-us": "en"}
	_                   = app.Config().Bind("validator-messages", "", &validatorMessages)
	enUs                = en.New()
	zhCn                = zh.New()
	universalTranslator = ut.New(enUs, enUs, zhCn)

	defaultTrans, _ = universalTranslator.GetTranslator(zhCn.Locale())
	enTrans, _      = universalTranslator.GetTranslator(enUs.Locale())
)

// ValidateStruct 验证结构体
func (v *Validator) ValidateStruct(obj interface{}) error {
	if kindOfData(obj) == reflect.Struct {
		v.lazyInit()
		if err := v.validate.Struct(obj); err != nil {
			return error(err)
		}
	}

	return nil
}

// Engine 获取验证器
func (v *Validator) Engine() interface{} {
	v.lazyInit()
	return v.validate
}

// lazyInit 延迟初始化
func (v *Validator) lazyInit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")

		// 获取form tag
		v.validate.RegisterTagNameFunc(func(field reflect.StructField) string {
			name := field.Tag.Get("form")
			if name != "" {
				return name
			}
			return field.Name
		})

		_ = translations.RegisterDefaultTranslations(v.validate, enTrans)

		for tag, languages := range validatorMessages {
			var trueTag, messages = tag, languages
			registerTranslation(v.validate, trueTag, messages)
		}
	})
}

func registerTranslation(validate *validator.Validate, tag string, languages map[string]string) {
	_ = validate.RegisterTranslation(tag, defaultTrans, func(ut ut.Translator) error {
		return ut.Add(tag, languages["zh-cn"], true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tag, fe.Param(), fe.Field())
		return t
	})
	_ = validate.RegisterTranslation(tag, enTrans, func(ut ut.Translator) error {
		return ut.Add(tag, languages["en-us"], true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tag, fe.Param(), fe.Field())
		return t
	})
}

// ValidErrors 验证之后的错误信息
type ValidErrors struct {
	ErrorsInfo map[string]string
	triggered  bool
}

func (validErrors *ValidErrors) add(key, value string) {
	validErrors.ErrorsInfo[key] = value
	validErrors.triggered = true
}

// IsValid 是否验证成功
func (validErrors *ValidErrors) IsValid() bool {
	return !validErrors.triggered
}

func newValidErrors() *ValidErrors {
	return &ValidErrors{
		triggered:  false,
		ErrorsInfo: make(map[string]string),
	}
}

// Bind 自定义错误信息, 如果没有匹配需要在 configs/validator-messages.toml 中添加对应处理数据
func Bind(c *gin.Context, param interface{}) *ValidErrors {
	lang := c.DefaultQuery("lang", "zh-cn")
	trans, _ := universalTranslator.GetTranslator(langMapping[lang])
	err := c.ShouldBind(param)
	var validErrors = newValidErrors()
	if err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if ok {
			for _, value := range errs {
				validErrors.add(value.Field(), value.Translate(trans))
			}
		}
	}
	return validErrors
}

func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}
