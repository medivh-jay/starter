package unique

import (
	"github.com/sony/sonyflake"
	"starter/pkg/app"
	"time"
)

var flake = sonyflake.NewSonyflake(sonyflake.Settings{
	StartTime: time.Date(2019, 8, 7, 0, 0, 0, 0, time.Local),
})

// 获取
func Id() uint64 {
	id, err := flake.NextID()
	if err != nil {
		app.Logger().WithField("log_type", "pkg.unique.unique").Error(err)
		return 0
	}
	return id
}
