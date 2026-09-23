package dto

// SubmitOfferChangeRequest 供应商提交新的单价、运费、货期与库存状态。
type SubmitOfferChangeRequest struct {
	OfferID      uint    `json:"offer_id" validate:"required,gt=0"`
	UnitPrice    float64 `json:"unit_price" validate:"required,gt=0"`
	Freight      string  `json:"freight" validate:"max=100"`
	DeliveryDays int     `json:"delivery_days" validate:"required,gt=0,lte=365"`
	StockStatus  string  `json:"stock_status" validate:"required,oneof=in_stock out_of_stock discontinued"`
}

// ReviewOfferChangeRequest 管理员审核修改单。
type ReviewOfferChangeRequest struct {
	Action string `json:"action" validate:"required,oneof=approve reject"`
	Remark string `json:"remark" validate:"max=200"`
}

// OfferChangeView 供应商与审核台共用的修改单视图。
type OfferChangeView struct {
	ID              uint    `json:"id"`
	OfferID         uint    `json:"offer_id"`
	ProductID       uint    `json:"product_id"`
	ProductName     string  `json:"product_name"`
	SupplierID      uint    `json:"supplier_id"`
	SupplierName    string  `json:"supplier_name"`
	BaseUnitPrice   float64 `json:"base_unit_price"`
	NewUnitPrice    float64 `json:"new_unit_price"`
	NewFreight      string  `json:"new_freight"`
	NewDeliveryDays int     `json:"new_delivery_days"`
	NewStockStatus  string  `json:"new_stock_status"`
	BaseVersion     uint    `json:"base_version"`
	CurrentVersion  uint    `json:"current_version"`
	Status          string  `json:"status"`
	ReviewRemark    string  `json:"review_remark,omitempty"`
	CreatedAt       string  `json:"created_at"`
	ReviewedAt      string  `json:"reviewed_at,omitempty"`
}
