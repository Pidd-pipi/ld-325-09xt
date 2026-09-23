package model

import "gorm.io/gorm"

// Offer 是某供应商针对某建材的当前生效报价。
// Version 为乐观版本号：每次报价内容被改写（审核通过、库存状态变更）后递增，
// 待审修改单通过 BaseVersion 判断自己是否已过期。
type Offer struct {
	gorm.Model
	ProductID    uint
	SupplierID   uint
	Supplier     Supplier
	UnitPrice    float64
	MOQ          int
	Freight      string
	DeliveryDays int
	StockStatus  string
	Version      uint `gorm:"not null;default:1"`
}
