package dto

import "time"

// SubmitOfferChangeRequest is a supplier revision of unit price, freight,
// delivery window and stock status.
type SubmitOfferChangeRequest struct {
	UnitPrice    float64 `json:"unit_price" validate:"gt=0"`
	Freight      string  `json:"freight" validate:"required,max=200"`
	DeliveryDays int     `json:"delivery_days" validate:"gte=0,lte=365"`
	StockStatus  string  `json:"stock_status" validate:"required,oneof=in_stock out_of_stock discontinued"`
}

type ReviewOfferChangeRequest struct {
	Approve bool   `json:"approve"`
	Note    string `json:"note" validate:"max=200"`
}

type OfferChangeView struct {
	ID              uint       `json:"id"`
	OfferID         uint       `json:"offer_id"`
	ProductID       uint       `json:"product_id"`
	ProductName     string     `json:"product_name"`
	SupplierID      uint       `json:"supplier_id"`
	SupplierName    string     `json:"supplier_name"`
	CurrentPrice    float64    `json:"current_price"`
	NewUnitPrice    float64    `json:"new_unit_price"`
	CurrentFreight  string     `json:"current_freight"`
	NewFreight      string     `json:"new_freight"`
	CurrentDelivery int        `json:"current_delivery_days"`
	NewDeliveryDays int        `json:"new_delivery_days"`
	CurrentStock    string     `json:"current_stock_status"`
	NewStockStatus  string     `json:"new_stock_status"`
	BaseVersion     int        `json:"base_version"`
	OfferVersion    int        `json:"offer_version"`
	Status          string     `json:"status"`
	ReviewNote      string     `json:"review_note"`
	CreatedAt       time.Time  `json:"created_at"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
}
