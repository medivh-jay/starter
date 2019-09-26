// Package email 发送邮件
// 	email.Send(email.NewHTMLSender("这里是标题", email.ParseHHTML(app.FilePath() + "/public/email/test.html", struct {
//		Name string
//	}{Name:"medivh"}), "844627855@qq.com"))
package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"starter/pkg/app"
	"strings"
	"sync"
	"time"
)

type (
	// Email 邮件信息
	Email struct {
		username string // 发送账户
		password string // 账户密码
		host     string // smtp服务器地址
		port     int    // smtp服务器端口
		ssl      bool
	}
	// Object 发送邮件的结构数据
	Object struct {
		To      []string          // 收件人邮件地址
		Header  map[string]string // 邮件头
		Content string            // 邮件正文
	}
	// HTML 邮件如果是HTML将被转为该结构
	HTML struct {
		body []byte
	}
	config struct {
		Username string `toml:"username"`
		Password string `toml:"password"`
		Host     string `toml:"host"`
		Ssl      bool   `toml:"ssl"`
		Port     int    `toml:"port"`
	}
)

var (
	email   *Email
	auth    smtp.Auth
	senders chan *Object
	once    sync.Once
	conf    config
)

// StartEmailSender 启动邮件服务
func StartEmailSender() {
	once.Do(func() {
		_ = app.Config().Bind("application", "email", &conf)
		email = &Email{
			username: conf.Username,
			password: conf.Password,
			host:     conf.Host,

			ssl: conf.Ssl,

			port: func() int {
				if conf.Port < 1 {
					return 25
				}
				return conf.Port
			}(),
		}

		auth = smtp.PlainAuth("", email.username, email.password, email.host)
		senders = make(chan *Object, 4096)

		go task()
	})
}

// ParseHHTML 解析HTML数据
//  path 为HTML模板文件地址
//  data 为填充数据
func ParseHHTML(path string, data interface{}) *HTML {
	html := new(HTML)
	html.body = make([]byte, 0)
	parse, err := template.ParseFiles(path)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.email.email").Error(err)
		return nil
	}

	buffer := bytes.NewBuffer(html.body)
	if err = parse.Execute(buffer, data); err != nil {
		panic(err)
		return nil
	}

	html.body = buffer.Bytes()
	return html
}

// NewSender 发送邮件
func NewSender(subject, content string, to ...string) *Object {
	object := &Object{
		Content: content,
		Header:  make(map[string]string),
		To:      to,
	}
	object.writeHeader("Subject", subject).
		writeHeader("From", email.username).
		writeHeader("To", strings.Join(to, ";")).
		writeHeader("Mime-Version", "1.0").
		writeHeader("Date", time.Now().String())

	return object
}

// NewHTMLSender 发送邮件
func NewHTMLSender(subject string, html *HTML, to ...string) *Object {
	object := NewSender(subject, string(html.body), to...)
	object.writeHeader("Content-Type", "text/html;chartset=UTF-8")
	return object
}

func (object *Object) writeHeader(key, value string) *Object {
	object.Header[key] = value
	return object
}

func (object *Object) convertToBody() []byte {
	headers := ""
	for key, value := range object.Header {
		headers += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	return []byte(headers + "\r\n" + object.Content)
}

// Send 发送邮件, 邮件服务启动时现在默认是启动了一个协程, 在这里将把发送对象直接发送到协程chan中, 由协程接收然后发送
// 因为并不是所有人都会使用消息队列来处理,毕竟不是谁都是千万业务, 之后可能将会独立出来一个接口实现, 实现自我配置
func Send(sender *Object) {
	senders <- sender
}

func task() {
	for {
		select {
		case object := <-senders:
			if email.ssl {
				sendBySsl(object)
			} else {
				err := smtp.SendMail(fmt.Sprintf("%s:%d", email.host, email.port), auth, email.username, object.To, object.convertToBody())
				if err != nil {
					app.Logger().WithField("log_type", "pkg.email.email").Error(err)
				}
			}
		}
	}
}
