package repository

import (
	"fmt"
	"time"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
)

type UserDataRepository interface {
	CreateFavorite(value model.Favorite) (model.Favorite, error)
	ListFavorites(user string) ([]model.Favorite, error)
	CreateAlert(value model.PriceAlert) (model.PriceAlert, error)
	FindAlert(user string, productID uint) (model.PriceAlert, bool, error)
	ListAlerts(user string) ([]model.PriceAlert, error)
	TriggerAlerts(tx *gorm.DB, productID uint, lowestPrice float64) (int, error)
	CreateBudget(value model.Budget) (model.Budget, error)
}

type userDataRepository struct{ db *gorm.DB }

func NewUserDataRepository(db *gorm.DB) UserDataRepository { return &userDataRepository{db} }

func (r *userDataRepository) CreateFavorite(value model.Favorite) (model.Favorite, error) {
	if err := r.db.Create(&value).Error; err != nil {
		return value, fmt.Errorf("create favorite: %w", err)
	}
	return value, nil
}

func (r *userDataRepository) ListFavorites(user string) ([]model.Favorite, error) {
	var values []model.Favorite
	if err := r.db.Preload("Product").Where("user_id = ?", user).Find(&values).Error; err != nil {
		return nil, fmt.Errorf("list favorites: %w", err)
	}
	return values, nil
}

func (r *userDataRepository) CreateAlert(value model.PriceAlert) (model.PriceAlert, error) {
	if err := r.db.Create(&value).Error; err != nil {
		return value, fmt.Errorf("create alert: %w", err)
	}
	return value, nil
}

// FindAlert returns the user's single subscription for a product, if any.
func (r *userDataRepository) FindAlert(user string, productID uint) (model.PriceAlert, bool, error) {
	var alert model.PriceAlert
	err := r.db.Where("user_id = ? AND product_id = ?", user, productID).First(&alert).Error
	if err == gorm.ErrRecordNotFound {
		return alert, false, nil
	}
	if err != nil {
		return alert, false, fmt.Errorf("find alert: %w", err)
	}
	return alert, true, nil
}

func (r *userDataRepository) ListAlerts(user string) ([]model.PriceAlert, error) {
	var values []model.PriceAlert
	if err := r.db.Preload("Product").Where("user_id = ?", user).
		Order("created_at DESC").Find(&values).Error; err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	return values, nil
}

// TriggerAlerts fires at most once every still-active subscription of a product
// whose target price or drop-percent condition is met by lowestPrice.
func (r *userDataRepository) TriggerAlerts(tx *gorm.DB, productID uint, lowestPrice float64) (int, error) {
	var candidates []model.PriceAlert
	if err := tx.Where("product_id = ? AND status = ?", productID, constants.AlertActive).
		Find(&candidates).Error; err != nil {
		return 0, fmt.Errorf("list active alerts: %w", err)
	}
	fired := 0
	for _, alert := range candidates {
		meetsTarget := alert.TargetPrice > 0 && lowestPrice <= alert.TargetPrice
		meetsDrop := alert.BasePrice > 0 && alert.DropPercent > 0 &&
			lowestPrice <= alert.BasePrice*(1-alert.DropPercent/constants.PercentBase)
		if !meetsTarget && !meetsDrop {
			continue
		}
		if err := tx.Model(&model.PriceAlert{}).Where("id = ? AND status = ?", alert.ID, constants.AlertActive).
			Updates(map[string]any{
				"status":          constants.AlertTriggered,
				"triggered_price": lowestPrice,
				"triggered_at":    time.Now(),
			}).Error; err != nil {
			return fired, fmt.Errorf("trigger alert %d: %w", alert.ID, err)
		}
		fired++
	}
	return fired, nil
}

func (r *userDataRepository) CreateBudget(value model.Budget) (model.Budget, error) {
	if err := r.db.Create(&value).Error; err != nil {
		return value, fmt.Errorf("create budget: %w", err)
	}
	return value, nil
}
