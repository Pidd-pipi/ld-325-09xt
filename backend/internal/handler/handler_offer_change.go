package handler

import (
	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// OfferChangeHandler 供应商提交报价修改、查看审核结果，管理员审核。
type OfferChangeHandler struct {
	service  *service.OfferChangeService
	validate *validator.Validate
}

func NewOfferChangeHandler(s *service.OfferChangeService, v *validator.Validate) *OfferChangeHandler {
	return &OfferChangeHandler{s, v}
}

// Submit 供应商提交新的单价、运费、货期和库存状态。
func (h *OfferChangeHandler) Submit(c *gin.Context) {
	supplierID, ok := service.ParseSupplierID(c.GetString(constants.UserIDContextKey))
	if !ok {
		c.Error(apperrors.ErrUnauthorized)
		return
	}
	var req dto.SubmitOfferChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	view, err := h.service.Submit(supplierID, req)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, view)
}

// ListMine 供应商查看自己的全部修改单与审核结果。
func (h *OfferChangeHandler) ListMine(c *gin.Context) {
	supplierID, ok := service.ParseSupplierID(c.GetString(constants.UserIDContextKey))
	if !ok {
		c.Error(apperrors.ErrUnauthorized)
		return
	}
	views, err := h.service.ListForSupplier(supplierID)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, views)
}

// ListPending 管理员查看待审队列。
func (h *OfferChangeHandler) ListPending(c *gin.Context) {
	views, err := h.service.ListPending()
	if err != nil {
		c.Error(err)
		return
	}
	success(c, views)
}

// Review 管理员审核修改单：版本已变化时返回 409，修改单标记失效。
func (h *OfferChangeHandler) Review(c *gin.Context) {
	var path struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&path); err != nil {
		c.Error(err)
		return
	}
	var req dto.ReviewOfferChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	view, err := h.service.Review(path.ID, req)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, view)
}
