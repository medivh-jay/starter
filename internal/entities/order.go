package entities

import "starter/pkg/orm"

// MySQL 示例
type Order struct {
	orm.Database
	ItemId string `json:"item_id" gorm:"column:item_id" form:"item_id" binding:"required"` // 订单id
	Total  int    `json:"total" gorm:"column:total" binding:"required,max=99"`             // 总数量
	Amount int    `json:"amount" gorm:"amount" binding:"required,min=100,max=1000000000"`  // 总金额
}
