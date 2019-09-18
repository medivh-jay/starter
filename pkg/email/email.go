// 发送邮件
// 	email.Send(email.NewHtmlSender("这里是标题", email.ParseHtml(app.FilePath() + "/public/email/test.html", struct {
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
	Email struct {
		username string // 发送账户
		password string // 账户密码
		host     string // smtp服务器地址
		port     int    // smtp服务器端口
		ssl      bool
	}
	Object struct {
		To      []string          // 收件人邮件地址
		Header  map[string]string // 邮件头
		Content string            // 邮件正文
	}
	Html struct {
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

// 启动邮件服务
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

func ParseHtml(path string, data interface{}) *Html {
	html := new(Html)
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

// 发送邮件
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

// 发送邮件
func NewHtmlSender(subject string, html *Html, to ...string) *Object {
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
