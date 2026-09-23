package repository

import "gorm.io/gorm"

// InTransaction executes fn inside one database transaction. Services pass the
// transaction handle to repository methods that must commit atomically.
func InTransaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.Transaction(fn)
}
