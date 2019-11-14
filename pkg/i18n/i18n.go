// Package i18n i18n 国际化
//  example:
//  bundle := i18n.NewBundle(language.Chinese).LoadFiles("./locales", toml.Unmarshal)
//	log.Println(bundle.NewPrinter(language.English).Translate("Hello", i18n.Data{"name": "medivh", "count": 156}, 156))
//  # return 你好世界
package i18n

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

// LangKey 默认从URL取语言值得键名
var LangKey = "lang"

// Data 自定义的需要被模板解析的信息, 在不需要传入结构体时可以使用此类型
type Data map[string]interface{}

// Message 具体翻译模板的内容
type Message struct {
	ID    string             // message key
	Other *template.Template // message template
	Zero  *template.Template // message template
	One   *template.Template // message template
	Two   *template.Template // message template
	Few   *template.Template // message template
	Many  *template.Template // message template
}

// Bundle 国际化对外操作结构体
type Bundle struct {
	mu         sync.Mutex
	defaultTag language.Tag
	messages   map[language.Tag]map[string]*Message
}

// Messages 所有的模板信息
type Messages map[language.Tag]map[string]*Message

// Unmarshal 模板文本decode方法
type Unmarshal func(data []byte, v interface{}) error

type few struct {
	min, max int
}

// Printer 一个新的message translate 对象, 一般不需要调用他
type Printer struct {
	few        few // In this range belongs to "few"
	many       int // Greater than or equal to this value is many
	messages   Messages
	acceptTags []language.Tag
	defaultTag language.Tag
}

// NewBundle 得到一个写的国际化实例
func NewBundle(tag language.Tag) *Bundle {
	return &Bundle{
		defaultTag: tag,
		messages:   make(Messages),
	}
}

// SetMessage add message
func (bundle *Bundle) SetMessage(tag language.Tag, key string, message map[string]string) {
	bundle.mu.Lock()
	defer bundle.mu.Unlock()
	bundle.messages[tag][key] = &Message{
		ID:    key,
		Other: createMessageTemplate(key, message["other"]),
		Zero:  createMessageTemplate(key, message["zero"]),
		One:   createMessageTemplate(key, message["one"]),
		Two:   createMessageTemplate(key, message["two"]),
		Few:   createMessageTemplate(key, message["few"]),
		Many:  createMessageTemplate(key, message["many"]),
	}
}

// LoadFiles walk file dir and load messages
// language file like
//  + path
//  | -- zh.toml
//  | -- en.toml
func (bundle *Bundle) LoadFiles(path string, unmarshaler Unmarshal) *Bundle {
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		fileInfo := strings.Split(info.Name(), ".")
		lang := fileInfo[0]
		content, _ := ioutil.ReadFile(path)
		var data map[string]map[string]string
		err = unmarshaler(content, &data)
		if err != nil {
			return err
		}
		tag := language.MustParse(lang)
		bundle.messages[tag] = make(map[string]*Message)
		for messageID, content := range data {
			bundle.SetMessage(tag, messageID, content)
		}
		return nil
	})
	return bundle
}

func createMessageTemplate(messageID, text string) *template.Template {
	if text == "" {
		return nil
	}

	t, err := template.New(messageID).Parse(text)
	if err != nil {
		return nil
	}

	return t
}

// GetAcceptLanguages 根据 URL 传参和 Accept-Language 获取查询语言
//  按权重对接受语言排序
//  如果URL存在正确language将会是最高权重
func GetAcceptLanguages(ctx *gin.Context) []language.Tag {
	var returnTags = make([]language.Tag, 0)
	tags, _, _ := language.ParseAcceptLanguage(ctx.GetHeader("Accept-Language"))
	queryLang, exists := ctx.GetQuery(LangKey)
	if exists {
		tag, err := language.Parse(queryLang)
		if err == nil {
			returnTags = append(returnTags, tag)
		}
	}
	returnTags = append(returnTags, tags...)
	return returnTags
}

// NewPrinter 根据传入语言tag获得具体翻译组件
func (bundle *Bundle) NewPrinter(tags ...language.Tag) *Printer {
	return &Printer{acceptTags: tags, defaultTag: bundle.defaultTag, messages: bundle.messages}
}

// SetFewRule 自定义 few 信息模板的few规则, 在min-max范围内将使用few模板
func (p *Printer) SetFewRule(min, max int) *Printer {
	p.few = few{min, max}
	return p
}

// SetManyRule 自定义 Many信息模板规则，大于等于 min 将使用 many 模板
func (p *Printer) SetManyRule(min int) *Printer {
	p.many = min
	return p
}

func (p *Printer) getAcceptTag() language.Tag {
	for _, tag := range p.acceptTags {
		if _, ok := p.messages[tag]; ok {
			return tag
		}
	}
	return p.defaultTag
}

// Translate 根据传入模板ID进行翻译, data can be nil
func (p *Printer) Translate(key string, data interface{}, plurals ...int) string {
	var rs bytes.Buffer
	var err error
	tag := p.getAcceptTag()

	messages := p.messages[tag]
	message, ok := messages[key]
	if !ok {
		message = p.messages[p.defaultTag][key]
	}

	if message == nil {
		return "translate template not found"
	}

	var msg *template.Template
	if len(plurals) == 0 {
		msg = message.Other
	} else {
		msg = p.template(plurals[0], message)
	}

	if msg == nil {
		return ""
	}

	err = msg.Execute(&rs, data)
	if err != nil {
		return fmt.Sprintf("translate message error: %v", err)
	}

	content, _ := ioutil.ReadAll(&rs)
	return string(content)
}

func (p *Printer) template(plural int, message *Message) *template.Template {
	var t *template.Template
	if plural >= 0 && plural <= 2 {
		t = [3]*template.Template{message.Zero, message.One, message.Two}[plural]
	}

	if t == nil {
		if p.few.min > 0 && p.few.max > 0 && p.few.min <= plural && p.few.max >= plural {
			t = message.Few
		}

		if p.many > 2 && p.many <= plural {
			t = message.Many
		}
	}

	if t == nil {
		t = message.Other
	}

	return t
}
