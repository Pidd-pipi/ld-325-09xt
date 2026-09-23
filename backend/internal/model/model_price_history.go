package model

import (
	"gorm.io/gorm"
	"time"
)

type PriceHistory struct {
	gorm.Model
	ProductID uint
	OfferID   uint
	Price     float64
	// OldUnitPrice stores the previous quote price when this row records an
	// approved price revision; zero for the daily trend snapshots.
	OldUnitPrice float64
	ChangeSource string
	RecordedAt   time.Time
}
