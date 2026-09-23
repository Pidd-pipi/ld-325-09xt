package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"gorm.io/gorm"
)

type OfferChangeService struct {
	changes repository.OfferChangeRepository
	offers  repository.OfferRepository
	tx      *repository.TxManager
	logger  *slog.Logger
}

func NewOfferChangeService(
	changes repository.OfferChangeRepository,
	offers repository.OfferRepository,
	tx *repository.TxManager,
	logger *slog.Logger,
) *OfferChangeService {
	return &OfferChangeService{changes: changes, offers: offers, tx: tx, logger: logger}
}

// Submit 供应商提交一次报价修改。该报价已有待审修改时，后来的提交不入库并返回 409。
func (s *OfferChangeService) Submit(supplierID uint, input dto.SubmitOfferChangeRequest) (dto.OfferChangeView, error) {
	offer, err := s.offers.Get(input.OfferID)
	if err != nil {
		return dto.OfferChangeView{}, fmt.Errorf("load offer for change: %w", err)
	}
	if supplierID != offer.SupplierID {
		return dto.OfferChangeView{}, apperrors.NewBusinessValidation("只能修改本店铺的报价")
	}

	change := model.OfferChange{}
	err = s.tx.RunInTx(func(tx *gorm.DB) error {
		changeRepo := repository.NewOfferChangeRepositoryTx(tx)
		exists, err := changeRepo.PendingExistsByOffer(input.OfferID)
		if err != nil {
			return err
		}
		if exists {
			return apperrors.NewConflictError("该报价已有待审核修改，请等待审核结果后再提交")
		}
		change = model.OfferChange{
			OfferID:         offer.ID,
			ProductID:       offer.ProductID,
			SupplierID:      offer.SupplierID,
			BaseUnitPrice:   offer.UnitPrice,
			NewUnitPrice:    input.UnitPrice,
			NewFreight:      input.Freight,
			NewDeliveryDays: input.DeliveryDays,
			NewStockStatus:  input.StockStatus,
			BaseVersion:     offer.Version,
			Status:          constants.ChangeStatusPending,
		}
		if err := changeRepo.Create(&change); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return dto.OfferChangeView{}, err
	}

	stored, err := s.changes.Get(change.ID)
	if err != nil {
		return dto.OfferChangeView{}, fmt.Errorf("reload offer change: %w", err)
	}
	return toOfferChangeView(stored, offer.Version), nil
}

// Review 管理员审核修改单。
// 审核时报价版本已经变化，则把这次修改标记为失效并返回 409 冲突；
// 通过则在同一事务内改写报价、旧价写入历史、触发符合条件的预警。
func (s *OfferChangeService) Review(id uint, input dto.ReviewOfferChangeRequest) (dto.OfferChangeView, error) {
	if input.Action == constants.ChangeReviewReject {
		return s.reject(id, input.Remark)
	}
	return s.approve(id, input.Remark)
}

func (s *OfferChangeService) reject(id uint, remark string) (dto.OfferChangeView, error) {
	if err := s.changes.MarkStatus(id, constants.ChangeStatusRejected, remark); err != nil {
		return dto.OfferChangeView{}, err
	}
	stored, err := s.changes.Get(id)
	if err != nil {
		return dto.OfferChangeView{}, fmt.Errorf("reload rejected change: %w", err)
	}
	return toOfferChangeView(stored, stored.Offer.Version), nil
}

func (s *OfferChangeService) approve(id uint, remark string) (dto.OfferChangeView, error) {
	var staleConflict *apperrors.BusinessError
	err := s.tx.RunInTx(func(tx *gorm.DB) error {
		changeRepo := repository.NewOfferChangeRepositoryTx(tx)

		change, err := s.loadChangeForReview(tx, id)
		if err != nil {
			return err
		}
		offer, err := repository.FindOfferForUpdate(tx, change.OfferID)
		if err != nil {
			return err
		}

		// 乐观并发：报价已被其他修改/库存更新推进版本，本修改单失效。
		// 失效状态需要随事务提交，因此先正常落账，事务返回成功后再对外抛出冲突。
		if offer.Version != change.BaseVersion {
			now := time.Now()
			change.Status = constants.ChangeStatusStale
			change.ReviewRemark = remark
			change.ReviewedAt = &now
			if err := changeRepo.Save(&change); err != nil {
				return err
			}
			staleConflict = apperrors.NewConflictError("报价版本已变化，该修改已标记为失效，请基于最新报价重新提交")
			return nil
		}

		oldPrice := offer.UnitPrice
		if err := repository.ApplyChange(tx, &offer, change); err != nil {
			return err
		}
		history := model.PriceHistory{
			ProductID:  offer.ProductID,
			OfferID:    offer.ID,
			Price:      oldPrice,
			RecordedAt: time.Now(),
		}
		if err := repository.NewPriceHistoryRepositoryTx(tx).Create(&history); err != nil {
			return err
		}

		now := time.Now()
		change.Status = constants.ChangeStatusApproved
		change.ReviewRemark = remark
		change.ReviewedAt = &now
		if err := changeRepo.Save(&change); err != nil {
			return err
		}

		return s.fireAlerts(tx, offer, now)
	})
	if err != nil {
		return dto.OfferChangeView{}, err
	}
	if staleConflict != nil {
		s.logger.Info("offer change marked stale", "change_id", id)
		return dto.OfferChangeView{}, staleConflict
	}

	stored, err := s.changes.Get(id)
	if err != nil {
		return dto.OfferChangeView{}, fmt.Errorf("reload approved change: %w", err)
	}
	return toOfferChangeView(stored, stored.Offer.Version), nil
}

func (s *OfferChangeService) loadChangeForReview(tx *gorm.DB, id uint) (model.OfferChange, error) {
	var change model.OfferChange
	if err := tx.First(&change, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return change, apperrors.NewNotFoundError("报价修改单不存在")
		}
		return change, fmt.Errorf("lock offer change: %w", err)
	}
	if change.Status != constants.ChangeStatusPending {
		return change, apperrors.NewConflictError("该修改单已被处理，请勿重复审核")
	}
	return change, nil
}

