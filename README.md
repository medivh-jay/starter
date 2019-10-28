<p align="center">
<img align="center" src="assets/STARTER-logo.svg" width="300px" />
</p>
<h3 align="center">STARTER - WEB Develop Integration 
  <a href="https://goreportcard.com/report/github.com/medivh-jay/starter"><img src="https://goreportcard.com/badge/github.com/medivh-jay/starter" /></a>&nbsp;
  <a href="https://travis-ci.org/medivh-jay/starter"><img src="https://api.travis-ci.org/medivh-jay/starter.svg?branch=master" /></a>&nbsp;
  <a href="https://godoc.org/github.com/medivh-jay/starter"><img src="https://godoc.org/github.com/medivh-jay/starter?status.svg" /></a>&nbsp;
  <a align="center" href="https://github.com/medivh-jay/starter/wiki/Step-1:%E4%B8%8B%E8%BD%BD%E6%9C%AC%E9%A1%B9%E7%9B%AE">wiki</a></h3>

- 提供一个完整的目录架构来支持go的web,微服务,restful接口的开发
- 本身包含了各种常用三方代码库,提供大部分基础功能,并完善基础建设
- 提供了基本完整的后端admin的curl接口操作和一个稍显简陋(前端开发经验不足)的后端界面
- 提供的配置管理功能可按需添加,除了基础的服务监听端口之外,不需要的服务都可以不必添加配置,当然如果你需要某些服务,那么配置就是必须的了
- 本身提供了一个 docker-compose.yaml 文件以支持下载 starter 之后通过docker快速的体验starter功能
- 使用 gin 的三种mode进行配置环境的管理, 分别为 debug, release,和test

### 说明
本项目确实提供了一个开发的目录架构,由 [project-layout](https://github.com/golang-standards/project-layout) 提供, 目前来说, project-layout 提供的接口能更好的规范代码结构


#### Quick Install
```bash
go get github.com/medivh-jay/gocreator
gocreator -m mode-name
```

## 功能

* 后端admin管理curl-api 和简单的 rbac 功能支持
* 验证码功能 
* elasticsearch服务
* 邮件发送功能
* 分页工具提供更简单的分页接口功能 
* 提供更简单方便的 i18n 国际化模块
* 使用logrus记录日志 
* mgo 支持
* mongodb 官方 mongo-driver 支持
* JWT (JsonWebToken) 支持
* 提供敏感词过滤功能
* 提供接口 csrf token 生成和验证支持
* 提供了简单的 password hash 支持
* 提供 kafka 消息队列 
* redis 支持
* 使用 validator.v9 作为表单验证组件
* 使用 gin 作为基础web服务框架
* 后端目前由 LayUI 支持

## 目录结构
```code 
├── LICENSE
├── Makefile
├── README.md
├── api
├── assets
├── build
├── cmd
├── configs
│   ├── README.md
│   ├── debug
│   │   ├── admin
│   │   ├── manager
│   │   └── services
│   ├── release
│   └── test
├── docker-compose.yaml
├── docs
├── examples
├── go.mod
├── go.sum
├── internal
├── locales
│   ├── README.md
│   └── admin
│       ├── en.toml
│       └── zh.toml
├── logs
├── pkg
│   ├── app
│   ├── captcha
│   ├── database
│   ├── elastic
│   ├── email
│   ├── i18n
│   ├── log
│   ├── middlewares
│   ├── pager
│   ├── password
│   ├── payments
│   ├── queue
│   ├── rbac
│   ├── redis
│   ├── sensitivewords
│   ├── server
│   ├── sessions
│   ├── unique
│   └── validator
├── scripts
├── test
├── tools
└── web

```
### 你首先需要了解的目录
##### cmd 
cmd 目录提供的是程序的入口文件,这里不应该包含太多的代码, 一般一个 main.go 足矣, 由 main.go 关联 internal 内部代码编译具体服务,编译的二进制程序应该存在于项目根目录

##### configs
configs 是配置目录, 如上边结构, 比如你有一个二进制程序为 app , 那么对应的你应该在 configs/debug/app 下建立个人开发配置文件, 同理, configs/test/app 为测试配置, configs/release/app 为正式配置
程序本身需要的基本配置文件为
- application.toml 应用基本配置, 可参考目前代码里边的 configs/debug/manager/application.toml
- payments.toml 支付配置信息
- validator-messages.toml validator.v9 的国际化配置信息
- keywords.csv 敏感词文件, csv,以行为单位

##### internal 
程序业务代码, 比如 models, controllers 等, 这里的代码属于自己规划了

##### locales
国际化翻译文件

##### logs 
日志文件

##### pkg 
是我所写的基础功能代码


##### 其他目录可参考目录下的 README.md 

#### 还有
本项目本身 internal, cmd, configs, locales 下边包含了我写的使用示例文件和代码,在使用本工具时,可清除里边的代码,写自己的代码
,web目录下边是使用 layUI 做的一个简单的web后端, 如果不需要, 也可直接删除

##### 提供基础CURD后端界面操作,wiki有模板使用说明
![admin](web/admin/static/images/admin.png)

## 感谢所有开源项目的支持 

- [LayUI](https://www.layui.com/)
- [X-admin](http://x.xuebingsi.com/)
- [gin](https://github.com/gin-gonic/gin)
- [BurntSushi/toml](https://github.com/BurntSushi/toml)
- [dgrijalva/jwt-go](https://github.com/dgrijalva/jwt-go)
- [jinzhu/gorm](https://github.com/jinzhu/gorm)
- [json-iterator](https://github.com/json-iterator/go)
- [sony/sonyflake](https://github.com/sony/sonyflake)
- [swaggo](https://github.com/swaggo)
- [mgo](https://gopkg.in/mgo.v2)
- [wangEditor](http://www.wangeditor.com/index.html)
- [mojocn/base64Captcha](github.com/mojocn/base64Captcha)
- [elogrus](https://github.com/sohlich/elogrus)
- [rifflock/lfshook](https://github.com/rifflock/lfshook)
