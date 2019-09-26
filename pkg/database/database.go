package database

// Table 数据表结构体必须实现此接口
type Table interface {
	TableName() string
}

// UpdateOrCreate 更新或增加时的前置操作, 在对应数据结构体中实现, 只作用在 managers 操作下
type UpdateOrCreate interface {
	PreOperation()
}