// fireAlerts 在审核事务内触发所有满足条件的订阅：每条订阅只触发一次，
// 触发时记下触发价格与时间，并写入一条站内消息。
func (s *OfferChangeService) fireAlerts(tx *gorm.DB, offer model.Offer, firedAt time.Time) error {
	alerts, err := repository.ListActiveByProductTx(tx, offer.ProductID)
	if err != nil {
		return err
	}
	if len(alerts) == 0 {
		return nil
	}
	lowest, err := repository.LowestInStockByProductTx(tx, offer.ProductID)
	if err != nil {
		return err
	}
	productName, err := repository.ProductNameTx(tx, offer.ProductID)
	if err != nil {
		return err
	}
	supplierName, err := repository.SupplierNameTx(tx, offer.SupplierID)
	if err != nil {
		return err
	}
	for _, alert := range alerts {
		matched, triggerPrice := matchAlert(alert, lowest)
		if !matched {
			continue
		}
		alert.Status = constants.AlertStatusTriggered
		alert.TriggeredPrice = lowest
		alert.TriggeredAt = &firedAt
		if err := repository.SaveAlertTx(tx, &alert); err != nil {
			return err
		}
		notification := model.Notification{
			UserID:      alert.UserID,
			ProductID:   offer.ProductID,
			OfferID:     offer.ID,
			SupplierID:  offer.SupplierID,
			Title:       "价格提醒已触发：" + productName,
			Content:     buildNotificationContent(productName, supplierName, lowest, triggerPrice),
			TriggeredAt: firedAt,
		}
		if err := repository.CreateNotificationTx(tx, &notification); err != nil {
			return err
		}
		s.logger.Info("price alert fired", "alert_id", alert.ID, "product_id", offer.ProductID, "price", lowest)
	}
	return nil
}

// ListForSupplier 返回供应商可见的修改单及其审核结果。
func (s *OfferChangeService) ListForSupplier(supplierID uint) ([]dto.OfferChangeView, error) {
	rows, err := s.changes.ListBySupplier(supplierID)
	if err != nil {
		return nil, fmt.Errorf("list changes for supplier: %w", err)
	}
	return toOfferChangeViews(rows), nil
}

// ListPending 返回管理员审核队列。
func (s *OfferChangeService) ListPending() ([]dto.OfferChangeView, error) {
	rows, err := s.changes.ListByStatus(constants.ChangeStatusPending)
	if err != nil {
		return nil, fmt.Errorf("list pending changes: %w", err)
	}
	return toOfferChangeViews(rows), nil
}

// ParseSupplierID 从身份标识中解析供应商 ID（演示令牌 subject 使用 supplier-<id>）。
func ParseSupplierID(subject string) (uint, bool) {
	const prefix = constants.DemoSupplierPrefix
	if len(subject) <= len(prefix) || subject[:len(prefix)] != prefix {
		return 0, false
	}
	id, err := strconv.ParseUint(subject[len(prefix):], 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

func toOfferChangeViews(rows []model.OfferChange) []dto.OfferChangeView {
	views := make([]dto.OfferChangeView, 0, len(rows))
	for _, row := range rows {
		views = append(views, toOfferChangeView(row, row.Offer.Version))
	}
	return views
}

func toOfferChangeView(row model.OfferChange, currentVersion uint) dto.OfferChangeView {
	view := dto.OfferChangeView{
		ID:              row.ID,
		OfferID:         row.OfferID,
		ProductID:       row.ProductID,
		ProductName:     row.Product.Name,
		SupplierID:      row.SupplierID,
		SupplierName:    row.Supplier.Name,
		BaseUnitPrice:   row.BaseUnitPrice,
		NewUnitPrice:    row.NewUnitPrice,
		NewFreight:      row.NewFreight,
		NewDeliveryDays: row.NewDeliveryDays,
		NewStockStatus:  row.NewStockStatus,
		BaseVersion:     row.BaseVersion,
		CurrentVersion:  currentVersion,
		Status:          row.Status,
		ReviewRemark:    row.ReviewRemark,
		CreatedAt:       formatTime(row.CreatedAt),
	}
	if row.ReviewedAt != nil {
		view.ReviewedAt = formatTimePtr(row.ReviewedAt)
	}
	return view
}

func buildNotificationContent(productName, supplierName string, lowest, triggerPrice float64) string {
	return fmt.Sprintf("「%s」当前有货最低价 %.2f 元（%s），已达到你设置的触发价 %.2f 元。",
		productName, lowest, supplierName, triggerPrice)
}
