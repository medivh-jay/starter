## 项目

- 本项目作为启动器，整合了需要用到的常用三方库,并且封装了部分操作
- 在使用时可按自己需要全盘修改所有内容
- 在自己的启动文件中设置 gin.Mode 为 release 之后程序将读取 application.toml 而不是 development.toml
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

### 使用方式
- 下载此项目的 zip 包， 然后将 go.mod 的moudle name 替换为自己定义的名字,然后将对应原来代码里边的 starter 全部替换为自己定义的名字即可
- 建议业务代码写到 internal 下边，按模块分类, 比如internal/admin 为管理相关代码,  internal/api 为接口相关代码(里边目前有两个默认的,只是示例,可以直接删除掉)
- 然后将 main.go 在 cmd 目录下建立对应的 admin 和 api 目录，分别在里边建立 main.go
- session中间件需要在调用 server.Run() 之前注入
- 如果要提供普通的增删改查操作, 可以查看 internal/manager/router.go 里边的示例,基本只需要建立好数据模型的结构体之后就能进行管理的CURD和基本查询操作了
- 
