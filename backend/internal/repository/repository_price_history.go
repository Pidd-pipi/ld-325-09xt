package repository

import (
	"fmt"
	"time"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
)

type PriceHistoryRepository interface {
	List(productID uint, since time.Time) ([]model.PriceHistory, error)
	RecordRevision(tx *gorm.DB, row model.PriceHistory) error
}

type priceHistoryRepository struct{ db *gorm.DB }

func NewPriceHistoryRepository(db *gorm.DB) PriceHistoryRepository {
	return &priceHistoryRepository{db}
}

func (r *priceHistoryRepository) List(productID uint, since time.Time) ([]model.PriceHistory, error) {
	var rows []model.PriceHistory
	if err := r.db.Where("product_id = ? AND recorded_at >= ?", productID, since).Order("recorded_at ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list history: %w", err)
	}
	return rows, nil
}

// RecordRevision archives the previous quote price before a new revision takes
// effect, so the old price remains visible in the trend history.
func (r *priceHistoryRepository) RecordRevision(tx *gorm.DB, row model.PriceHistory) error {
	if row.RecordedAt.IsZero() {
		row.RecordedAt = time.Now()
	}
	if row.ChangeSource == "" {
		row.ChangeSource = constants.PriceSourceRevision
	}
	if err := tx.Create(&row).Error; err != nil {
		return fmt.Errorf("record price history: %w", err)
	}
	return nil
}
