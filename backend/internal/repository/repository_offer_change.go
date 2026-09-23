package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrPendingChangeExists = errors.New("pending offer change already exists")

type OfferChangeRepository interface {
	CreatePending(tx *gorm.DB, change model.OfferChange) (model.OfferChange, error)
	HasPending(offerID uint) (bool, error)
	List(filter OfferChangeFilter) ([]model.OfferChange, error)
	Get(id uint) (model.OfferChange, error)
	LockForUpdate(tx *gorm.DB, id uint) (model.OfferChange, error)
	MarkStale(tx *gorm.DB, id uint, note string) error
	MarkReviewed(tx *gorm.DB, id uint, status, note string) error
}

// OfferChangeFilter narrows change listings. Zero values mean no constraint.
type OfferChangeFilter struct {
	SupplierID uint
	Status     string
}

type offerChangeRepository struct{ db *gorm.DB }

func NewOfferChangeRepository(db *gorm.DB) OfferChangeRepository {
	return &offerChangeRepository{db}
}

func (r *offerChangeRepository) CreatePending(tx *gorm.DB, change model.OfferChange) (model.OfferChange, error) {
	change.Status = constants.ChangePending
	if err := tx.Create(&change).Error; err != nil {
		if isDuplicatePending(err) {
			return change, fmt.Errorf("create offer change: %w", ErrPendingChangeExists)
		}
		return change, fmt.Errorf("create offer change: %w", err)
	}
	return change, nil
}

func (r *offerChangeRepository) HasPending(offerID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.OfferChange{}).
		Where("offer_id = ? AND status = ?", offerID, constants.ChangePending).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("count pending changes: %w", err)
	}
	return count > 0, nil
}

func (r *offerChangeRepository) List(filter OfferChangeFilter) ([]model.OfferChange, error) {
	query := r.db.Model(&model.OfferChange{}).
		Preload("Offer").Preload("Supplier").Preload("Offer.Product")
	if filter.SupplierID > 0 {
		query = query.Where("supplier_id = ?", filter.SupplierID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	var rows []model.OfferChange
	if err := query.Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list offer changes: %w", err)
	}
	return rows, nil
}

func (r *offerChangeRepository) Get(id uint) (model.OfferChange, error) {
	var change model.OfferChange
	if err := r.db.Preload("Offer").Preload("Supplier").Preload("Offer.Product").
		First(&change, id).Error; err != nil {
		return change, fmt.Errorf("find offer change: %w", err)
	}
	return change, nil
}

func (r *offerChangeRepository) LockForUpdate(tx *gorm.DB, id uint) (model.OfferChange, error) {
	var change model.OfferChange
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&change, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return change, apperrors.ErrNotFound
		}
		return change, fmt.Errorf("lock offer change: %w", err)
	}
	return change, nil
}

func (r *offerChangeRepository) MarkStale(tx *gorm.DB, id uint, note string) error {
	return r.mark(tx, id, constants.ChangeStale, note)
}

func (r *offerChangeRepository) MarkReviewed(tx *gorm.DB, id uint, status, note string) error {
	return r.mark(tx, id, status, note)
}

func (r *offerChangeRepository) mark(tx *gorm.DB, id uint, status, note string) error {
	if err := tx.Model(&model.OfferChange{}).Where("id = ?", id).
		Updates(map[string]any{"status": status, "review_note": note}).Error; err != nil {
		return fmt.Errorf("mark offer change %s: %w", status, err)
	}
	return nil
}

// isDuplicatePending detects the partial unique index violation on PostgreSQL.
func isDuplicatePending(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") && strings.Contains(message, "offer_changes")
}
