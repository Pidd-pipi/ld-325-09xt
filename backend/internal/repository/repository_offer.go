package repository

import (
	"errors"
	"fmt"

	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OfferRepository interface {
	ListByProduct(uint) ([]model.Offer, error)
	ListBySupplier(uint) ([]model.Offer, error)
	Get(id uint) (model.Offer, error)
	UpdateStatus(id uint, status string) (model.Offer, error)
	MinUnitPrice(tx *gorm.DB, productID uint) (float64, error)
	// Transactional operations used by the change-review workflow.
	LockForUpdate(tx *gorm.DB, id uint) (model.Offer, error)
	ApplyRevision(tx *gorm.DB, offer model.Offer) error
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

func (r *offerRepository) ListBySupplier(id uint) ([]model.Offer, error) {
	var data []model.Offer
	if err := r.db.Preload("Product").Where("supplier_id = ?", id).Order("updated_at DESC").Find(&data).Error; err != nil {
		return nil, fmt.Errorf("list supplier offers: %w", err)
	}
	return data, nil
}

func (r *offerRepository) Get(id uint) (model.Offer, error) {
	var offer model.Offer
	if err := r.db.First(&offer, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return offer, apperrors.ErrNotFound
		}
		return offer, fmt.Errorf("find offer: %w", err)
	}
	return offer, nil
}

func (r *offerRepository) UpdateStatus(id uint, status string) (model.Offer, error) {
	var item model.Offer
	if err := r.db.First(&item, id).Error; err != nil {
		return item, fmt.Errorf("find offer: %w", err)
	}
	item.StockStatus = status
	if err := r.db.Save(&item).Error; err != nil {
		return item, fmt.Errorf("update offer: %w", err)
	}
	return item, nil
}

func (r *offerRepository) MinUnitPrice(tx *gorm.DB, productID uint) (float64, error) {
	var lowest *float64
	if err := tx.Model(&model.Offer{}).
		Where("product_id = ?", productID).
		Select("MIN(unit_price)").
		Scan(&lowest).Error; err != nil {
		return 0, fmt.Errorf("min offer price: %w", err)
	}
	if lowest == nil {
		return 0, nil
	}
	return *lowest, nil
}

func (r *offerRepository) LockForUpdate(tx *gorm.DB, id uint) (model.Offer, error) {
	var offer model.Offer
	// PostgreSQL honors FOR UPDATE; it is a no-op clause on other dialects.
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&offer, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return offer, apperrors.ErrNotFound
		}
		return offer, fmt.Errorf("lock offer: %w", err)
	}
	return offer, nil
}

func (r *offerRepository) ApplyRevision(tx *gorm.DB, offer model.Offer) error {
	if err := tx.Save(&offer).Error; err != nil {
		return fmt.Errorf("apply offer revision: %w", err)
	}
	return nil
}
