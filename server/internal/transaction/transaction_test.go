package transaction

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPaymentHTTPRejectsInvalidInputBeforeService(t *testing.T) {
	mux := http.NewServeMux()
	NewHandler(nil, func(context.Context, string) (int, error) { return 1, nil }).Routes(mux)
	for _, tc := range []struct {
		body string
		auth bool
		want int
	}{
		{`{}`, false, 401}, {`null`, true, 400}, {`{"amount_cents":1} {}`, true, 400}, {`{"card_number":"123"}`, true, 400},
	} {
		r := httptest.NewRequest("POST", "/orders/1/transactions", strings.NewReader(tc.body))
		if tc.auth {
			r.Header.Set("Authorization", "Bearer token")
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Errorf("%s: got %d want %d", tc.body, w.Code, tc.want)
		}
	}
}
