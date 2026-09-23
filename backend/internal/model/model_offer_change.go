package model

import (
	"gorm.io/gorm"
	"time"
)

// OfferChange 是供应商提交的报价修改单，进入平台审核流程。
// BaseVersion 记录提交时报价的版本；审核时若报价已推进到新版本，
// 该修改单被标记为 stale（失效）并返回冲突，由供应商基于新价重新提交。
type OfferChange struct {
	gorm.Model
	OfferID         uint     `gorm:"index;not null"`
	Offer           Offer    `gorm:"foreignKey:OfferID"`
	ProductID       uint     `gorm:"index;not null"`
	Product         Product  `gorm:"foreignKey:ProductID"`
	SupplierID      uint     `gorm:"index;not null"`
	Supplier        Supplier `gorm:"foreignKey:SupplierID"`
	BaseUnitPrice   float64  `gorm:"not null"`
	NewUnitPrice    float64  `gorm:"not null"`
	NewFreight      string
	NewDeliveryDays int
	NewStockStatus  string
	BaseVersion     uint   `gorm:"not null"`
	Status          string `gorm:"index;not null;uniqueIndex:idx_offer_pending,where:status='pending'"`
	ReviewRemark    string
	ReviewedAt      *time.Time
}
