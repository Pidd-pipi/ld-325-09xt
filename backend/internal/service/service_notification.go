package service

import (
	"fmt"

	"github.com/blueship581/cybuildprice/backend/internal/dto"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
)

type NotificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) List(userID string) ([]dto.NotificationView, error) {
	rows, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("list notifications service: %w", err)
	}
	views := make([]dto.NotificationView, 0, len(rows))
	for _, row := range rows {
		views = append(views, toNotificationView(row))
	}
	return views, nil
}

func toNotificationView(row model.Notification) dto.NotificationView {
	return dto.NotificationView{
		ID:          row.ID,
		Title:       row.Title,
		Content:     row.Content,
		ProductID:   row.ProductID,
		Read:        row.Read,
		TriggeredAt: formatTime(row.TriggeredAt),
		CreatedAt:   formatTime(row.CreatedAt),
	}
}
