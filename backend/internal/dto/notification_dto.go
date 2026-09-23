package dto

// NotificationView 站内降价提醒消息。
type NotificationView struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	ProductID   uint   `json:"product_id"`
	Read        bool   `json:"read"`
	TriggeredAt string `json:"triggered_at"`
	CreatedAt   string `json:"created_at"`
}
