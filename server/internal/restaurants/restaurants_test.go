package restaurants

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBranchAccessBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name  string
		op    operation
		scope access
		want  error
	}{
		{"assigned manager can update hours", setHours, access{manager: true}, nil},
		{"assigned manager cannot appoint owner", addOwner, access{manager: true}, errForbidden},
		{"organization manager cannot archive organization", archiveOrg, access{orgRole: "MANAGER"}, errForbidden},
		{"analyst cannot update branch", updateBranch, access{orgRole: "ANALYST"}, errForbidden},
		{"owner cannot approve branch", verifyBranch, access{orgRole: "OWNER"}, errForbidden},
		{"pending branch hidden from public", getBranch, access{}, errMissing},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := authorize(request{op: tc.op}, tc.scope); !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func TestRestaurantHTTPRejectsInvalidRequests(t *testing.T) {
	mux := http.NewServeMux()
	NewHandler(nil, nil).Routes(mux)
	for _, tc := range []struct {
		method, path string
		want         int
	}{
		{"POST", "/restaurants", 401}, {"GET", "/me/restaurants", 401}, {"GET", "/restaurants?page_size=101", 400}, {"GET", "/api/restaurants/store/branches/0", 400}, {"GET", "/restaurants", 503},
	} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.want {
			t.Errorf("%s %s: got %d, want %d", tc.method, tc.path, w.Code, tc.want)
		}
	}
}
