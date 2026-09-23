package service

import (
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	"github.com/blueship581/cybuildprice/backend/internal/model"
)

func TestAlertFiresOnceOnTargetPrice(t *testing.T) {
	db := newChangeTestDB(t)
	svc := newChangeService(db)
	userSvc := newUserDataService(db)
	product, offer := seedOffer(t, db, 100)

	if _, err := userSvc.Subscribe("user-1", dto.CreateAlertRequest{ProductID: product.ID, TargetPrice: 90}); err != nil {
		t.Fatal(err)
	}
	submitted, err := svc.Submit(offer.ID, 0, changeRequest(80))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Review(submitted.ID, dto.ReviewOfferChangeRequest{Approve: true}); err != nil {
		t.Fatal(err)
	}
	views, err := userSvc.Alerts("user-1")
	if err != nil || len(views) != 1 {
		t.Fatalf("unexpected alerts: %d %v", len(views), err)
	}
	view := views[0]
	if view.Status != constants.AlertTriggered || view.TriggeredPrice != 80 || view.TriggeredAt == nil {
		t.Fatalf("alert did not fire with price/time: %+v", view)
	}
	if view.CurrentLowest != 80 || view.TriggerPrice != 90 {
		t.Fatalf("alert view prices wrong: %+v", view)
	}
}

func TestAlertFiresOnDropPercent(t *testing.T) {
	db := newChangeTestDB(t)
	svc := newChangeService(db)
	userSvc := newUserDataService(db)
	product, offer := seedOffer(t, db, 100)

	if _, err := userSvc.Subscribe("user-2", dto.CreateAlertRequest{ProductID: product.ID, DropPercent: 10}); err != nil {
		t.Fatal(err)
	}
	// 95 is only a 5% drop: must stay active.
	first, err := svc.Submit(offer.ID, 0, changeRequest(95))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Review(first.ID, dto.ReviewOfferChangeRequest{Approve: true}); err != nil {
		t.Fatal(err)
	}
	views, _ := userSvc.Alerts("user-2")
	if views[0].Status != constants.AlertActive {
		t.Fatalf("5%% drop should not trigger 10%% alert: %+v", views[0])
	}
	// After approval a new submission is allowed; 88 is a 12% drop: must fire.
	second, err := svc.Submit(offer.ID, 0, changeRequest(88))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Review(second.ID, dto.ReviewOfferChangeRequest{Approve: true}); err != nil {
		t.Fatal(err)
	}
	views, _ = userSvc.Alerts("user-2")
	if views[0].Status != constants.AlertTriggered || views[0].TriggeredPrice != 88 {
		t.Fatalf("12%% drop should trigger at 88: %+v", views[0])
	}
}

func TestAlertDoesNotFireWhenAboveTarget(t *testing.T) {
	db := newChangeTestDB(t)
	svc := newChangeService(db)
	userSvc := newUserDataService(db)
	product, offer := seedOffer(t, db, 100)

	if _, err := userSvc.Subscribe("user-3", dto.CreateAlertRequest{ProductID: product.ID, TargetPrice: 50}); err != nil {
		t.Fatal(err)
	}
	submitted, err := svc.Submit(offer.ID, 0, changeRequest(70))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Review(submitted.ID, dto.ReviewOfferChangeRequest{Approve: true}); err != nil {
		t.Fatal(err)
	}
	views, _ := userSvc.Alerts("user-3")
	if views[0].Status != constants.AlertActive {
		t.Fatalf("price 70 above target 50 must stay active: %+v", views[0])
	}
}

func TestSubscribeIsIdempotentPerProduct(t *testing.T) {
	db := newChangeTestDB(t)
	userSvc := newUserDataService(db)
	product, _ := seedOffer(t, db, 100)

	req := dto.CreateAlertRequest{ProductID: product.ID, TargetPrice: 80}
	first, err := userSvc.Subscribe("user-4", req)
	if err != nil {
		t.Fatal(err)
	}
	second, err := userSvc.Subscribe("user-4", req)
	if err != nil || first.ID != second.ID {
		t.Fatalf("re-subscribe must return the single existing alert: %+v %+v %v", first, second, err)
	}
	var count int64
	db.Model(&model.PriceAlert{}).Where("user_id = ?", "user-4").Count(&count)
	if count != 1 {
		t.Fatalf("expected one subscription, got %d", count)
	}
}

func TestSubscribeRequiresCondition(t *testing.T) {
	db := newChangeTestDB(t)
	userSvc := newUserDataService(db)
	product, _ := seedOffer(t, db, 100)

	if _, err := userSvc.Subscribe("user-5", dto.CreateAlertRequest{ProductID: product.ID}); err == nil {
		t.Fatal("subscription without target or drop should fail validation")
	}
}
