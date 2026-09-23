package service

import (
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
)

func TestAlertCreateRequiresCondition(t *testing.T) {
	f := setupOfferChangeFixture(t)
	_, err := f.alerts.Create("buyer-1", dto.CreateAlertRequest{ProductID: f.product.ID})
	if err == nil {
		t.Fatal("alert without target or drop must be rejected")
	}
}

func TestAlertCreateRejectsDuplicate(t *testing.T) {
	f := setupOfferChangeFixture(t)
	input := dto.CreateAlertRequest{ProductID: f.product.ID, TargetPrice: 90}
	if _, err := f.alerts.Create("buyer-1", input); err != nil {
		t.Fatalf("first alert: %v", err)
	}
	if _, err := f.alerts.Create("buyer-1", input); err == nil {
		t.Fatal("duplicate active alert must be rejected")
	}
}

func TestAlertListViewContainsCurrentLowestAndTriggerPrice(t *testing.T) {
	f := setupOfferChangeFixture(t)
	if _, err := f.alerts.Create("buyer-1", dto.CreateAlertRequest{
		ProductID: f.product.ID, TargetPrice: 95, DropPercent: 10,
	}); err != nil {
		t.Fatalf("create alert: %v", err)
	}
	views, err := f.alerts.List("buyer-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 {
		t.Fatalf("expected one view, got %d", len(views))
	}
	view := views[0]
	if view.CurrentLowest != 100 {
		t.Fatalf("expected current lowest 100, got %v", view.CurrentLowest)
	}
	if view.BaselinePrice != 100 {
		t.Fatalf("expected baseline 100, got %v", view.BaselinePrice)
	}
	if view.TriggerPrice == nil || *view.TriggerPrice != 90 {
		t.Fatalf("expected trigger price 90 (min of target 95 and 10%% drop 90), got %v", view.TriggerPrice)
	}
	if view.Status != constants.AlertStatusActive {
		t.Fatalf("unexpected status %s", view.Status)
	}
}
