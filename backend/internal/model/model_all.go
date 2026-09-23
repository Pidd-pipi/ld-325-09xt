package model

import "gorm.io/gorm"

func AllModels() []any {
	return []any{&Category{}, &Product{}, &Supplier{}, &Offer{}, &OfferChange{}, &PriceHistory{}, &Favorite{}, &PriceAlert{}, &Budget{}}
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(AllModels()...); err != nil {
		return err
	}
	// At most one pending revision per quote. A partial unique index keeps the
	// constraint at the storage layer even under concurrent submissions.
	return db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_offer_change_one_pending
		ON offer_changes (offer_id) WHERE status = 'pending'`).Error
}
