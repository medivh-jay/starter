// 国际化
//  example:
//  bundle := i18n.NewBundle(language.Chinese).LoadFiles("./locales", toml.Unmarshal)
//	log.Println(bundle.NewPrinter(language.English).Translate("Hello", i18n.Data{"name": "medivh", "count": 156}, 156))
//  # return 你好世界
package i18n

import (
	"bytes"
	"fmt"
	"golang.org/x/text/language"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

type Data map[string]interface{}

type Message struct {
	Id    string             // message key
	Other *template.Template // message template
	Zero  *template.Template // message template
	One   *template.Template // message template
	Two   *template.Template // message template
	Few   *template.Template // message template
	Many  *template.Template // message template
}

type Bundle struct {
	mu         sync.Mutex
	defaultTag language.Tag
	messages   map[language.Tag]map[string]*Message
}

type Messages map[language.Tag]map[string]*Message
type Unmarshal func(data []byte, v interface{}) error

type few struct {
	min, max int
}

type printer struct {
	few        few // In this range belongs to "few"
	many       int // Greater than or equal to this value is many
	messages   Messages
	acceptTag  language.Tag
	defaultTag language.Tag
}

func NewBundle(tag language.Tag) *Bundle {
	return &Bundle{
		defaultTag: tag,
		messages:   make(Messages),
	}
}

// add message
func (bundle *Bundle) SetMessage(tag language.Tag, key string, message map[string]string) {
	bundle.mu.Lock()
	defer bundle.mu.Unlock()
	bundle.messages[tag][key] = &Message{
		Id:    key,
		Other: createMessageTemplate(key, message["other"]),
		Zero:  createMessageTemplate(key, message["zero"]),
		One:   createMessageTemplate(key, message["one"]),
		Two:   createMessageTemplate(key, message["two"]),
		Few:   createMessageTemplate(key, message["few"]),
		Many:  createMessageTemplate(key, message["many"]),
	}
}

// walk file dir and load messages
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
		for messageId, content := range data {
			bundle.SetMessage(tag, messageId, content)
		}
		return nil
	})
	return bundle
}

func createMessageTemplate(messageId, text string) *template.Template {
	if text == "" {
		return nil
	}

	t, err := template.New(messageId).Parse(text)
	if err != nil {
		return nil
	}

	return t
}

func (bundle *Bundle) NewPrinter(tag language.Tag) *printer {
	return &printer{acceptTag: tag, defaultTag: bundle.defaultTag, messages: bundle.messages}
}

func (p *printer) SetFewRule(min, max int) *printer {
	p.few = few{min, max}
	return p
}

func (p *printer) SetManyRule(min int) *printer {
	p.many = min
	return p
}

// data can be nil
func (p *printer) Translate(key string, data interface{}, plurals ...int) string {
	var rs bytes.Buffer
	var err error
	messages, ok := p.messages[p.acceptTag]
	if !ok {
		messages = p.messages[p.defaultTag]
	}

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

func (p *printer) template(plural int, message *Message) *template.Template {
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
