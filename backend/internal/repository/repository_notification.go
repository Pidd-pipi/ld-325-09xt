package repository

import (
	"fmt"

	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	ListByUser(userID string) ([]model.Notification, error)
	Create(notification *model.Notification) error
}

type notificationRepository struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db}
}

func (r *notificationRepository) ListByUser(userID string) ([]model.Notification, error) {
	var rows []model.Notification
	if err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	return rows, nil
}

func (r *notificationRepository) Create(notification *model.Notification) error {
	if err := r.db.Create(notification).Error; err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

// CreateNotificationTx 在审核事务内写入站内消息。
func CreateNotificationTx(tx *gorm.DB, notification *model.Notification) error {
	if err := tx.Create(notification).Error; err != nil {
		return fmt.Errorf("create notification in tx: %w", err)
	}
	return nil
}
