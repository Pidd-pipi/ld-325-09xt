package dto

// CreateAlertRequest 创建价格预警。target_price 与 drop_percent 至少设置其一，
// 约束在 service 层交叉校验（validator 难以表达“二选一”）。
type CreateAlertRequest struct {
	ProductID   uint    `json:"product_id" validate:"required,gt=0"`
	TargetPrice float64 `json:"target_price" validate:"gte=0"`
	DropPercent float64 `json:"drop_percent" validate:"gte=0,lte=100"`
}

// AlertView 是预警订阅在页面上的完整视图：当前最低价、触发价与状态。
type AlertView struct {
	ID             uint     `json:"id"`
	ProductID      uint     `json:"product_id"`
	ProductName    string   `json:"product_name"`
	TargetPrice    *float64 `json:"target_price,omitempty"`
	DropPercent    *float64 `json:"drop_percent,omitempty"`
	BaselinePrice  float64  `json:"baseline_price"`
	TriggerPrice   *float64 `json:"trigger_price,omitempty"`
	CurrentLowest  float64  `json:"current_lowest"`
	Status         string   `json:"status"`
	TriggeredPrice float64  `json:"triggered_price,omitempty"`
	TriggeredAt    string   `json:"triggered_at,omitempty"`
	CreatedAt      string   `json:"created_at"`
}
