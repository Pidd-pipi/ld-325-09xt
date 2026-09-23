package service

import (
	"errors"
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
)

func TestSubmitRejectsSecondPendingChange(t *testing.T) {
	db := newChangeTestDB(t)
	svc := newChangeService(db)
	_, offer := seedOffer(t, db, 100)

	if _, err := svc.Submit(offer.ID, 0, changeRequest(90)); err != nil {
		t.Fatalf("first submit failed: %v", err)
	}
	_, err := svc.Submit(offer.ID, 0, changeRequest(80))
	if err == nil {
		t.Fatal("second submit while pending should be rejected")
	}
	var business *apperrors.BusinessError
	if !errors.As(err, &business) || business.Code != constants.ErrorConflict {
		t.Fatalf("expected conflict business error, got %v", err)
	}
}

func TestReviewApproveArchivesOldPriceAndBumpsVersion(t *testing.T) {
	db := newChangeTestDB(t)
	svc := newChangeService(db)
	product, offer := seedOffer(t, db, 100)

	submitted, err := svc.Submit(offer.ID, 0, changeRequest(80))
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	view, err := svc.Review(submitted.ID, dto.ReviewOfferChangeRequest{Approve: true})
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	if view.Status != constants.ChangeApproved || view.OfferVersion != constants.OfferInitialVersion+1 || view.CurrentPrice != 80 {
		t.Fatalf("unexpected approved view: %+v", view)
	}
	var updated model.Offer
	if err := db.First(&updated, offer.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updated.UnitPrice != 80 || updated.Version != constants.OfferInitialVersion+1 {
		t.Fatalf("offer not updated: %+v", updated)
	}
	var history []model.PriceHistory
	if err := db.Where("offer_id = ? AND change_source = ?", offer.ID, constants.PriceSourceRevision).Find(&history).Error; err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Price != 100 || history[0].OldUnitPrice != 100 {
		t.Fatalf("old price not archived: %+v", history)
	}
	_ = product
}

func TestReviewStaleWhenOfferVersionMoved(t *testing.T) {
	db := newChangeTestDB(t)
	svc := newChangeService(db)
	_, offer := seedOffer(t, db, 100)

	submitted, err := svc.Submit(offer.ID, 0, changeRequest(90))
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	// Simulate the quote changing through another approved revision.
	if err := db.Model(&model.Offer{}).Where("id = ?", offer.ID).
		Updates(map[string]any{"unit_price": 95, "version": offer.Version + 1}).Error; err != nil {
		t.Fatal(err)
	}
	_, err = svc.Review(submitted.ID, dto.ReviewOfferChangeRequest{Approve: true})
	if err == nil {
		t.Fatal("review of stale change should conflict")
	}
	var business *apperrors.BusinessError
	if !errors.As(err, &business) || business.Code != constants.ErrorConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	var change model.OfferChange
	if err := db.First(&change, submitted.ID).Error; err != nil {
		t.Fatal(err)
	}
	if change.Status != constants.ChangeStale {
		t.Fatalf("expected stale status, got %s", change.Status)
	}
	var live model.Offer
	if err := db.First(&live, offer.ID).Error; err != nil {
		t.Fatal(err)
	}
	if live.UnitPrice != 95 {
		t.Fatal("stale review must not modify the live offer")
	}
}

func TestReviewRejectKeepsOfferUntouched(t *testing.T) {
	db := newChangeTestDB(t)
	svc := newChangeService(db)
	_, offer := seedOffer(t, db, 100)

	submitted, err := svc.Submit(offer.ID, 0, changeRequest(80))
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	view, err := svc.Review(submitted.ID, dto.ReviewOfferChangeRequest{Approve: false, Note: "资料不全"})
	if err != nil {
		t.Fatalf("reject failed: %v", err)
	}
	if view.Status != constants.ChangeRejected {
		t.Fatalf("expected rejected, got %s", view.Status)
	}
	var live model.Offer
	if err := db.First(&live, offer.ID).Error; err != nil {
		t.Fatal(err)
	}
	if live.UnitPrice != 100 || live.Version != constants.OfferInitialVersion {
		t.Fatalf("rejected change must not alter offer: %+v", live)
	}
	var count int64
	db.Model(&model.PriceHistory{}).Where("change_source = ?", constants.PriceSourceRevision).Count(&count)
	if count != 0 {
		t.Fatalf("rejected change must not write history: %d rows", count)
	}
}

func TestSubmitRejectsOtherSupplierOffer(t *testing.T) {
	db := newChangeTestDB(t)
	svc := newChangeService(db)
	_, offer := seedOffer(t, db, 100)

	_, err := svc.Submit(offer.ID, offer.SupplierID+999, changeRequest(90))
	var business *apperrors.BusinessError
	if !errors.As(err, &business) || business.Code != constants.ErrorUnauthorized {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestSubmitMissingOfferReturnsNotFound(t *testing.T) {
	db := newChangeTestDB(t)
	svc := newChangeService(db)

	_, err := svc.Submit(9999, 0, changeRequest(90))
	var business *apperrors.BusinessError
	if !errors.As(err, &business) || business.Code != constants.ErrorNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestReviewAllowsNewSubmissionAfterApproval(t *testing.T) {
	db := newChangeTestDB(t)
	svc := newChangeService(db)
	_, offer := seedOffer(t, db, 100)

	first, err := svc.Submit(offer.ID, 0, changeRequest(90))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Review(first.ID, dto.ReviewOfferChangeRequest{Approve: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Submit(offer.ID, 0, changeRequest(85)); err != nil {
		t.Fatalf("new submission after approval should succeed: %v", err)
	}
	rows, err := svc.List(repository.OfferChangeFilter{})
	if err != nil || len(rows) != 2 {
		t.Fatalf("expected two change records, got %d err %v", len(rows), err)
	}
}
