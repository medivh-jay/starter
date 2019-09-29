package orm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // 使用MySQL
	"starter/pkg/app"
	"starter/pkg/unique"
	"sync/atomic"
	"time"
)

type (
	// Orm gorm 连接对象, 包含Master和Slaves, 由配置决定, Slaves 使用 atomic 包进行循环获取
	Orm struct {
		Master *gorm.DB
		Slaves []*gorm.DB
	}
	// Database gorm 支持的数据嵌套, 自定义的数据表结构体导入该结构体，将默认拥有这三个字段
	Database struct {
		ID        uint64 `gorm:"primary_key;column:id;" json:"id"`
		CreatedAt int    `gorm:"column:created_at;index:created_at" json:"created_at"`
		UpdatedAt int    `gorm:"column:updated_at;index:updated_at" json:"updated_at"`
	}

	connInfo struct {
		Addr     string `toml:"addr"`
		Username string `toml:"username"`
		Password string `toml:"password"`
		DbName   string `toml:"dbname"`
		MaxIdle  int    `toml:"max_idle"`
		MaxOpen  int    `toml:"max_open"`
	}

	config struct {
		Master connInfo   `toml:"master"`
		Slaves []connInfo `toml:"slave"`
	}
)

var (
	orm       = &Orm{}
	slavesLen int
	err       error
	cursor    int64
	conf      config
)

func createConnectionURL(username, password, addr, dbName string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", username, password, addr, dbName)
}

// Start 启动数据库
func Start() {
	err = app.Config().Bind("application", "database", &conf)
	if err == app.ErrNodeNotExists {
		return
	}
	orm.Master, err = gorm.Open("mysql", createConnectionURL(conf.Master.Username, conf.Master.Password, conf.Master.Addr, conf.Master.DbName))
	if err != nil {
		app.Logger().WithField("log_type", "pkg.orm.orm").Warn("database connect error, you can't use orm support: ")
		app.Logger().WithField("log_type", "pkg.orm.orm").Warn(err)
	}
	orm.Master.LogMode(true)
	orm.Master.DB().SetMaxIdleConns(conf.Master.MaxIdle)
	orm.Master.DB().SetMaxOpenConns(conf.Master.MaxOpen)

	for _, slave := range conf.Slaves {
		connect, err := gorm.Open("mysql", createConnectionURL(slave.Username, slave.Password, slave.Addr, slave.DbName))
		if err != nil {
			app.Logger().WithField("log_type", "pkg.orm.orm").Warn("database connect error, you can't use orm support")
			app.Logger().WithField("log_type", "pkg.orm.orm").Warn(err)
		}
		orm.Slaves = append(orm.Slaves, connect)
	}

	slavesLen = len(orm.Slaves)
}

// Slave 获得一个从库连接对象, 使用 atomic.AddInt64 计算调用次数，然后按 Slave 连接个数和次数进行取模操作之后获取指定index的Slave
func Slave() *gorm.DB {
	rs := atomic.AddInt64(&cursor, 1)
	return orm.Slaves[rs%int64(slavesLen)]
}

// Master 获得主库连接
func Master() *gorm.DB {
	return orm.Master
}

// BeforeCreate 创建数据前置操作, 自定义结构体可重新实现该方法
func (db *Database) BeforeCreate(scope *gorm.Scope) error {
	if db.ID == 0 {
		_ = scope.SetColumn("id", unique.ID())
	}
	t := time.Now().Unix()
	_ = scope.SetColumn("created_at", t)
	_ = scope.SetColumn("updated_at", t)
	return nil
}

// BeforeUpdate 更新数据前置操作, 自定义结构体可重新实现该方法
func (db *Database) BeforeUpdate(scope *gorm.Scope) error {
	t := time.Now().Unix()
	scope.Set("gorm:update_column", true)
	_ = scope.SetColumn("updated_at", t)
	return nil
}
