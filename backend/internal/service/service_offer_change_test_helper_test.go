package service

import (
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newChangeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := model.Migrate(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func seedOffer(t *testing.T, db *gorm.DB, price float64) (model.Product, model.Offer) {
	t.Helper()
	supplier := model.Supplier{Name: "测试商家", Status: constants.SupplierApproved}
	product := model.Product{Name: "测试岩板"}
	if err := db.Create(&supplier).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}
	offer := model.Offer{ProductID: product.ID, SupplierID: supplier.ID, UnitPrice: price, Freight: "包邮", DeliveryDays: 3, StockStatus: constants.StatusInStock, Version: constants.OfferInitialVersion}
	if err := db.Create(&offer).Error; err != nil {
		t.Fatal(err)
	}
	return product, offer
}

func newChangeService(db *gorm.DB) *OfferChangeService {
	return NewOfferChangeService(
		db,
		repository.NewOfferRepository(db),
		repository.NewOfferChangeRepository(db),
		repository.NewPriceHistoryRepository(db),
		repository.NewUserDataRepository(db),
	)
}

func newUserDataService(db *gorm.DB) *UserDataService {
	return NewUserDataService(db, repository.NewUserDataRepository(db), repository.NewOfferRepository(db))
}

func changeRequest(price float64) dto.SubmitOfferChangeRequest {
	return dto.SubmitOfferChangeRequest{UnitPrice: price, Freight: "满 10 件包邮", DeliveryDays: 2, StockStatus: constants.StatusInStock}
}
