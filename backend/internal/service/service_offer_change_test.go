package service

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type offerChangeFixture struct {
	db               *gorm.DB
	changes          *OfferChangeService
	alerts           *AlertService
	notifications    *NotificationService
	offerRepo        repository.OfferRepository
	changeRepo       repository.OfferChangeRepository
	alertRepo        repository.AlertRepository
	notificationRepo repository.NotificationRepository
	product          model.Product
	supplier         model.Supplier
	offer            model.Offer
}

func setupOfferChangeFixture(t *testing.T) offerChangeFixture {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := model.Migrate(db); err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	offerRepo := repository.NewOfferRepository(db)
	changeRepo := repository.NewOfferChangeRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	category := model.Category{Name: "瓷砖"}
	if err := db.Create(&category).Error; err != nil {
		t.Fatal(err)
	}
	product := model.Product{Name: "测试岩板", CategoryID: category.ID}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}
	supplier := model.Supplier{Name: "测试商家", Status: constants.SupplierApproved}
	if err := db.Create(&supplier).Error; err != nil {
		t.Fatal(err)
	}
	offer := model.Offer{
		ProductID:    product.ID,
		SupplierID:   supplier.ID,
		UnitPrice:    100,
		MOQ:          1,
		Freight:      "包邮",
		DeliveryDays: 3,
		StockStatus:  constants.StatusInStock,
		Version:      constants.OfferInitialVersion,
	}
	if err := db.Create(&offer).Error; err != nil {
		t.Fatal(err)
	}

	changes := NewOfferChangeService(changeRepo, offerRepo, repository.NewTxManager(db), logger)
	alerts := NewAlertService(alertRepo, offerRepo)
	notifications := NewNotificationService(notificationRepo)

	return offerChangeFixture{
		db: db, changes: changes, alerts: alerts, notifications: notifications,
		offerRepo: offerRepo, changeRepo: changeRepo,
		alertRepo: alertRepo, notificationRepo: notificationRepo,
		product: product, supplier: supplier, offer: offer,
	}
}

func submitChangeRequest(f offerChangeFixture, price float64) dto.SubmitOfferChangeRequest {
	return dto.SubmitOfferChangeRequest{
		OfferID:      f.offer.ID,
		UnitPrice:    price,
		Freight:      "包邮",
		DeliveryDays: 2,
		StockStatus:  constants.StatusInStock,
	}
}

