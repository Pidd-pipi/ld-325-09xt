package service

import (
	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
)

// List returns quote revisions, optionally scoped to one supplier or status, so
// suppliers can follow their own review results and admins can see the queue.
func (s *OfferChangeService) List(filter repository.OfferChangeFilter) ([]dto.OfferChangeView, error) {
	rows, err := s.changeRepo.List(filter)
	if err != nil {
		return nil, err
	}
	views := make([]dto.OfferChangeView, 0, len(rows))
	for _, row := range rows {
		views = append(views, s.toView(row, row.Offer))
	}
	return views, nil
}

func (s *OfferChangeService) toView(change model.OfferChange, offer model.Offer) dto.OfferChangeView {
	view := dto.OfferChangeView{
		ID:              change.ID,
		OfferID:         change.OfferID,
		ProductID:       offer.ProductID,
		SupplierID:      change.SupplierID,
		CurrentPrice:    offer.UnitPrice,
		NewUnitPrice:    change.NewUnitPrice,
		CurrentFreight:  offer.Freight,
		NewFreight:      change.NewFreight,
		CurrentDelivery: offer.DeliveryDays,
		NewDeliveryDays: change.NewDeliveryDays,
		CurrentStock:    offer.StockStatus,
		NewStockStatus:  change.NewStockStatus,
		BaseVersion:     change.BaseVersion,
		OfferVersion:    offer.Version,
		Status:          change.Status,
		ReviewNote:      change.ReviewNote,
		CreatedAt:       change.CreatedAt,
	}
	if change.Status != constants.ChangePending {
		reviewed := change.UpdatedAt
		view.ReviewedAt = &reviewed
	}
	if change.Offer.Product.ID != 0 {
		view.ProductName = change.Offer.Product.Name
	}
	if change.Supplier.ID != 0 {
		view.SupplierName = change.Supplier.Name
	}
	return view
}
