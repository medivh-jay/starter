package database

// 数据表结构体必须实现此接口
type Table interface {
	TableName() string
}

type UpdateOrCreate interface {
	PreOperation()
}
