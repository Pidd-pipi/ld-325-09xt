package dto

import "time"

type CreateAlertRequest struct {
	ProductID   uint    `json:"product_id" validate:"required,gt=0"`
	TargetPrice float64 `json:"target_price" validate:"gte=0"`
	DropPercent float64 `json:"drop_percent" validate:"gte=0,lte=100"`
}

// AlertView is the alert-page payload: current lowest price, the price at which
// the subscription fires, and its lifecycle status.
type AlertView struct {
	ID             uint       `json:"id"`
	ProductID      uint       `json:"product_id"`
	ProductName    string     `json:"product_name"`
	Brand          string     `json:"brand"`
	Model          string     `json:"model"`
	Unit           string     `json:"unit"`
	BasePrice      float64    `json:"base_price"`
	TargetPrice    float64    `json:"target_price"`
	DropPercent    float64    `json:"drop_percent"`
	TriggerPrice   float64    `json:"trigger_price"`
	CurrentLowest  float64    `json:"current_lowest"`
	Status         string     `json:"status"`
	TriggeredPrice float64    `json:"triggered_price,omitempty"`
	TriggeredAt    *time.Time `json:"triggered_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}
