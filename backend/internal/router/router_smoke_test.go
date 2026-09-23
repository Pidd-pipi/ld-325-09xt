package router_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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

func startSmokeServer(t *testing.T) (*httptest.Server, map[string]string) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := model.Migrate(db); err != nil {
		t.Fatal(err)
	}
	if err := model.Seed(db); err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	secret := "smoke-secret"
	gin.SetMode(gin.ReleaseMode)
	server := httptest.NewServer(router.New(db, logger, secret))
	tokens := map[string]string{}
	for _, role := range []string{constants.RoleAdmin, constants.RoleSupplier, constants.RoleUser} {
		subject := constants.DemoUserID
		if role == constants.RoleAdmin {
			subject = constants.DemoAdminID
		}
		if role == constants.RoleSupplier {
			subject = constants.DemoSupplierPrefix + "1"
		}
		token, err := middleware.NewDemoToken(secret, subject, role)
		if err != nil {
			t.Fatal(err)
		}
		tokens[role] = token
	}
	return server, tokens
}

func doRequest(t *testing.T, server *httptest.Server, method, path, token string, body any) (envelope, int) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, server.URL+constants.APIPrefix+path, reader)
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var payload envelope
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode response %s: %v", string(raw), err)
	}
	return payload, resp.StatusCode
}

func TestPriceChangeReviewAndAlertEndToEnd(t *testing.T) {
	server, tokens := startSmokeServer(t)
	defer server.Close()

	// 1. 用户对商品 1（演示最低有货价 398）建立目标价 360 的提醒。
	payload, status := doRequest(t, server, http.MethodPost, "/alerts", tokens[constants.RoleUser], map[string]any{
		"product_id":   1,
		"target_price": 360,
	})
	if status != http.StatusOK || payload.Code != 0 {
		t.Fatalf("create alert failed: status=%d payload=%+v", status, payload)
	}

	// 2. 供应商对 offer 1（398 元）提交 350 元改价。
	payload, status = doRequest(t, server, http.MethodPost, "/supplier/offer-changes", tokens[constants.RoleSupplier], map[string]any{
		"offer_id":      1,
		"unit_price":    350,
		"freight":       "包邮",
		"delivery_days": 2,
		"stock_status":  "in_stock",
	})
	if status != http.StatusOK || payload.Code != 0 {
		t.Fatalf("submit change failed: status=%d payload=%+v", status, payload)
	}
	var change struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(payload.Data, &change); err != nil {
		t.Fatal(err)
	}

	// 3. 第二次提交被拒（409），不入库。
	_, conflict := doRequest(t, server, http.MethodPost, "/supplier/offer-changes", tokens[constants.RoleSupplier], map[string]any{
		"offer_id":      1,
		"unit_price":    340,
		"freight":       "包邮",
		"delivery_days": 2,
		"stock_status":  "in_stock",
	})
	if conflict != http.StatusConflict {
		t.Fatalf("expected 409 on duplicate submit, got %d", conflict)
	}

	// 4. 管理员审核通过。
	payload, status = doRequest(t, server, http.MethodPost,
		"/admin/offer-changes/"+jsonNumber(change.ID)+"/review", tokens[constants.RoleAdmin],
		map[string]any{"action": "approve"})
	if status != http.StatusOK || payload.Code != 0 {
		t.Fatalf("approve failed: status=%d payload=%+v", status, payload)
	}

	// 5. 提醒列表显示已触发、触发价与当前最低价。
	payload, status = doRequest(t, server, http.MethodGet, "/alerts", tokens[constants.RoleUser], nil)
	if status != http.StatusOK {
		t.Fatalf("list alerts failed: %d", status)
	}
	var alerts []struct {
		Status        string   `json:"status"`
		CurrentLowest float64  `json:"current_lowest"`
		TriggerPrice  *float64 `json:"trigger_price"`
		TriggeredAt   string   `json:"triggered_at"`
	}
	if err := json.Unmarshal(payload.Data, &alerts); err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 1 || alerts[0].Status != constants.AlertStatusTriggered {
		t.Fatalf("expected one triggered alert, got %+v", alerts)
	}
	if alerts[0].CurrentLowest != 350 || alerts[0].TriggerPrice == nil || *alerts[0].TriggerPrice != 360 {
		t.Fatalf("unexpected alert view: %+v", alerts[0])
	}
	if alerts[0].TriggeredAt == "" {
		t.Fatal("triggered_at must be recorded")
	}

	// 6. 站内通知已写入。
	payload, status = doRequest(t, server, http.MethodGet, "/notifications", tokens[constants.RoleUser], nil)
	if status != http.StatusOK {
		t.Fatalf("list notifications failed: %d", status)
	}
	var notes []map[string]any
	if err := json.Unmarshal(payload.Data, &notes); err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected one notification, got %d", len(notes))
	}

	// 7. 供应商可以查看审核结果。
	payload, status = doRequest(t, server, http.MethodGet, "/supplier/offer-changes", tokens[constants.RoleSupplier], nil)
	if status != http.StatusOK {
		t.Fatalf("supplier list failed: %d", status)
	}
	var supplierChanges []struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(payload.Data, &supplierChanges); err != nil {
		t.Fatal(err)
	}
	if len(supplierChanges) != 1 || supplierChanges[0].Status != constants.ChangeStatusApproved {
		t.Fatalf("expected one approved change visible to supplier, got %+v", supplierChanges)
	}

	// 8. 旧价已写入历史，趋势仍可用。
	_, status = doRequest(t, server, http.MethodGet, "/products/1/trend?range=30d", "", nil)
	if status != http.StatusOK {
		t.Fatalf("trend endpoint failed: %d", status)
	}
}

