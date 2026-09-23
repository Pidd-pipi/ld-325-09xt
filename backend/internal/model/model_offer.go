package model

import "gorm.io/gorm"

type Offer struct {
	gorm.Model
	ProductID    uint
	Product      Product
	SupplierID   uint
	Supplier     Supplier
	UnitPrice    float64
	MOQ          int
	Freight      string
	DeliveryDays int
	StockStatus  string
	// Version is the optimistic-concurrency token of the approved quote.
	Version int `gorm:"not null;default:1"`
}
