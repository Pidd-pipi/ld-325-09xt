package dto

// DemoTokenRequest 演示环境的身份切换请求，仅用于本地/演示 RBAC 体验。
type DemoTokenRequest struct {
	Role string `json:"role" validate:"required,oneof=admin supplier user"`
}

// DemoTokenView 各角色对应的演示令牌。
type DemoTokenView struct {
	Role      string `json:"role"`
	UserID    string `json:"user_id"`
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}
