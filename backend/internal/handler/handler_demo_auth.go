package handler

import (
	"time"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	"github.com/blueship581/cybuildprice/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// DemoAuthHandler 仅用于演示环境的角色切换，方便体验
// 供应商提交改价、管理员审核、用户接收预警的完整链路。
type DemoAuthHandler struct {
	secret   string
	validate *validator.Validate
}

func NewDemoAuthHandler(secret string, v *validator.Validate) *DemoAuthHandler {
	return &DemoAuthHandler{secret: secret, validate: v}
}

type demoIdentity struct {
	userID string
}

var demoIdentities = map[string]demoIdentity{
	constants.RoleAdmin:    {userID: constants.DemoAdminID},
	constants.RoleSupplier: {userID: constants.DemoSupplierPrefix + "1"},
	constants.RoleUser:     {userID: constants.DemoUserID},
}

func (h *DemoAuthHandler) Token(c *gin.Context) {
	var req dto.DemoTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	identity := demoIdentities[req.Role]
	token, err := middleware.NewDemoToken(h.secret, identity.userID, req.Role)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, dto.DemoTokenView{
		Role:      req.Role,
		UserID:    identity.userID,
		Token:     token,
		ExpiresIn: int(constants.DemoTokenLifetime / time.Second),
	})
}
