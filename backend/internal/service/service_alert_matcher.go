package service

import (
	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/model"
)

// effectiveTriggerPrice 计算订阅的有效触发价：
// 目标价与按降幅（相对订阅建立时的基线价）折算价同时存在时取较低者，
// 即“符合目标价或降幅”任一条件即触发。没有任何有效条件时返回 nil。
func effectiveTriggerPrice(alert model.PriceAlert) *float64 {
	var candidates []float64
	if alert.TargetPrice > 0 {
		candidates = append(candidates, alert.TargetPrice)
	}
	if alert.DropPercent > 0 && alert.BaselinePrice > 0 {
		dropPrice := alert.BaselinePrice * (1 - alert.DropPercent/constants.PercentBase)
		candidates = append(candidates, dropPrice)
	}
	if len(candidates) == 0 {
		return nil
	}
	threshold := candidates[0]
	for _, value := range candidates[1:] {
		if value < threshold {
			threshold = value
		}
	}
	return &threshold
}

// matchAlert 判断订阅是否应被当前“有货最低价”触发。
// 返回是否命中以及命中时的触发价。无在售报价时不触发。
func matchAlert(alert model.PriceAlert, lowest float64) (bool, float64) {
	if lowest <= 0 {
		return false, 0
	}
	threshold := effectiveTriggerPrice(alert)
	if threshold == nil {
		return false, 0
	}
	return lowest <= *threshold, *threshold
}
