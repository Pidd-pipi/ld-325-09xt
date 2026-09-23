package model

import (
	"gorm.io/gorm"
	"time"
)

type Favorite struct {
	gorm.Model
	UserID    string
	ProductID uint
	Folder    string
	Product   Product
}

// PriceAlert is a single active subscription per user and product. It fires at
// most once and then keeps the price and time at which it was triggered.
type PriceAlert struct {
	gorm.Model
	UserID    string `gorm:"index:idx_alert_user_product,unique"`
	ProductID uint   `gorm:"index:idx_alert_user_product,unique"`
	// BasePrice is the lowest product price captured when the subscription was
	// created; it is the reference for drop-percent alerts.
	BasePrice      float64
	TargetPrice    float64
	DropPercent    float64
	Status         string `gorm:"index;not null;default:active"`
	TriggeredPrice float64
	TriggeredAt    *time.Time
	Product        Product
}

type Budget struct {
	gorm.Model
	UserID   string
	RoomType string
	Area     float64
	Estimate float64
	Payload  string
}
