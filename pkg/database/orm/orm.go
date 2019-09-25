package orm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"starter/pkg/app"
	"starter/pkg/unique"
	"sync/atomic"
	"time"
)

type (
	Orm struct {
		Master *gorm.DB
		Slaves []*gorm.DB
	}

	Database struct {
		Id        uint64 `gorm:"primary_key;column:id;" json:"id"`
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
	cursor    int64 = 0
	conf      config
)

func createConnectionUrl(username, password, addr, dbName string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", username, password, addr, dbName)
}

// 启动数据库
func Start() {
	_ = app.Config().Bind("application", "database", &conf)
	orm.Master, err = gorm.Open("mysql", createConnectionUrl(conf.Master.Username, conf.Master.Password, conf.Master.Addr, conf.Master.DbName))
	if err != nil {
		app.Logger().WithField("log_type", "pkg.orm.orm").Warn("database connect error, you can't use orm support")
		app.Logger().WithField("log_type", "pkg.orm.orm").Warn(err)
	}
	orm.Master.LogMode(true)
	orm.Master.DB().SetMaxIdleConns(conf.Master.MaxIdle)
	orm.Master.DB().SetMaxOpenConns(conf.Master.MaxOpen)

	for _, slave := range conf.Slaves {
		connect, err := gorm.Open("mysql", createConnectionUrl(slave.Username, slave.Password, slave.Addr, slave.DbName))
		if err != nil {
			app.Logger().WithField("log_type", "pkg.orm.orm").Warn("database connect error, you can't use orm support")
			app.Logger().WithField("log_type", "pkg.orm.orm").Warn(err)
		}
		orm.Slaves = append(orm.Slaves, connect)
	}

	slavesLen = len(orm.Slaves)
}

func Slave() *gorm.DB {
	rs := atomic.AddInt64(&cursor, 1)
	return orm.Slaves[rs%int64(slavesLen)]
}

func Master() *gorm.DB {
	return orm.Master
}

func (db *Database) BeforeCreate(scope *gorm.Scope) error {
	if db.Id == 0 {
		_ = scope.SetColumn("id", unique.Id())
	}
	t := time.Now().Unix()
	_ = scope.SetColumn("created_at", t)
	_ = scope.SetColumn("updated_at", t)
	return nil
}

func (db *Database) BeforeUpdate(scope *gorm.Scope) error {
	t := time.Now().Unix()
	scope.Set("gorm:update_column", true)
	_ = scope.SetColumn("updated_at", t)
	return nil
}
