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

// PriceAlert 是用户对某建材的一条降价预警订阅。
// 同一用户对同一建材至多保留一条有效订阅。
// BaselinePrice 为订阅建立时的当前最低价，配合 DropPercent 判断降幅条件；
// 触发后写入触发价与触发时间，状态转为 triggered，且只触发一次。
type PriceAlert struct {
	gorm.Model
	UserID         string `gorm:"index;not null"`
	ProductID      uint   `gorm:"index;not null"`
	Product        Product
	TargetPrice    float64
	DropPercent    float64
	BaselinePrice  float64
	Status         string `gorm:"index;not null"`
	TriggeredPrice float64
	TriggeredAt    *time.Time
}

// Budget 是保存的装修预算试算结果。
type Budget struct {
	gorm.Model
	UserID   string
	RoomType string
	Area     float64
	Estimate float64
	Payload  string
}

// Notification 是触发预警后写入的站内消息。
type Notification struct {
	gorm.Model
	UserID      string `gorm:"index;not null"`
	ProductID   uint
	OfferID     uint
	SupplierID  uint
	Title       string `gorm:"not null"`
	Content     string
	TriggeredAt time.Time
	Read        bool
}
