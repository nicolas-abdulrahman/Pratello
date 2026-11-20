package review

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReviewHTTPValidation(t *testing.T) {
	mux := http.NewServeMux()
	NewHandler(&Service{}, func(context.Context, string) (int, error) { return 1, nil }).Routes(mux)
	for _, tc := range []struct {
		path, body, credential string
		want                   int
	}{
		{"/orders/1/review", `{"rating":5}`, "", 401},
		{"/orders/1/review", `{"comment":"rating missing"}`, "user", 400},
		{"/orders/1/review", `{"rating":6}`, "user", 400},
		{"/orders/1/items/1/review", `{"rating":4}`, "user", 400},
		{"/orders/1/items/1/review", `{"rating":2,"comment":"text"}`, "guest", 403},
		{"/orders/1/items/1/review", `{"rating":2,"item_id":7}`, "user", 400},
		{"/orders/1/review", `{"rating":5} {}`, "user", 400},
	} {
		r := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
		if tc.credential == "user" {
			r.Header.Set("Authorization", "Bearer user")
		}
		if tc.credential == "guest" {
			r.Header.Set("X-Order-Token", "guest")
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Errorf("%s: %d want %d", tc.body, w.Code, tc.want)
		}
	}
}
