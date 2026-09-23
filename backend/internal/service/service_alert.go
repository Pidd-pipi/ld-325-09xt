package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
)

type AlertService struct {
	repo   repository.AlertRepository
	offers repository.OfferRepository
}

func NewAlertService(repo repository.AlertRepository, offers repository.OfferRepository) *AlertService {
	return &AlertService{repo: repo, offers: offers}
}

// Create 为用户建立一条价格预警。同一用户对同一建材只保留一条有效订阅：
// 已存在生效订阅时直接返回冲突，避免重复建单后降价无结果。
func (s *AlertService) Create(userID string, input dto.CreateAlertRequest) (dto.AlertView, error) {
	if input.TargetPrice <= 0 && input.DropPercent <= 0 {
		return dto.AlertView{}, apperrors.NewBusinessValidation("目标价与降幅百分比至少需要设置一项")
	}
	_, err := s.repo.FindActiveByUserProduct(userID, input.ProductID)
	switch {
	case err == nil:
		return dto.AlertView{}, apperrors.NewBusinessValidation("该建材已存在生效中的价格预警")
	case !errors.Is(err, apperrors.ErrNotFound):
		return dto.AlertView{}, fmt.Errorf("check existing alert: %w", err)
	}

	baseline, err := s.offers.LowestInStockByProduct(input.ProductID)
	if err != nil {
		return dto.AlertView{}, fmt.Errorf("load baseline price: %w", err)
	}
	if baseline <= 0 {
		return dto.AlertView{}, apperrors.NewBusinessValidation("该建材暂无有货报价，暂无法建立价格预警")
	}

	target := 0.0
	if input.TargetPrice > 0 {
		target = input.TargetPrice
	}
	alert := model.PriceAlert{
		UserID:        userID,
		ProductID:     input.ProductID,
		TargetPrice:   target,
		DropPercent:   input.DropPercent,
		BaselinePrice: baseline,
		Status:        constants.AlertStatusActive,
	}
	if err := s.repo.Create(&alert); err != nil {
		return dto.AlertView{}, fmt.Errorf("create alert service: %w", err)
	}
	stored, err := s.repo.Get(alert.ID)
	if err != nil {
		return dto.AlertView{}, fmt.Errorf("reload created alert: %w", err)
	}
	return s.view(stored, baseline)
}

// List 返回用户的全部订阅视图，附带当前最低价与触发价。
func (s *AlertService) List(userID string) ([]dto.AlertView, error) {
	rows, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("list alerts service: %w", err)
	}
	views := make([]dto.AlertView, 0, len(rows))
	for _, row := range rows {
		lowest, err := s.offers.LowestInStockByProduct(row.ProductID)
		if err != nil {
			return nil, fmt.Errorf("load current lowest: %w", err)
		}
		view, err := s.view(row, lowest)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *AlertService) view(alert model.PriceAlert, currentLowest float64) (dto.AlertView, error) {
	view := dto.AlertView{
		ID:             alert.ID,
		ProductID:      alert.ProductID,
		ProductName:    alert.Product.Name,
		BaselinePrice:  alert.BaselinePrice,
		CurrentLowest:  currentLowest,
		Status:         alert.Status,
		TriggeredPrice: alert.TriggeredPrice,
		CreatedAt:      formatTime(alert.CreatedAt),
	}
	if alert.TargetPrice > 0 {
		target := alert.TargetPrice
		view.TargetPrice = &target
	}
	if alert.DropPercent > 0 {
		drop := alert.DropPercent
		view.DropPercent = &drop
	}
	if threshold := effectiveTriggerPrice(alert); threshold != nil {
		trigger := *threshold
		view.TriggerPrice = &trigger
	}
	if alert.TriggeredAt != nil {
		view.TriggeredAt = formatTimePtr(alert.TriggeredAt)
	}
	return view, nil
}

func formatTime(t time.Time) string {
	return t.Format(constants.TimeDisplayLayout)
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(constants.TimeDisplayLayout)
}
