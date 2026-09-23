package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
)

type OfferChangeRepository interface {
	Create(change *model.OfferChange) error
	PendingExistsByOffer(offerID uint) (bool, error)
	Get(id uint) (model.OfferChange, error)
	ListBySupplier(supplierID uint) ([]model.OfferChange, error)
	ListByStatus(status string) ([]model.OfferChange, error)
	Save(change *model.OfferChange) error
	MarkStatus(id uint, status, remark string) error
}

type offerChangeRepository struct{ db *gorm.DB }

func NewOfferChangeRepository(db *gorm.DB) OfferChangeRepository {
	return &offerChangeRepository{db}
}

func (r *offerChangeRepository) Create(change *model.OfferChange) error {
	if err := r.db.Create(change).Error; err != nil {
		if isDuplicateKey(err) {
			return apperrors.NewConflictError("该报价已有待审核修改，请等待审核结果后再提交")
		}
		return fmt.Errorf("create offer change: %w", err)
	}
	return nil
}

// PendingExistsByOffer 报告某报价是否已有待审修改单。
// 同一报价同时只允许一条待审修改：已有待审时，供应商后来的提交不入库。
func (r *offerChangeRepository) PendingExistsByOffer(offerID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.OfferChange{}).
		Where("offer_id = ? AND status = ?", offerID, constants.ChangeStatusPending).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("count pending offer change: %w", err)
	}
	return count > 0, nil
}

func (r *offerChangeRepository) Get(id uint) (model.OfferChange, error) {
	var change model.OfferChange
	err := r.db.Preload("Offer").Preload("Product").Preload("Supplier").First(&change, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return change, apperrors.NewNotFoundError("报价修改单不存在")
	}
	if err != nil {
		return change, fmt.Errorf("get offer change: %w", err)
	}
	return change, nil
}

func (r *offerChangeRepository) ListBySupplier(supplierID uint) ([]model.OfferChange, error) {
	var rows []model.OfferChange
	err := r.db.Preload("Offer").Preload("Product").Preload("Supplier").
		Where("supplier_id = ?", supplierID).
		Order("created_at DESC").Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list offer changes by supplier: %w", err)
	}
	return rows, nil
}

func (r *offerChangeRepository) ListByStatus(status string) ([]model.OfferChange, error) {
	var rows []model.OfferChange
	err := r.db.Preload("Offer").Preload("Product").Preload("Supplier").
		Where("status = ?", status).
		Order("created_at ASC").Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list offer changes by status: %w", err)
	}
	return rows, nil
}

func (r *offerChangeRepository) Save(change *model.OfferChange) error {
	if err := r.db.Save(change).Error; err != nil {
		return fmt.Errorf("save offer change: %w", err)
	}
	return nil
}

// NewOfferChangeRepositoryTx 绑定到已有事务。
func NewOfferChangeRepositoryTx(tx *gorm.DB) OfferChangeRepository {
	return &offerChangeRepository{db: tx}
}

// MarkStatus 把待审修改单更新为指定终态（reject 路径使用）。
func (r *offerChangeRepository) MarkStatus(id uint, status, remark string) error {
	var change model.OfferChange
	if err := r.db.First(&change, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("报价修改单不存在")
		}
		return fmt.Errorf("find offer change: %w", err)
	}
	if change.Status != constants.ChangeStatusPending {
		return apperrors.NewConflictError("该修改单已被处理，请勿重复审核")
	}
	now := time.Now()
	change.Status = status
	change.ReviewRemark = remark
	change.ReviewedAt = &now
	if err := r.db.Save(&change).Error; err != nil {
		return fmt.Errorf("mark offer change: %w", err)
	}
	return nil
}
