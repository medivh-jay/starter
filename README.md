## STARTER  [wiki](https://github.com/medivh-jay/starter/wiki/%E5%A6%82%E4%BD%95%E4%BD%BF%E7%94%A8)

- 本项目是一个整合了一些基础开发工具并稍微封装了一些日常操作的工具整合
- 可以使用它进行正常的开发而不必要去一个个重新装go库，重新组织结构
- 基本做到下载, 定义表, 定义控制器, 定义路由, 一键CURD, 做一个快乐的CURD boy
- 直接下载本项目就行
- 修改 module name 为自己的
- 基本需求  mysql  和 mongo 是必须的, 可以使用 docker 安装
- 在使用时可按自己需要全盘修改所有内容,毕竟每个人的需求都不同
- 工具做了一些配置管理, configs 配置文件目录下氛围  debug, release 和 test , 分别对应 gin 的三个mode, 指定不同的mode将加载不同目录下的配置

##### 下边是一些使用到的库
- 路由使用了gin
- 验证器使用了validator.v9
- orm使用了gorm
- mongo使用了mongo官方驱动
- 支持了mgo,自行选择使用的库, 分别是 mongo.Start() 和 mgo.Start(),同一张表结构只能被一个驱动操作,因为两个驱动某些属性不同可能混合操作会出现问题
- redis使用了go-redis
- elasticsearch 使用了 github.com/olivere/elastic/v7
- 使用了 gin 的session 中间件，使用redis存储
- 使用了github.com/dgrijalva/jwt-go作为json web token 验证器
- pkg/managers 包实现了 mongo和MySQL的curd操作，可提供后台管理api
- password 提供了 密码hash 和 密码验证
- 使用了 sony 唯一id生成工具作为MySQL默认主键id
- 使用了 toml 作为配置工具
- 使用 swagger 作为文档生成工具
- 使用了 facebook/gracehttp 作为优雅重启方案
- 编译之后在本项目根路径运行才能正确读取 configs 等各种静态资源
- Makefile 需要自己按需建立
- 如果需要帮助可以查看 [wiki](https://github.com/medivh-jay/starter/wiki/%E5%A6%82%E4%BD%95%E4%BD%BF%E7%94%A8) 或者邮件 jay.medivh@gmail.com 

##### 一个前端基本不会的人拼界面真的很不容易, 放个图好了
![admin](web/admin/static/images/admin.png)

- 感谢 

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
