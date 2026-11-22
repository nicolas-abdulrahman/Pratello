package menu

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pratello/server/internal/user"
)

type stubService func(context.Context, request) (any, error)

func (f stubService) execute(ctx context.Context, r request) (any, error) { return f(ctx, r) }

func TestMenuHTTPBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, body, token string
		want                            int
	}{
		{"missing auth", "POST", "/items", `{}`, "", 401},
		{"bad JSON", "POST", "/items", `{"name":`, "token", 400},
		{"trailing data", "POST", "/items", `{} {}`, "token", 400},
		{"too large", "PUT", "/menu", `{"name":"` + strings.Repeat("x", bodyLimit) + `"}`, "token", 413},
		{"invalid tenant resource ID", "GET", "/items/2147483648", "", "", 400},
		{"invalid pagination", "GET", "/items?page_size=101", "", "", 400},
		{"ambiguous branch", "GET", "/items?branch_id=1&branch_id=2", "", "", 400},
		{"missing item", "GET", "/items/99", "", "", 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := Handler{service: stubService(func(_ context.Context, r request) (any, error) {
				if tc.want != 404 || r.op != getItem || r.item != 99 {
					t.Fatalf("unexpected business call: %+v", r)
				}
				return nil, ErrMissing
			}), identify: func(context.Context, string) (int, error) { return 3, nil }}
			mux := http.NewServeMux()
			h.Routes(mux)
			r := httptest.NewRequest(tc.method, "/restaurants/test"+tc.path, strings.NewReader(tc.body))
			if tc.token != "" {
				r.Header.Set("Authorization", "Bearer "+tc.token)
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d; %s", w.Code, tc.want, w.Body.String())
			}
			var response struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || response.Error == "" {
				t.Fatalf("expected API error envelope, got %s", w.Body.String())
			}
		})
	}
}

type fakeStore struct{ session }

func (f fakeStore) transact(_ context.Context, _ string, _ bool, fn func(session) (any, error)) (any, error) {
	return fn(f.session)
}

type fakeSession struct {
	session
	role      user.Scope
	published bool
	catalog   catalog
}

func (f fakeSession) scope(context.Context, int, int) (user.Scope, error) { return f.role, nil }
func (f fakeSession) publication(context.Context, int) (bool, error)      { return f.published, nil }
func (f fakeSession) load(context.Context, int) (catalog, error)          { return f.catalog, nil }

func TestMenuPermissions(t *testing.T) {
	cost := int64(500)
	c := catalog{Items: map[int]item{1: {ID: 1, Cost: &cost}}}
	for _, tc := range []struct {
		name               string
		role               user.Scope
		published, private bool
		want               error
	}{
		{"customer unverified", user.Scope{}, false, false, ErrMissing},
		{"customer verified", user.Scope{}, true, false, nil},
		{"employee previews", user.Scope{Employee: true}, false, false, nil},
		{"manager previews", user.Scope{Manager: true}, false, false, nil},
		{"owner previews", user.Scope{Organization: "OWNER"}, false, true, nil},
		{"analyst previews", user.Scope{Organization: "ANALYST"}, false, true, nil},
		{"administrator previews", user.Scope{Administrator: true}, false, true, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newService(fakeStore{fakeSession{role: tc.role, published: tc.published, catalog: c}})
			out, err := s.execute(context.Background(), request{op: getItem, slug: "test", item: 1, branch: 2})
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
			if err == nil && (out.(itemDetail).Cost != nil) != tc.private {
				t.Fatal("wrong visibility of production cost")
			}
		})
	}
}

