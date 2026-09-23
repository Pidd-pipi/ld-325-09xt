package service

import (
	"encoding/json"
	"fmt"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"gorm.io/gorm"
)

type UserDataService struct {
	repo      repository.UserDataRepository
	offerRepo repository.OfferRepository
	db        *gorm.DB
}

func NewUserDataService(db *gorm.DB, repo repository.UserDataRepository, offerRepo repository.OfferRepository) *UserDataService {
	return &UserDataService{db: db, repo: repo, offerRepo: offerRepo}
}

func (s *UserDataService) Favorite(user string, input dto.CreateFavoriteRequest) (model.Favorite, error) {
	return s.repo.CreateFavorite(model.Favorite{UserID: user, ProductID: input.ProductID, Folder: input.Folder})
}

func (s *UserDataService) Favorites(user string) ([]model.Favorite, error) {
	return s.repo.ListFavorites(user)
}

func (s *UserDataService) Budget(user string, input dto.BudgetRequest) (model.Budget, error) {
	rate := 350.0
	if input.RoomType == constants.BudgetRoomKitchen {
		rate = 580
	}
	if input.RoomType == constants.BudgetRoomBathroom {
		rate = 720
	}
	lineItems := map[string]float64{"materials": input.Area * rate, "allowance": input.Area * rate * 0.08}
	payload, err := json.Marshal(lineItems)
	if err != nil {
		return model.Budget{}, fmt.Errorf("encode budget payload: %w", err)
	}
	return s.repo.CreateBudget(model.Budget{UserID: user, RoomType: input.RoomType, Area: input.Area, Estimate: input.Area * rate * 1.08, Payload: string(payload)})
}
