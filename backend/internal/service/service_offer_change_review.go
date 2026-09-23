package service

import (
	"fmt"
	"time"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"gorm.io/gorm"
)

// Review approves or rejects a pending revision. When the quote version moved
// after the revision was submitted, the revision is marked stale (committed) and
// a conflict is returned. On approval the old price is archived and matching
// subscriptions fire once.
func (s *OfferChangeService) Review(id uint, req dto.ReviewOfferChangeRequest) (dto.OfferChangeView, error) {
	status := constants.ChangeRejected
	if req.Approve {
		status = constants.ChangeApproved
	}
	staleConflict := (*apperrors.BusinessError)(nil)
	err := repository.InTransaction(s.db, func(tx *gorm.DB) error {
		change, err := s.changeRepo.LockForUpdate(tx, id)
		if err != nil {
			return apperrors.NewBusinessError(constants.ErrorNotFound, "修改记录不存在", apperrors.ErrNotFound)
		}
		if change.Status != constants.ChangePending {
			return apperrors.Conflict("该修改已审核，请勿重复操作")
		}
		offer, err := s.offerRepo.LockForUpdate(tx, change.OfferID)
		if err != nil {
			return apperrors.NewBusinessError(constants.ErrorNotFound, "报价不存在", apperrors.ErrNotFound)
		}
		if offer.Version != change.BaseVersion {
			note := fmt.Sprintf("报价版本已变更为 v%d，本次修改失效", offer.Version)
			if err := s.changeRepo.MarkStale(tx, change.ID, note); err != nil {
				return err
			}
			// Commit the stale marking, then surface the conflict outside the
			// transaction so the status change is not rolled back.
			staleConflict = apperrors.NewBusinessError(constants.ErrorConflict,
				fmt.Sprintf("报价版本已变化（当前 v%d / 提交基于 v%d），本次修改已失效", offer.Version, change.BaseVersion),
				ErrChangeStale)
			return nil
		}
		if req.Approve {
			historyRow := model.PriceHistory{
				ProductID:    offer.ProductID,
				OfferID:      offer.ID,
				Price:        offer.UnitPrice,
				OldUnitPrice: offer.UnitPrice,
				ChangeSource: constants.PriceSourceRevision,
				RecordedAt:   time.Now(),
			}
			if err := s.historyRepo.RecordRevision(tx, historyRow); err != nil {
				return err
			}
			offer.UnitPrice = change.NewUnitPrice
			offer.Freight = change.NewFreight
			offer.DeliveryDays = change.NewDeliveryDays
			offer.StockStatus = change.NewStockStatus
			offer.Version++
			if err := s.offerRepo.ApplyRevision(tx, offer); err != nil {
				return err
			}
			lowest, err := s.offerRepo.MinUnitPrice(tx, offer.ProductID)
			if err != nil {
				return err
			}
			if _, err := s.userDataRepo.TriggerAlerts(tx, offer.ProductID, lowest); err != nil {
				return err
			}
		}
		if err := s.changeRepo.MarkReviewed(tx, change.ID, status, req.Note); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return dto.OfferChangeView{}, err
	}
	if staleConflict != nil {
		return dto.OfferChangeView{}, staleConflict
	}
	full, err := s.changeRepo.Get(id)
	if err != nil {
		return dto.OfferChangeView{}, fmt.Errorf("reload reviewed change: %w", err)
	}
	return s.toView(full, full.Offer), nil
}
