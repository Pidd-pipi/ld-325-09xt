package router

import (
	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/handler"
	"github.com/blueship581/cybuildprice/backend/internal/middleware"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"github.com/blueship581/cybuildprice/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"log/slog"
)

func New(db *gorm.DB, logger *slog.Logger, jwtSecret string) *gin.Engine {
	v := validator.New()

	// 基础数据
	product := handler.NewProductHandler(service.NewProductService(repository.NewProductRepository(db), logger), v)
	offers := handler.NewOfferHandler(service.NewOfferService(repository.NewOfferRepository(db)), v)
	trend := handler.NewTrendHandler(service.NewPriceHistoryService(repository.NewPriceHistoryRepository(db)))
	user := handler.NewUserDataHandler(service.NewUserDataService(repository.NewUserDataRepository(db)), v)
	supplier := handler.NewSupplierHandler(service.NewSupplierService(repository.NewSupplierRepository(db)), v)

	// 价格预警与通知
	offerRepo := repository.NewOfferRepository(db)
	alertService := service.NewAlertService(repository.NewAlertRepository(db), offerRepo)
	alerts := handler.NewAlertHandler(alertService, v)
	notifications := handler.NewNotificationHandler(service.NewNotificationService(repository.NewNotificationRepository(db)))

	// 报价修改审核
	offerChangeService := service.NewOfferChangeService(
		repository.NewOfferChangeRepository(db),
		offerRepo,
		repository.NewTxManager(db),
		logger,
	)
	offerChanges := handler.NewOfferChangeHandler(offerChangeService, v)

	// 演示身份切换
	demoAuth := handler.NewDemoAuthHandler(jwtSecret, v)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestID(), middleware.RequestLogger(logger), middleware.ErrorHandler(), middleware.JWTOrDemoAuth(jwtSecret))
	r.GET(constants.HealthPath, func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	api := r.Group(constants.APIPrefix)
	api.POST("/auth/demo-token", demoAuth.Token)

	api.GET("/products", product.List)
	api.GET("/products/:id", product.Get)
	api.POST("/products/compare", product.Compare)
	api.GET("/products/:id/offers", offers.List)
	api.GET("/products/:id/trend", trend.Get)
	api.GET("/suppliers", supplier.List)

	// 管理员
	api.PATCH("/admin/suppliers/:id/status", middleware.RequireRole(constants.RoleAdmin), supplier.UpdateStatus)
	api.GET("/admin/offer-changes", middleware.RequireRole(constants.RoleAdmin), offerChanges.ListPending)
	api.POST("/admin/offer-changes/:id/review", middleware.RequireRole(constants.RoleAdmin), offerChanges.Review)

	// 供应商
	api.PATCH("/supplier/offers/:id/status", middleware.RequireRole(constants.RoleSupplier, constants.RoleAdmin), offers.UpdateStatus)
	api.GET("/supplier/offer-changes", middleware.RequireRole(constants.RoleSupplier, constants.RoleAdmin), offerChanges.ListMine)
	api.POST("/supplier/offer-changes", middleware.RequireRole(constants.RoleSupplier, constants.RoleAdmin), offerChanges.Submit)

	// 用户：收藏、预算、价格预警、站内消息
	api.GET("/favorites", user.ListFavorites)
	api.POST("/favorites", user.CreateFavorite)
	api.POST("/budgets", user.CreateBudget)
	api.GET("/alerts", alerts.List)
	api.POST("/alerts", alerts.Create)
	api.GET("/notifications", notifications.List)

	return r
}
