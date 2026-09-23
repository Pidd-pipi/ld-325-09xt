package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/middleware"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/router"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func setupRouter(t *testing.T, secret string) (*gin.Engine, *gorm.DB, model.Offer) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := model.Migrate(db); err != nil {
		t.Fatal(err)
	}
	supplier := model.Supplier{Name: "集成测试商家", Status: constants.SupplierApproved}
	product := model.Product{Name: "集成测试岩板"}
	db.Create(&supplier)
	db.Create(&product)
	offer := model.Offer{ProductID: product.ID, SupplierID: supplier.ID, UnitPrice: 100, Freight: "包邮", DeliveryDays: 3, StockStatus: constants.StatusInStock, Version: 1}
	db.Create(&offer)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return router.New(db, logger, secret), db, offer
}

func perform(r *gin.Engine, method, path, token, body string) (int, envelope) {
	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var payload envelope
	_ = json.Unmarshal(rec.Body.Bytes(), &payload)
	return rec.Code, payload
}

func TestOfferChangeReviewAlertFlow(t *testing.T) {
	const secret = "integration-secret"
	r, _, offer := setupRouter(t, secret)

	supplierToken, _ := middleware.NewDemoToken(secret, "1", constants.RoleSupplier)
	adminToken, _ := middleware.NewDemoToken(secret, "admin-1", constants.RoleAdmin)

	// Supplier submits a revision.
	status, payload := perform(r, http.MethodPost,
		"/api/v1/supplier/offers/"+strconv.Itoa(int(offer.ID))+"/changes", supplierToken,
		`{"unit_price":80,"freight":"包邮到家","delivery_days":2,"stock_status":"in_stock"}`)
	if status != http.StatusOK || payload.Code != 0 {
		t.Fatalf("submit failed: %d %s", status, payload.Message)
	}

	// A second submission while pending is rejected with 409.
	status, _ = perform(r, http.MethodPost,
		"/api/v1/supplier/offers/"+strconv.Itoa(int(offer.ID))+"/changes", supplierToken,
		`{"unit_price":70,"freight":"包邮到家","delivery_days":1,"stock_status":"in_stock"}`)
	if status != http.StatusConflict {
		t.Fatalf("expected 409 for second pending change, got %d", status)
	}

	// User creates a target-price subscription before approval.
	status, payload = perform(r, http.MethodPost, "/api/v1/alerts", "",
		`{"product_id":`+strconv.Itoa(int(offer.ProductID))+`,"target_price":90,"drop_percent":0}`)
	if status != http.StatusOK || payload.Code != 0 {
		t.Fatalf("create alert failed: %d %s", status, payload.Message)
	}

	// Admin approves the pending revision.
	_, queue := perform(r, http.MethodGet, "/api/v1/admin/offer-changes?status=pending", adminToken, "")
	var changes []struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(queue.Data, &changes); err != nil || len(changes) != 1 {
		t.Fatalf("expected one pending change, got %s err %v", queue.Data, err)
	}
	status, review := perform(r, http.MethodPost,
		"/api/v1/admin/offer-changes/"+strconv.Itoa(int(changes[0].ID))+"/review", adminToken,
		`{"approve":true,"note":"同意"}`)
	if status != http.StatusOK || review.Code != 0 {
		t.Fatalf("review failed: %d %s", status, review.Message)
	}

	// The alert page shows the fired subscription.
	_, alertsPayload := perform(r, http.MethodGet, "/api/v1/alerts", "", "")
	var alerts []struct {
		Status         string  `json:"status"`
		TriggeredPrice float64 `json:"triggered_price"`
		CurrentLowest  float64 `json:"current_lowest"`
		TriggerPrice   float64 `json:"trigger_price"`
		TriggeredAt    *string `json:"triggered_at"`
	}
	if err := json.Unmarshal(alertsPayload.Data, &alerts); err != nil || len(alerts) != 1 {
		t.Fatalf("expected one alert, got %s err %v", alertsPayload.Data, err)
	}
	alert := alerts[0]
	if alert.Status != constants.AlertTriggered || alert.TriggeredPrice != 80 ||
		alert.CurrentLowest != 80 || alert.TriggerPrice != 90 || alert.TriggeredAt == nil {
		t.Fatalf("unexpected alert view: %+v", alert)
	}

	// Supplier can read the review result.
	_, mine := perform(r, http.MethodGet, "/api/v1/supplier/offer-changes", supplierToken, "")
	var mineRows []struct {
		Status     string `json:"status"`
		ReviewNote string `json:"review_note"`
	}
	if err := json.Unmarshal(mine.Data, &mineRows); err != nil || len(mineRows) != 1 {
		t.Fatalf("supplier change list wrong: %s err %v", mine.Data, err)
	}
	if mineRows[0].Status != constants.ChangeApproved || mineRows[0].ReviewNote != "同意" {
		t.Fatalf("unexpected review result: %+v", mineRows[0])
	}
}

func TestStaleReviewReturnsConflict(t *testing.T) {
	const secret = "integration-secret"
	r, db, offer := setupRouter(t, secret)
	supplierToken, _ := middleware.NewDemoToken(secret, "1", constants.RoleSupplier)
	adminToken, _ := middleware.NewDemoToken(secret, "admin-1", constants.RoleAdmin)

	_, _ = perform(r, http.MethodPost,
		"/api/v1/supplier/offers/"+strconv.Itoa(int(offer.ID))+"/changes", supplierToken,
		`{"unit_price":80,"freight":"包邮","delivery_days":2,"stock_status":"in_stock"}`)

	// Quote moves out of band after submission.
	db.Model(&model.Offer{}).Where("id = ?", offer.ID).
		Updates(map[string]any{"unit_price": 95, "version": 2})

	_, queue := perform(r, http.MethodGet, "/api/v1/admin/offer-changes?status=pending", adminToken, "")
	var changes []struct {
		ID uint `json:"id"`
	}
	_ = json.Unmarshal(queue.Data, &changes)
	status, payload := perform(r, http.MethodPost,
		"/api/v1/admin/offer-changes/"+strconv.Itoa(int(changes[0].ID))+"/review", adminToken,
		`{"approve":true}`)
	if status != http.StatusConflict || payload.Code != constants.ErrorConflict {
		t.Fatalf("expected 409 conflict, got %d code %d", status, payload.Code)
	}
	var change model.OfferChange
	db.First(&change, changes[0].ID)
	if change.Status != constants.ChangeStale {
		t.Fatalf("expected stale persisted, got %s", change.Status)
	}
}
