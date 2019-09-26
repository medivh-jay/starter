package entities

import "starter/pkg/database/orm"

// Order 使用 MySQL gorm 驱动示例
type Order struct {
	orm.Database
	ItemID string `json:"item_id" gorm:"column:item_id" form:"item_id"`                        // 订单id
	Total  int    `json:"total" gorm:"column:total" form:"total" binding:"max=99"`             // 总数量
	Amount int    `json:"amount" gorm:"amount" form:"amount" binding:"min=100,max=1000000000"` // 总金额
}

// TableName 获得表名
func (Order) TableName() string {
	return "orders"
}
