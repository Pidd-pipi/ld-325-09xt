package model

import "gorm.io/gorm"

// OfferChange is a supplier-submitted quote revision that awaits admin review.
type OfferChange struct {
	gorm.Model
	OfferID         uint
	Offer           Offer
	SupplierID      uint
	Supplier        Supplier
	NewUnitPrice    float64
	NewFreight      string
	NewDeliveryDays int
	NewStockStatus  string
	// BaseVersion records the quote version the supplier was editing. Approving a
	// change whose base no longer matches the live version fails with a conflict.
	BaseVersion int
	Status      string `gorm:"index;not null;default:pending"`
	ReviewNote  string
}
