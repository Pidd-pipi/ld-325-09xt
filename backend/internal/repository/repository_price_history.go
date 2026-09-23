package repository

import (
	"fmt"
	"time"

	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
)

type PriceHistoryRepository interface {
	List(uint, time.Time) ([]model.PriceHistory, error)
	Create(history *model.PriceHistory) error
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

// Create 写入一条价格历史；service 在事务中调用时传入绑定事务的仓储。
func (r *priceHistoryRepository) Create(history *model.PriceHistory) error {
	if err := r.db.Create(history).Error; err != nil {
		return fmt.Errorf("create price history: %w", err)
	}
	return nil
}

// NewPriceHistoryRepositoryTx 绑定到已有事务（例如审核事务）。
func NewPriceHistoryRepositoryTx(tx *gorm.DB) PriceHistoryRepository {
	return &priceHistoryRepository{db: tx}
}
