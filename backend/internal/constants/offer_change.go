package constants

// OfferChangeStatus 是商家提交的报价修改单生命周期状态。
const (
	ChangeStatusPending  = "pending"  // 待审核
	ChangeStatusApproved = "approved" // 审核通过，新价已生效
	ChangeStatusRejected = "rejected" // 审核驳回
	ChangeStatusStale    = "stale"    // 审核时报价版本已变化，修改单失效
)

// 审核动作。
const (
	ChangeReviewApprove = "approve" // 审核动作：通过
	ChangeReviewReject  = "reject"  // 审核动作：驳回
)

// PriceAlertStatus 是价格预警订阅的状态。
const (
	AlertStatusActive    = "active"    // 订阅生效中
	AlertStatusTriggered = "triggered" // 已触发一次，等待用户处理
	AlertStatusInactive  = "inactive"  // 用户停用
)

const (
	OfferInitialVersion = 1 // 报价创建时的初始版本
)
