package service

import (
	"fmt"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
)

// Subscribe creates the single price alert for a user and product. Re-submitting
// returns the existing subscription instead of creating a duplicate. The current
// lowest quote is captured as the baseline for drop-percent alerts.
func (s *UserDataService) Subscribe(user string, input dto.CreateAlertRequest) (dto.AlertView, error) {
	if input.TargetPrice <= 0 && input.DropPercent <= 0 {
		return dto.AlertView{}, apperrors.NewBusinessError(constants.ErrorValidation, "目标价和降幅至少需要填写一项", apperrors.ErrInvalidInput)
	}
	existing, found, err := s.repo.FindAlert(user, input.ProductID)
	if err != nil {
		return dto.AlertView{}, fmt.Errorf("find existing alert: %w", err)
	}
	if found {
		return s.alertView(existing)
	}
	basePrice, err := s.offerRepo.MinUnitPrice(s.db, input.ProductID)
	if err != nil {
		return dto.AlertView{}, fmt.Errorf("capture base price: %w", err)
	}
	alert := model.PriceAlert{
		UserID:      user,
		ProductID:   input.ProductID,
		BasePrice:   basePrice,
		TargetPrice: input.TargetPrice,
		DropPercent: input.DropPercent,
		Status:      constants.AlertActive,
	}
	created, err := s.repo.CreateAlert(alert)
	if err != nil {
		return dto.AlertView{}, fmt.Errorf("create alert: %w", err)
	}
	return s.alertView(created)
}

// Alerts returns the user's subscriptions with the live lowest price, the price
// at which each subscription fires, and its status.
func (s *UserDataService) Alerts(user string) ([]dto.AlertView, error) {
	rows, err := s.repo.ListAlerts(user)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	views := make([]dto.AlertView, 0, len(rows))
	for _, row := range rows {
		view, err := s.alertView(row)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *UserDataService) alertView(alert model.PriceAlert) (dto.AlertView, error) {
	currentLowest, err := s.offerRepo.MinUnitPrice(s.db, alert.ProductID)
	if err != nil {
		return dto.AlertView{}, fmt.Errorf("current lowest price: %w", err)
	}
	view := dto.AlertView{
		ID:             alert.ID,
		ProductID:      alert.ProductID,
		BasePrice:      alert.BasePrice,
		TargetPrice:    alert.TargetPrice,
		DropPercent:    alert.DropPercent,
		CurrentLowest:  currentLowest,
		Status:         alert.Status,
		TriggeredPrice: alert.TriggeredPrice,
		TriggeredAt:    alert.TriggeredAt,
		CreatedAt:      alert.CreatedAt,
	}
	view.TriggerPrice = resolveTriggerPrice(alert)
	if alert.Product.ID != 0 {
		view.ProductName = alert.Product.Name
		view.Brand = alert.Product.Brand
		view.Model = alert.Product.ModelNumber
		view.Unit = alert.Product.Unit
	}
	return view, nil
}

// resolveTriggerPrice reports the threshold that fires the subscription: the
// explicit target price when set, otherwise the baseline discounted by the
// requested percentage. The price at which it actually fired is kept separate
// in TriggeredPrice.
func resolveTriggerPrice(alert model.PriceAlert) float64 {
	if alert.TargetPrice > 0 {
		return alert.TargetPrice
	}
	if alert.BasePrice > 0 && alert.DropPercent > 0 {
		return roundPrice(alert.BasePrice * (1 - alert.DropPercent/constants.PercentBase))
	}
	return 0
}

func roundPrice(value float64) float64 {
	return float64(int(value*constants.PriceRoundingFactor+0.5)) / constants.PriceRoundingFactor
}
