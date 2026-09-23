package repository

import (
	"fmt"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
)

type AlertRepository interface {
	Create(alert *model.PriceAlert) error
	Get(id uint) (model.PriceAlert, error)
	FindActiveByUserProduct(userID string, productID uint) (model.PriceAlert, error)
	ListByUser(userID string) ([]model.PriceAlert, error)
	Save(alert *model.PriceAlert) error
}

type alertRepository struct{ db *gorm.DB }

func NewAlertRepository(db *gorm.DB) AlertRepository { return &alertRepository{db} }

func (r *alertRepository) Create(alert *model.PriceAlert) error {
	if err := r.db.Create(alert).Error; err != nil {
		return fmt.Errorf("create alert: %w", err)
	}
	return nil
}

func (r *alertRepository) Get(id uint) (model.PriceAlert, error) {
	var alert model.PriceAlert
	err := r.db.Preload("Product").First(&alert, id).Error
	if err == gorm.ErrRecordNotFound {
		return alert, apperrors.NewNotFoundError("价格提醒不存在")
	}
	if err != nil {
		return alert, fmt.Errorf("get alert: %w", err)
	}
	return alert, nil
}

func (r *alertRepository) FindActiveByUserProduct(userID string, productID uint) (model.PriceAlert, error) {
	var alert model.PriceAlert
	err := r.db.Where("user_id = ? AND product_id = ? AND status = ?",
		userID, productID, constants.AlertStatusActive).First(&alert).Error
	if err == gorm.ErrRecordNotFound {
		return alert, apperrors.ErrNotFound
	}
	if err != nil {
		return alert, fmt.Errorf("find active alert: %w", err)
	}
	return alert, nil
}

func (r *alertRepository) ListByUser(userID string) ([]model.PriceAlert, error) {
	var rows []model.PriceAlert
	if err := r.db.Preload("Product").
		Where("user_id = ?", userID).
		Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	return rows, nil
}

func (r *alertRepository) Save(alert *model.PriceAlert) error {
	if err := r.db.Save(alert).Error; err != nil {
		return fmt.Errorf("save alert: %w", err)
	}
	return nil
}

// ListActiveByProductTx 在事务内列出某建材所有生效订阅。
func ListActiveByProductTx(tx *gorm.DB, productID uint) ([]model.PriceAlert, error) {
	var rows []model.PriceAlert
	if err := tx.Where("product_id = ? AND status = ?", productID, constants.AlertStatusActive).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list active alerts by product: %w", err)
	}
	return rows, nil
}

// SaveAlertTx 在事务内更新订阅状态。
func SaveAlertTx(tx *gorm.DB, alert *model.PriceAlert) error {
	if err := tx.Save(alert).Error; err != nil {
		return fmt.Errorf("save alert in tx: %w", err)
	}
	return nil
}

// ProductNameTx 在事务内读取建材名称，用于组装通知文案。
func ProductNameTx(tx *gorm.DB, productID uint) (string, error) {
	var product model.Product
	if err := tx.Select("name").First(&product, productID).Error; err != nil {
		return "", fmt.Errorf("load product name: %w", err)
	}
	return product.Name, nil
}

// SupplierNameTx 在事务内读取供应商名称。
func SupplierNameTx(tx *gorm.DB, supplierID uint) (string, error) {
	var supplier model.Supplier
	if err := tx.Select("name").First(&supplier, supplierID).Error; err != nil {
		return "", fmt.Errorf("load supplier name: %w", err)
	}
	return supplier.Name, nil
}