func TestStaleReviewReturnsConflictAndMarksStale(t *testing.T) {
	server, tokens := startSmokeServer(t)
	defer server.Close()

	payload, _ := doRequest(t, server, http.MethodPost, "/supplier/offer-changes", tokens[constants.RoleSupplier], map[string]any{
		"offer_id":      7,
		"unit_price":    70,
		"freight":       "包邮",
		"delivery_days": 3,
		"stock_status":  "in_stock",
	})
	var change struct {
		ID uint `json:"id"`
	}
	json.Unmarshal(payload.Data, &change)

	// 审核前报价被库存更新推进版本。
	if _, status := doRequest(t, server, http.MethodPatch, "/supplier/offers/7/status", tokens[constants.RoleSupplier],
		map[string]any{"stock_status": "out_of_stock"}); status != http.StatusOK {
		t.Fatalf("status update failed: %d", status)
	}

	_, status := doRequest(t, server, http.MethodPost,
		"/admin/offer-changes/"+jsonNumber(change.ID)+"/review", tokens[constants.RoleAdmin],
		map[string]any{"action": "approve"})
	if status != http.StatusConflict {
		t.Fatalf("expected 409 on stale review, got %d", status)
	}

	payload, _ = doRequest(t, server, http.MethodGet, "/supplier/offer-changes", tokens[constants.RoleSupplier], nil)
	var rows []struct {
		Status string `json:"status"`
	}
	json.Unmarshal(payload.Data, &rows)
	if len(rows) != 1 || rows[0].Status != constants.ChangeStatusStale {
		t.Fatalf("expected stale change, got %+v", rows)
	}
}

func TestExistingEntriesStillAvailable(t *testing.T) {
	server, _ := startSmokeServer(t)
	defer server.Close()

	// 比价、商品详情、供应商列表等原有入口继续可用。
	for _, path := range []string{"/products", "/products/1", "/suppliers", "/products/1/offers", "/products/1/trend?range=30d"} {
		_, status := doRequest(t, server, http.MethodGet, path, "", nil)
		if status != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", path, status)
		}
	}
	payload, status := doRequest(t, server, http.MethodPost, "/products/compare", "", map[string]any{"ids": []uint{1, 2}})
	if status != http.StatusOK || payload.Code != 0 {
		t.Fatalf("compare failed: %d %+v", status, payload)
	}
}

func jsonNumber(id uint) string {
	raw, _ := json.Marshal(id)
	return string(raw)
}
