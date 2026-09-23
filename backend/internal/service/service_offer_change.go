package service

import (
	"errors"
	"fmt"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"gorm.io/gorm"
)

// ErrChangeStale is returned when a review targets a revision whose quote has
// already moved to another version. The revision is marked stale first.
var ErrChangeStale = errors.New("offer change is stale")

type OfferChangeService struct {
	db           *gorm.DB
	offerRepo    repository.OfferRepository
	changeRepo   repository.OfferChangeRepository
	historyRepo  repository.PriceHistoryRepository
	userDataRepo repository.UserDataRepository
}

func NewOfferChangeService(db *gorm.DB, offerRepo repository.OfferRepository, changeRepo repository.OfferChangeRepository, historyRepo repository.PriceHistoryRepository, userDataRepo repository.UserDataRepository) *OfferChangeService {
	return &OfferChangeService{db: db, offerRepo: offerRepo, changeRepo: changeRepo, historyRepo: historyRepo, userDataRepo: userDataRepo}
}

// Submit stores a supplier revision. Only one pending revision per quote is
// allowed; later submissions are rejected until the first one is reviewed.
func (s *OfferChangeService) Submit(offerID, supplierID uint, req dto.SubmitOfferChangeRequest) (dto.OfferChangeView, error) {
	offer, err := s.offerRepo.Get(offerID)
	if err != nil {
		return dto.OfferChangeView{}, apperrors.NewBusinessError(constants.ErrorNotFound, "报价不存在", apperrors.ErrNotFound)
	}
	if supplierID > 0 && offer.SupplierID != supplierID {
		return dto.OfferChangeView{}, apperrors.NewBusinessError(constants.ErrorUnauthorized, "只能修改自家报价", apperrors.ErrUnauthorized)
	}
	pending, err := s.changeRepo.HasPending(offerID)
	if err != nil {
		return dto.OfferChangeView{}, fmt.Errorf("submit offer change: %w", err)
	}
	if pending {
		return dto.OfferChangeView{}, apperrors.Conflict("该报价已有待审核修改，请等待审核结果后再提交")
	}
	change := model.OfferChange{
		OfferID:         offerID,
		SupplierID:      offer.SupplierID,
		NewUnitPrice:    req.UnitPrice,
		NewFreight:      req.Freight,
		NewDeliveryDays: req.DeliveryDays,
		NewStockStatus:  req.StockStatus,
		BaseVersion:     offer.Version,
	}
	var created model.OfferChange
	err = repository.InTransaction(s.db, func(tx *gorm.DB) error {
		saved, createErr := s.changeRepo.CreatePending(tx, change)
		if createErr != nil {
			if errors.Is(createErr, repository.ErrPendingChangeExists) {
				return apperrors.Conflict("该报价已有待审核修改，请等待审核结果后再提交")
			}
			return fmt.Errorf("submit offer change: %w", createErr)
		}
		created = saved
		return nil
	})
	if err != nil {
		return dto.OfferChangeView{}, err
	}
	return s.toView(created, offer), nil
}