func TestSelectionPricingAndSnapshot(t *testing.T) {
	at := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	branch := 2
	c := catalog{
		Items:      map[int]item{1: {ID: 1, Name: "Combo", Base: 2000, Combo: true, Available: true}, 2: {ID: 2, Name: "Burger", Available: true}},
		Components: map[int][]component{1: {{ID: 5, Item: 2, Name: "Burger", Quantity: 2}}},
		Recipes:    map[int][]recipe{2: {{ID: 7, Name: "Onions", Removable: true}, {ID: 8, Name: "Bun"}}},
		Sides:      map[int][]side{1: {{ID: 4, Title: "Drink", Min: 1, Max: 1, Options: []option{{ID: 10, Name: "Juice", Price: 200, Available: true}, {ID: 11, Name: "Water", Available: true}}}}},
		Sales: map[int][]sale{1: {{ID: 1, Type: "fixed_cents", Value: 800, Start: at, End: at.Add(time.Hour)},
			{ID: 2, Branch: &branch, Type: "fixed_cents", Value: 500, Start: at, End: at.Add(time.Hour)},
			{ID: 3, Branch: &branch, Type: "percent", Value: 10, Start: at, End: at.Add(time.Hour)}}},
		OptionStock: map[int]bool{}, ItemStock: map[int]bool{},
	}
	sel := Selection{ItemID: 1, Quantity: 2, OptionIDs: []int{10}, ComponentRemovals: []ComponentRemoval{{ComboItemID: 5, Unit: 2, RemoveIngredientIDs: []int{7}}}, Notes: "To go"}
	first, err := c.snapshot(2, sel, at)
	if err != nil {
		t.Fatal(err)
	}
	if first.ItemName != "Combo" || first.UnitPriceCents != 2000 || first.LineTotalCents != 4000 || len(first.Options) != 1 || first.Options[0].Group != "Drink" || first.Notes != "To go\nBurger #2: without Onions" {
		t.Fatalf("unexpected snapshot: %+v", first)
	}
	if andElsewhere, err := c.snapshot(3, sel, at); err != nil || andElsewhere.UnitPriceCents != 1400 {
		t.Fatalf("organization price: %+v %v", andElsewhere, err)
	}
	if later, err := c.snapshot(2, sel, at.Add(time.Hour)); err != nil || later.UnitPriceCents != 2200 {
		t.Fatalf("end exclusive: %+v %v", later, err)
	}
	for _, tc := range []struct {
		name string
		ids  []int
		want error
	}{
		{"mandatory", nil, ErrInvalid}, {"over max", []int{10, 11}, ErrInvalid}, {"duplicate", []int{10, 10}, ErrInvalid}, {"foreign option", []int{10, 99}, ErrInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			choice := sel
			choice.OptionIDs = tc.ids
			if _, err := c.snapshot(2, choice, at); !errors.Is(err, tc.want) {
				t.Fatal(err)
			}
		})
	}
	clone := sel
	clone.ComponentRemovals = []ComponentRemoval{{ComboItemID: 5, Unit: 1, RemoveIngredientIDs: []int{8}}}
	if _, err := c.snapshot(2, clone, at); !errors.Is(err, ErrInvalid) {
		t.Fatalf("non-removable ingredient: %v", err)
	}
	c.OptionStock[10] = false
	if _, err := c.snapshot(2, sel, at); !errors.Is(err, ErrNotOrderable) {
		t.Fatalf("disabled option: %v", err)
	}
	c.OptionStock[11] = false
	if c.view(1, 2, false, at).Available {
		t.Fatal("mandatory group cannot be satisfied")
	}
	delete(c.OptionStock, 10)
	c.ItemStock[2] = false
	if c.view(1, 2, false, at).Available {
		t.Fatal("required component is out of stock")
	}
	delete(c.ItemStock, 2)
	c.Items[1] = item{ID: 1, Name: "Edited", Available: true, Base: math.MaxInt64}
	c.Sides[1][0].Title = "Renamed"
	if _, err := c.snapshot(2, sel, at.Add(time.Hour)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("overflow: %v", err)
	}
	if first.ItemName != "Combo" || first.Options[0].Group != "Drink" || first.UnitPriceCents != 2000 {
		t.Fatal("previous snapshot was modified")
	}
}