func TestSubmitBlocksSecondPendingChange(t *testing.T) {
	f := setupOfferChangeFixture(t)

	if _, err := f.changes.Submit(f.supplier.ID, submitChangeRequest(f, 90)); err != nil {
		t.Fatalf("first submit should succeed: %v", err)
	}
	_, err := f.changes.Submit(f.supplier.ID, submitChangeRequest(f, 80))
	if err == nil {
		t.Fatal("second submit while pending must be rejected")
	}
	var business *apperrors.BusinessError
	if !errors.As(err, &business) || !errors.Is(business, apperrors.ErrConflict) {
		t.Fatalf("expected conflict error, got %T %v", err, err)
	}
	var count int64
	if err := f.db.Model(&model.OfferChange{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("later submission must not be stored, got %d rows", count)
	}
}

func TestReviewApproveWritesHistoryAndFiresAlertOnce(t *testing.T) {
	f := setupOfferChangeFixture(t)

	// 用户以目标价 90 建立预警，基线价为 100。
	alertView, err := f.alerts.Create("buyer-1", dto.CreateAlertRequest{
		ProductID:   f.product.ID,
		TargetPrice: 90,
	})
	if err != nil {
		t.Fatalf("create alert: %v", err)
	}

	view, err := f.changes.Submit(f.supplier.ID, submitChangeRequest(f, 88))
	if err != nil {
		t.Fatalf("submit change: %v", err)
	}
	reviewed, err := f.changes.Review(view.ID, dto.ReviewOfferChangeRequest{Action: constants.ChangeReviewApprove})
	if err != nil {
		t.Fatalf("approve change: %v", err)
	}
	if reviewed.Status != constants.ChangeStatusApproved {
		t.Fatalf("expected approved, got %s", reviewed.Status)
	}

	// 报价已被改写，版本推进。
	updated, err := f.offerRepo.Get(f.offer.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.UnitPrice != 88 || updated.Version != constants.OfferInitialVersion+1 {
		t.Fatalf("offer not applied: price=%v version=%v", updated.UnitPrice, updated.Version)
	}

	// 旧价 100 已写入历史。
	var histories []model.PriceHistory
	if err := f.db.Where("offer_id = ?", f.offer.ID).Find(&histories).Error; err != nil {
		t.Fatal(err)
	}
	if len(histories) != 1 || histories[0].Price != 100 {
		t.Fatalf("expected old price 100 in history, got %+v", histories)
	}

	// 预警只触发一次，记录触发价与时间；站内消息写入。
	alerts, err := f.alertRepo.ListByUser("buyer-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 1 || alerts[0].Status != constants.AlertStatusTriggered {
		t.Fatalf("alert not triggered: %+v", alerts)
	}
	if alerts[0].TriggeredPrice != 88 || alerts[0].TriggeredAt == nil {
		t.Fatalf("trigger record missing: price=%v at=%v", alerts[0].TriggeredPrice, alerts[0].TriggeredAt)
	}
	messages, err := f.notifications.List("buyer-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected one notification, got %d", len(messages))
	}
	if alertView.ID == 0 {
		t.Fatal("alert view missing id")
	}
}

func TestReviewStaleWhenVersionChanged(t *testing.T) {
	f := setupOfferChangeFixture(t)

	view, err := f.changes.Submit(f.supplier.ID, submitChangeRequest(f, 88))
	if err != nil {
		t.Fatalf("submit change: %v", err)
	}

	// 审核前，报价版本被另一处变更（库存状态更新）推进。
	if _, err := f.offerRepo.UpdateStatus(f.offer.ID, constants.StatusOutOfStock); err != nil {
		t.Fatal(err)
	}

	_, err = f.changes.Review(view.ID, dto.ReviewOfferChangeRequest{Action: constants.ChangeReviewApprove})
	var business *apperrors.BusinessError
	if !errors.As(err, &business) || !errors.Is(business, apperrors.ErrConflict) {
		t.Fatalf("expected conflict on stale review, got %v", err)
	}

	change, err := f.changeRepo.Get(view.ID)
	if err != nil {
		t.Fatal(err)
	}
	if change.Status != constants.ChangeStatusStale {
		t.Fatalf("expected stale status, got %s", change.Status)
	}

	// 失效修改单没有改写报价。
	offer, err := f.offerRepo.Get(f.offer.ID)
	if err != nil {
		t.Fatal(err)
	}
	if offer.UnitPrice != 100 {
		t.Fatalf("stale change must not modify price, got %v", offer.UnitPrice)
	}
}

func TestApproveDoesNotFireWhenAboveTrigger(t *testing.T) {
	f := setupOfferChangeFixture(t)

	if _, err := f.alerts.Create("buyer-1", dto.CreateAlertRequest{
		ProductID:   f.product.ID,
		TargetPrice: 80,
	}); err != nil {
		t.Fatalf("create alert: %v", err)
	}
	view, err := f.changes.Submit(f.supplier.ID, submitChangeRequest(f, 95))
	if err != nil {
		t.Fatalf("submit change: %v", err)
	}
	if _, err := f.changes.Review(view.ID, dto.ReviewOfferChangeRequest{Action: constants.ChangeReviewApprove}); err != nil {
		t.Fatalf("approve: %v", err)
	}
	alerts, err := f.alertRepo.ListByUser("buyer-1")
	if err != nil {
		t.Fatal(err)
	}
	if alerts[0].Status != constants.AlertStatusActive {
		t.Fatalf("alert must remain active, got %s", alerts[0].Status)
	}
	messages, err := f.notifications.List("buyer-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 0 {
		t.Fatalf("no notification expected, got %d", len(messages))
	}
}

func TestReviewRejectKeepsOfferUnchanged(t *testing.T) {
	f := setupOfferChangeFixture(t)
	view, err := f.changes.Submit(f.supplier.ID, submitChangeRequest(f, 88))
	if err != nil {
		t.Fatalf("submit change: %v", err)
	}
	reviewed, err := f.changes.Review(view.ID, dto.ReviewOfferChangeRequest{
		Action: constants.ChangeReviewReject,
		Remark: "价格异常",
	})
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
	if reviewed.Status != constants.ChangeStatusRejected || reviewed.ReviewRemark != "价格异常" {
		t.Fatalf("unexpected review result: %+v", reviewed)
	}
	offer, err := f.offerRepo.Get(f.offer.ID)
	if err != nil {
		t.Fatal(err)
	}
	if offer.UnitPrice != 100 || offer.Version != constants.OfferInitialVersion {
		t.Fatalf("rejected change must not modify offer: price=%v version=%v", offer.UnitPrice, offer.Version)
	}
}
