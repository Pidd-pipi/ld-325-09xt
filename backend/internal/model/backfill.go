package model

import (
	"fmt"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"gorm.io/gorm"
)

// BackfillDefaults 为旧版本数据补齐新列默认值：
// 报价版本号从 0 回填为初始版本，预警空状态回填为生效中。
func BackfillDefaults(db *gorm.DB) error {
	if err := db.Exec("UPDATE offers SET version = ? WHERE version = 0", constants.OfferInitialVersion).Error; err != nil {
		return fmt.Errorf("backfill offer version: %w", err)
	}
	if err := db.Exec("UPDATE price_alerts SET status = ? WHERE status = '' OR status IS NULL", constants.AlertStatusActive).Error; err != nil {
		return fmt.Errorf("backfill alert status: %w", err)
	}
	return nil
}
