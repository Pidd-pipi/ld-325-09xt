package handler

import (
	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

type demoTokenRequest struct {
	Subject string `json:"subject" binding:"required"`
	Role    string `json:"role" binding:"required,oneof=admin supplier user"`
}

// DemoToken issues a short-lived JWT for the demo deployment. The platform has
// no password directory; reviewers use it to enter the supplier and admin views.
func DemoToken(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req demoTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		token, err := middleware.NewDemoToken(secret, req.Subject, req.Role)
		if err != nil {
			c.Error(err)
			return
		}
		success(c, gin.H{"token": token, "role": req.Role, "subject": req.Subject, "lifetime_hours": int(constants.DemoTokenLifetime.Hours())})
	}
}
