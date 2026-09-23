package router

import (
	"log/slog"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/handler"
	"github.com/blueship581/cybuildprice/backend/internal/middleware"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"github.com/blueship581/cybuildprice/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func New(db *gorm.DB, logger *slog.Logger, jwtSecret string) *gin.Engine {
	v := validator.New()
	productRepo := repository.NewProductRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	changeRepo := repository.NewOfferChangeRepository(db)
	historyRepo := repository.NewPriceHistoryRepository(db)
	userDataRepo := repository.NewUserDataRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)

	product := handler.NewProductHandler(service.NewProductService(productRepo, logger), v)
	offers := handler.NewOfferHandler(service.NewOfferService(offerRepo), v)
	change := handler.NewOfferChangeHandler(service.NewOfferChangeService(db, offerRepo, changeRepo, historyRepo, userDataRepo), v)
	trend := handler.NewTrendHandler(service.NewPriceHistoryService(historyRepo))
	user := handler.NewUserDataHandler(service.NewUserDataService(db, userDataRepo, offerRepo), v)
	supplier := handler.NewSupplierHandler(service.NewSupplierService(supplierRepo), v)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestID(), middleware.RequestLogger(logger), middleware.ErrorHandler(), middleware.JWTOrDemoAuth(jwtSecret))
	r.GET(constants.HealthPath, func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	api := r.Group(constants.APIPrefix)
	api.POST("/auth/demo-token", handler.DemoToken(jwtSecret))
	api.GET("/products", product.List)
	api.GET("/products/:id", product.Get)
	api.POST("/products/compare", product.Compare)
	api.GET("/products/:id/offers", offers.List)
	api.GET("/products/:id/trend", trend.Get)
	api.GET("/suppliers", supplier.List)
	api.PATCH("/admin/suppliers/:id/status", middleware.RequireRole(constants.RoleAdmin), supplier.UpdateStatus)
	api.GET("/supplier/offers", middleware.RequireRole(constants.RoleSupplier, constants.RoleAdmin), offers.ListMine)
	api.PATCH("/supplier/offers/:id/status", middleware.RequireRole(constants.RoleSupplier, constants.RoleAdmin), offers.UpdateStatus)
	api.POST("/supplier/offers/:id/changes", middleware.RequireRole(constants.RoleSupplier, constants.RoleAdmin), change.Submit)
	api.GET("/supplier/offer-changes", middleware.RequireRole(constants.RoleSupplier, constants.RoleAdmin), change.List)
	api.GET("/admin/offer-changes", middleware.RequireRole(constants.RoleAdmin), change.List)
	api.POST("/admin/offer-changes/:id/review", middleware.RequireRole(constants.RoleAdmin), change.Review)
	api.GET("/favorites", user.ListFavorites)
	api.POST("/favorites", user.CreateFavorite)
	api.GET("/alerts", user.ListAlerts)
	api.POST("/alerts", user.CreateAlert)
	api.POST("/budgets", user.CreateBudget)
	return r
}