func selectionGroupCatalog() catalog {
	return catalog{
		Items: map[int]item{1: {ID: 1, Name: "Lunch", Base: 1200, Available: true}},
		Sides: map[int][]side{1: {
			{Title: "Extras", Max: 2, Options: []option{{ID: 1, Price: 250, Available: true}}},
			{Title: "Drink", Min: 1, Max: 1, Options: []option{{ID: 2, Available: true}, {ID: 3, Available: true}}},
			{Title: "Sides", Min: 3, Max: 3, Options: []option{{ID: 4, Available: true}, {ID: 5, Available: true}, {ID: 6, Price: 75, Available: true}, {ID: 7, Available: true}}},
		}},
	}
}

func TestCAT050SelectionCountsAreIndependentPerGroup(t *testing.T) {
	at := time.Now()
	c := selectionGroupCatalog()
	for _, tc := range []struct {
		name  string
		ids   []int
		price int64
	}{
		{"included only", []int{2, 4, 5, 7}, 1200},
		{"paid extras in two groups", []int{1, 2, 4, 5, 6}, 1525},
		{"extra cannot make up missing side", []int{1, 2, 4, 5}, 0},
		{"second drink cannot make up missing side", []int{2, 3, 4, 5}, 0},
		{"too many sides", []int{2, 4, 5, 6, 7}, 0},
		{"repeat cannot make up missing side", []int{2, 4, 5, 5}, 0},
		{"foreign option", []int{2, 4, 5, 9}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			line, err := c.snapshot(2, Selection{ItemID: 1, Quantity: 2, OptionIDs: tc.ids}, at)
			if tc.price == 0 {
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("accepted invalid selection: %+v, %v", line, err)
				}
				return
			}
			if err != nil || line.UnitPriceCents != tc.price || line.LineTotalCents != tc.price*2 {
				t.Fatalf("wrong line: %+v, %v", line, err)
			}
		})
	}
}

func TestCAT051BranchOptionStockControlsAvailability(t *testing.T) {
	at := time.Now()
	c := selectionGroupCatalog()
	branch := c
	branch.OptionStock = map[int]bool{6: false}
	if !branch.view(1, 2, false, at).Available {
		t.Fatal("three remaining sides should be sufficient")
	}
	if _, err := branch.snapshot(2, Selection{ItemID: 1, Quantity: 1, OptionIDs: []int{2, 4, 5, 6}}, at); !errors.Is(err, ErrNotOrderable) {
		t.Fatalf("out-of-stock option accepted: %v", err)
	}
	branch.OptionStock[7] = false
	if branch.view(1, 2, false, at).Available || !c.view(1, 3, false, at).Available {
		t.Fatal("mandatory minimum or branch isolation violated")
	}
}

func TestCAT053FutureSaleDoesNotOverrideEarly(t *testing.T) {
	at := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	branch := 2
	v := item{ID: 1, Base: 2000}
	c := catalog{Sales: map[int][]sale{1: {
		{ID: 1, Type: "fixed_cents", Value: 800, Start: at, End: at.Add(4 * time.Hour)},
		{ID: 2, Branch: &branch, Type: "fixed_cents", Value: 400, Start: at, End: at.Add(3 * time.Hour)},
		{ID: 3, Branch: &branch, Type: "fixed_cents", Value: 100, Start: at.Add(time.Hour), End: at.Add(2 * time.Hour)},
	}}}
	for _, tc := range []struct {
		elapsed time.Duration
		branch  int
		price   int64
	}{
		{-time.Nanosecond, branch, 2000}, {0, branch, 1600}, {time.Hour, branch, 1900}, {2 * time.Hour, branch, 1600},
		{3 * time.Hour, branch, 1200}, {4 * time.Hour, branch, 2000}, {time.Hour, 3, 1200},
	} {
		if price := c.effective(v, tc.branch, at.Add(tc.elapsed)); price != tc.price {
			t.Errorf("branch %d after %s: got %d want %d", tc.branch, tc.elapsed, price, tc.price)
		}
	}
}
