package handler

import (
	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	"github.com/blueship581/cybuildprice/backend/internal/middleware"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"github.com/blueship581/cybuildprice/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type OfferChangeHandler struct {
	service  *service.OfferChangeService
	validate *validator.Validate
}

func NewOfferChangeHandler(s *service.OfferChangeService, v *validator.Validate) *OfferChangeHandler {
	return &OfferChangeHandler{s, v}
}

// Submit handles a supplier revision for one of their quotes.
func (h *OfferChangeHandler) Submit(c *gin.Context) {
	var path struct {
		ID uint `uri:"id" binding:"required"`
	}
	var req dto.SubmitOfferChangeRequest
	if err := c.ShouldBindUri(&path); err != nil {
		c.Error(err)
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	view, err := h.service.Submit(path.ID, middleware.SupplierID(c), req)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, view)
}

// List serves the review queue for admins and a supplier's own results.
func (h *OfferChangeHandler) List(c *gin.Context) {
	filter := repository.OfferChangeFilter{Status: c.Query("status")}
	if c.GetString(constants.RoleContextKey) == constants.RoleSupplier {
		filter.SupplierID = middleware.SupplierID(c)
	}
	rows, err := h.service.List(filter)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, rows)
}

// Review approves or rejects one pending revision (admin only).
func (h *OfferChangeHandler) Review(c *gin.Context) {
	var path struct {
		ID uint `uri:"id" binding:"required"`
	}
	var req dto.ReviewOfferChangeRequest
	if err := c.ShouldBindUri(&path); err != nil {
		c.Error(err)
		return
	}
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
