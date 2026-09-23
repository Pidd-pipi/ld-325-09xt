package model

import "gorm.io/gorm"

func AllModels() []any {
	return []any{
		&Category{},
		&Product{},
		&Supplier{},
		&Offer{},
		&OfferChange{},
		&PriceHistory{},
		&Favorite{},
		&PriceAlert{},
		&Notification{},
		&Budget{},
	}
}
func Migrate(db *gorm.DB) error { return db.AutoMigrate(AllModels()...) }
