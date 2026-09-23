package handler

import (
	"github.com/blueship581/cybuildprice/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// NotificationHandler 处理站内降价提醒消息。
type NotificationHandler struct{ service *service.NotificationService }

func NewNotificationHandler(s *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{s}
}

func (h *NotificationHandler) List(c *gin.Context) {
	views, err := h.service.List(c.GetString("user_id"))
	if err != nil {
		c.Error(err)
		return
	}
	success(c, views)
}
