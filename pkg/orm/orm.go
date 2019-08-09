package orm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"starter/pkg/config"
	"starter/pkg/unique"
	"sync/atomic"
	"time"
)

type Orm struct {
	Master *gorm.DB
	Slaves []*gorm.DB
}

type Database struct {
	Id        uint64 `gorm:"primary_key;column:id;" json:"id"`
	CreatedAt int    `gorm:"column:created_at;index:created_at" json:"created_at"`
	UpdatedAt int    `gorm:"column:updated_at;index:updated_at" json:"updated_at"`
}

var orm = &Orm{}
var slavesLen int
var err error
var cursor int64 = 0

func createConnectionUrl(username, password, addr, dbName string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", username, password, addr, dbName)
}

// 启动数据库
func Start() {
	database := config.Config.Database
	master := database.Master

	orm.Master, err = gorm.Open("mysql", createConnectionUrl(master.Username, master.Password, master.Addr, master.DbName))
	if err != nil {
		log.Fatalln(err)
	}
	orm.Master.LogMode(true)
	orm.Master.DB().SetMaxIdleConns(database.Master.MaxIdle)
	orm.Master.DB().SetMaxOpenConns(database.Master.MaxOpen)

	for _, slave := range database.Slaves {
		connect, err := gorm.Open("mysql", createConnectionUrl(slave.Username, slave.Password, slave.Addr, slave.DbName))
		if err != nil {
			log.Fatalln(err)
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
	_ = scope.SetColumn("updated_at", t)
	return nil
}
