package order

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOrderHTTPRejectsMalformedDraftAndMissingStaffAuth(t *testing.T) {
	mux := http.NewServeMux()
	NewHandler(nil, func(context.Context, string) (int, error) { return 0, nil }, "").Routes(mux)
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"POST", "/restaurants/cafe/branches/1/orders/local", `{"table_number":`, 400},
		{"POST", "/orders/1/items", `{"item_id":1,"quantity":2,"unit_price_cents":1}`, 400},
		{"POST", "/restaurants/cafe/branches/1/order-qr", `{}`, 401},
		{"POST", "/order-qr/invalid/open", "", 404},
	} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
		if w.Code != tc.status {
			t.Errorf("%s: got %d, want %d", tc.path, w.Code, tc.status)
		}
	}
}
