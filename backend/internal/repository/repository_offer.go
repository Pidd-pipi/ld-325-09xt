package repository

import (
	"errors"
	"fmt"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
)

type OfferRepository interface {
	ListByProduct(uint) ([]model.Offer, error)
	Get(uint) (model.Offer, error)
	UpdateStatus(uint, string) (model.Offer, error)
	LowestInStockByProduct(uint) (float64, error)
}

type offerRepository struct{ db *gorm.DB }

func NewOfferRepository(db *gorm.DB) OfferRepository { return &offerRepository{db} }

func (r *offerRepository) ListByProduct(id uint) ([]model.Offer, error) {
	var data []model.Offer
	if err := r.db.Preload("Supplier").Where("product_id = ?", id).Order("unit_price ASC").Find(&data).Error; err != nil {
		return nil, fmt.Errorf("list offers: %w", err)
	}
	return data, nil
}

func (r *offerRepository) Get(id uint) (model.Offer, error) {
	var item model.Offer
	err := r.db.Preload("Supplier").First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return item, apperrors.NewNotFoundError("报价不存在")
	}
	if err != nil {
		return item, fmt.Errorf("find offer: %w", err)
	}
	return item, nil
}

// UpdateStatus 更新库存状态并推进版本号：库存变化同样会使待审修改单过期。
func (r *offerRepository) UpdateStatus(id uint, status string) (model.Offer, error) {
	var item model.Offer
	if err := r.db.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return item, apperrors.NewNotFoundError("报价不存在")
		}
		return item, fmt.Errorf("find offer: %w", err)
	}
	item.StockStatus = status
	item.Version++
	if err := r.db.Save(&item).Error; err != nil {
		return item, fmt.Errorf("update offer: %w", err)
	}
	return item, nil
}

// LowestInStockByProduct 返回该建材当前“有货”报价中的最低单价；
// 无在售报价时返回 0。预警判定与提醒页的当前最低价均以此为准。
func (r *offerRepository) LowestInStockByProduct(productID uint) (float64, error) {
	var lowest *float64
	err := r.db.Model(&model.Offer{}).
		Where("product_id = ? AND stock_status = ?", productID, constants.StatusInStock).
		Select("MIN(unit_price)").Row().Scan(&lowest)
	if err != nil {
		return 0, fmt.Errorf("lowest in-stock offer: %w", err)
	}
	if lowest == nil {
		return 0, nil
	}
	return *lowest, nil
}

// ApplyChange 在调用方提供的事务内改写报价内容并推进版本号。
func ApplyChange(tx *gorm.DB, offer *model.Offer, change model.OfferChange) error {
	offer.UnitPrice = change.NewUnitPrice
	offer.Freight = change.NewFreight
	offer.DeliveryDays = change.NewDeliveryDays
	offer.StockStatus = change.NewStockStatus
	offer.Version++
	if err := tx.Save(offer).Error; err != nil {
		return fmt.Errorf("apply offer change: %w", err)
	}
	return nil
}

// FindOfferForUpdate 在事务内读取报价行，供版本核对。
func FindOfferForUpdate(tx *gorm.DB, id uint) (model.Offer, error) {
	var offer model.Offer
	if err := tx.First(&offer, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return offer, apperrors.NewNotFoundError("报价不存在")
		}
		return offer, fmt.Errorf("find offer in tx: %w", err)
	}
	return offer, nil
}

// LowestInStockByProductTx 在事务内读取某建材当前有货最低价，
// 供审核通过后在同一事务内判断是否触发预警。
func LowestInStockByProductTx(tx *gorm.DB, productID uint) (float64, error) {
	var lowest *float64
	err := tx.Model(&model.Offer{}).
		Where("product_id = ? AND stock_status = ?", productID, constants.StatusInStock).
		Select("MIN(unit_price)").Row().Scan(&lowest)
	if err != nil {
		return 0, fmt.Errorf("lowest in-stock offer in tx: %w", err)
	}
	if lowest == nil {
		return 0, nil
	}
	return *lowest, nil
}
