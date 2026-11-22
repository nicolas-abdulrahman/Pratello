package review

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"pratello/server/internal/db"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("AUTH_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL for PostgreSQL")
	}
	root, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	schema := fmt.Sprintf("review_test_%d", time.Now().UnixNano())
	if _, err = root.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err = root.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Error(err)
		}
	})
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, e := url.Parse(dsn)
		if e != nil {
			t.Fatal(e)
		}
		q := u.Query()
		q.Set("search_path", schema)
		u.RawQuery = q.Encode()
		dsn = u.String()
	} else {
		dsn += " search_path=" + schema
	}
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err = db.Migrate(conn); err != nil {
		t.Fatal(err)
	}
	return conn
}

type reviewFixture struct {
	t     *testing.T
	db    *sql.DB
	guest string
	mux   *http.ServeMux
}

func setupReviews(t *testing.T) reviewFixture {
	t.Helper()
	c := testDB(t)
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := c.Exec(q, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO restaurant(name_slug,name) VALUES('cafe','Cafe'),('other','Other');
	INSERT INTO account(email,password) VALUES('buyer@test.local','hash'),('other@test.local','hash');
	INSERT INTO "Address"("CEP",street,number) VALUES('12345678','Rua','1');
	INSERT INTO physical_restaurant(restaurant_id,address_id,name,cnpj,owner_cnpj,cnpj_verified) VALUES(1,1,'Cafe','12345678000100','12345678000100',true),(1,1,'Cafe 2','12345678000200','12345678000100',true);
	INSERT INTO category(restaurant_id,name) VALUES(1,'Lunch');
	INSERT INTO menu_item(restaurant_id,category_id,name,base_price_cents,is_combo,allows_review) VALUES(1,1,'Burger',1500,false,true),(1,1,'Combo',2500,true,true),(1,1,'Hidden',100,false,false);
	INSERT INTO "order"(physical_restaurant_id,account_id,order_type,status) VALUES(1,1,'local','delivered'),(2,1,'local','delivered'),(2,NULL,'local','delivered'),(1,1,'local','pending'),(1,1,'local','cancelled');
	INSERT INTO "order"(physical_restaurant_id,account_id,order_type,submitted_at) VALUES(1,1,'local',NULL);
	INSERT INTO order_item(order_id,item_id,item_name,unit_price_cents,line_total_cents,quantity) VALUES
	(1,1,'Burger',1500,3000,2),(1,2,'Combo',2500,2500,1),(2,1,'Burger',1500,1500,1),(3,1,'Burger',1500,1500,1),
	(4,1,'Burger',1500,3000,2),(5,1,'Burger',1500,4500,3),(6,1,'Burger',1500,6000,4),
	(2,3,'Hidden',100,100,1),(2,1,'Burger',1500,3000,2),(1,1,'Component',0,0,1);
	UPDATE order_item SET voided_at=now() WHERE id=9;
	UPDATE order_item SET parent_order_item_id=2 WHERE id=10;`)
	guest := strings.Repeat("a", 43)
	hash := sha256.Sum256([]byte(guest))
	exec(`UPDATE "order" SET guest_token_hash=$1 WHERE id=3`, hash[:])
	mux := http.NewServeMux()
	NewHandler(NewService(c), func(_ context.Context, key string) (int, error) {
		if key == "buyer" {
			return 1, nil
		}
		return 2, nil
	}).Routes(mux)
	return reviewFixture{t, c, guest, mux}
}

func (f reviewFixture) run(method, path, key, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	switch key {
	case "buyer", "other":
		r.Header.Set("Authorization", "Bearer "+key)
	case "":
	default:
		r.Header.Set("X-Order-Token", key)
	}
	w := httptest.NewRecorder()
	f.mux.ServeHTTP(w, r)
	return w
}

func (f reviewFixture) call(method, path, key, body string, want int) *httptest.ResponseRecorder {
	f.t.Helper()
	w := f.run(method, path, key, body)
	if w.Code != want {
		f.t.Fatalf("%s %s: got %d want %d: %s", method, path, w.Code, want, w.Body.String())
	}
	return w
}
func (f reviewFixture) post(path, key, body string, want int) {
	f.t.Helper()
	f.call("POST", path, key, body, want)
}

func (f reviewFixture) get(path string) page {
	f.t.Helper()
	w := f.call("GET", path, "", "", 200)
	var p page
	if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
		f.t.Fatal(err)
	}
	for _, private := range []string{`"order_item_id"`, `"order_id"`, `"account_id"`, `"guest_token_hash"`, f.guest} {
		if strings.Contains(w.Body.String(), private) {
			f.t.Fatalf("private data in review response: %s", private)
		}
	}
	return p
}
func (f reviewFixture) exec(q string) {
	f.t.Helper()
	if _, err := f.db.Exec(q); err != nil {
		f.t.Fatal(err)
	}
}

func (f reviewFixture) assertSummary(path string, count, sum int64) page {
	f.t.Helper()
	p := f.get(path)
	if p.Summary.Count != count || p.Summary.Sum != sum || (count == 0 && p.Summary.Average != nil) || (count > 0 && (p.Summary.Average == nil || *p.Summary.Average != float64(sum)/float64(count))) {
		f.t.Fatalf("wrong totals for %s: %+v, want count=%d sum=%d", path, p.Summary, count, sum)
	}
	return p
}

const (
	burger             = "/api/restaurants/cafe/items/1/reviews"
	combo              = "/restaurants/cafe/items/2/reviews"
	org                = "/restaurants/cafe/reviews"
	itemPurchase       = "/orders/1/items/1/review"
	guestPurchase      = "/orders/3/items/4/review"
	restaurantPurchase = "/orders/1/review"
	guestRestaurant    = "/orders/3/review"
)

func TestREV001RatingScales(t *testing.T) {
	t.Parallel()
	f := setupReviews(t)
	for _, path := range []string{itemPurchase, restaurantPurchase} {
		for _, body := range []string{`{"comment":"no rating"}`, `{"rating":0}`} {
			f.post(path, "buyer", body, 400)
		}
	}
	f.post(itemPurchase, "buyer", `{"rating":4}`, 400)
	f.post(restaurantPurchase, "buyer", `{"rating":6}`, 400)
	f.assertSummary(burger, 0, 0)
	f.assertSummary(org, 0, 0)
	f.post(itemPurchase, "buyer", `{"rating":3}`, 200)
	f.post("/orders/2/items/3/review", "buyer", `{"rating":1}`, 200)
	f.post(restaurantPurchase, "buyer", `{"rating":5}`, 200)
	f.post(guestRestaurant, f.guest, `{"rating":1}`, 200)
	f.assertSummary(burger, 2, 4)
	f.assertSummary(org, 2, 6)
}

func TestREV002PurchaseCredentials(t *testing.T) {
	t.Parallel()
	f := setupReviews(t)
	f.post(itemPurchase, "", `{"rating":3}`, 401)
	f.post(itemPurchase, "other", `{"rating":3}`, 403)
	f.post(itemPurchase, f.guest, `{"rating":3}`, 403)
	f.post(guestPurchase, strings.Repeat("b", 43), `{"rating":3}`, 403)
	f.post("/orders/1/items/3/review", "buyer", `{"rating":3}`, 404)
	f.assertSummary(burger, 0, 0)
}

func TestREV003IneligiblePurchases(t *testing.T) {
	t.Parallel()
	f := setupReviews(t)
	// Give the draft a delivered status to test that submission, not only status, is required.
	f.exec(`UPDATE "order" SET status='delivered' WHERE id=6`)
	for _, path := range []string{"/orders/4/items/5/review", "/orders/5/items/6/review", "/orders/6/items/7/review", "/orders/2/items/8/review", "/orders/2/items/9/review", "/orders/1/items/10/review", "/orders/4/review"} {
		f.post(path, "buyer", `{"rating":3}`, 409)
	}
	f.assertSummary(burger, 0, 0)
	f.assertSummary(org, 0, 0)
	f.assertSummary("/restaurants/cafe/items/3/reviews", 0, 0)
}

func TestREV004GuestItemRatingWithoutComment(t *testing.T) {
	t.Parallel()
	f := setupReviews(t)
	f.post(guestPurchase, f.guest, `{"rating":1,"comment":"Guest text"}`, 403)
	f.assertSummary(burger, 0, 0)
	f.post(guestPurchase, f.guest, `{"rating":1}`, 200)
	p := f.assertSummary(burger, 1, 1)
	if len(p.Data) != 1 || p.Data[0].Comment != nil {
		f.t.Fatalf("guest comment persisted: %+v", p.Data)
	}
}

func TestREV005RepeatBuyerDistinctPurchases(t *testing.T) {
	t.Parallel()
	f := setupReviews(t)
	f.post(itemPurchase, "buyer", `{"rating":3,"comment":"Great"}`, 200)
	f.post("/orders/2/items/3/review", "buyer", `{"rating":2,"comment":"Another visit"}`, 200)
	f.post(itemPurchase, "buyer", `{"rating":1}`, 409)
	p := f.assertSummary(burger, 2, 5)
	if len(p.Data) != 2 || p.Data[0].Comment == nil || *p.Data[0].Comment != "Another visit" || p.Data[1].Comment == nil || *p.Data[1].Comment != "Great" {
		f.t.Fatalf("independent reviews were not preserved: %+v", p.Data)
	}
}

func TestREV006ConcurrentGuestReplay(t *testing.T) {
	t.Parallel()
	f := setupReviews(t)
	var wg sync.WaitGroup
	results := make(chan int, 2)
	start := make(chan struct{})
	for range 2 {
		wg.Go(func() { <-start; results <- f.run("POST", guestPurchase, f.guest, `{"rating":1}`).Code })
	}
	close(start)
	wg.Wait()
	close(results)
	counts := map[int]int{}
	for status := range results {
		counts[status]++
	}
	if counts[200] != 1 || counts[409] != 1 {
		f.t.Fatalf("guest replay statuses: %v", counts)
	}
	f.assertSummary(burger, 1, 1)
}

func TestREV007ComboAndDishAreIndependent(t *testing.T) {
	t.Parallel()
	f := setupReviews(t)
	f.post(itemPurchase, "buyer", `{"rating":3}`, 200)
	f.post("/orders/1/items/2/review", "buyer", `{"rating":1}`, 200)
	f.assertSummary(burger, 1, 3)
	f.assertSummary(combo, 1, 1)
}

func TestREV008OrganizationUpdateAcrossBranches(t *testing.T) {
	t.Parallel()
	f := setupReviews(t)
	f.post(restaurantPurchase, "buyer", `{"rating":5,"comment":"First branch"}`, 200)
	before := f.assertSummary(org, 1, 5)
	f.post("/orders/2/review", "buyer", `{"rating":4,"comment":"Updated after another branch"}`, 200)
	after := f.assertSummary(org, 1, 4)
	if len(before.Data) != 1 || len(after.Data) != 1 || before.Data[0].ID != after.Data[0].ID || after.Data[0].Comment == nil || *after.Data[0].Comment != "Updated after another branch" || after.Summary.Sold != nil {
		f.t.Fatalf("restaurant review was not updated in place: %+v", after)
	}
}

func TestREV009RestaurantGuestReplay(t *testing.T) {
	t.Parallel()
	f := setupReviews(t)
	f.post(guestRestaurant, f.guest, `{"rating":2,"comment":"Guest text"}`, 403)
	f.post(guestRestaurant, f.guest, `{"rating":2}`, 200)
	f.post(guestRestaurant, f.guest, `{"rating":3}`, 409)
	p := f.assertSummary(org, 1, 2)
	if len(p.Data) != 1 || p.Data[0].Comment != nil {
		f.t.Fatalf("guest review: %+v", p.Data)
	}
}

func TestREV010PublicPaginationAndPrivacy(t *testing.T) {
	t.Parallel()
	f := setupReviews(t)
	f.post(itemPurchase, "buyer", `{"rating":3}`, 200)
	f.post(guestPurchase, f.guest, `{"rating":1}`, 200)
	first := f.assertSummary(burger+"?page_size=1", 2, 4)
	second := f.assertSummary(burger+"?page_size=1&page=2", 2, 4)
	past := f.assertSummary(burger+"?page_size=1&page=3", 2, 4)
	if len(first.Data) != 1 || len(second.Data) != 1 || first.Data[0].Rating != 1 || second.Data[0].Rating != 3 || len(past.Data) != 0 || second.Page != 2 || second.Size != 1 {
		f.t.Fatalf("wrong page contents: %+v %+v %+v", first, second, past)
	}
}

func TestREV011PublicationAndTenantGate(t *testing.T) {
	t.Parallel()
	f := setupReviews(t)
	f.post(itemPurchase, "buyer", `{"rating":3}`, 200)
	f.assertSummary(burger, 1, 3)
	f.exec(`INSERT INTO physical_restaurant(restaurant_id,address_id,name,cnpj,owner_cnpj,cnpj_verified) VALUES(2,1,'Other','12345678000300','12345678000100',true)`)
	f.call("GET", "/restaurants/other/items/1/reviews", "", "", 404)
	f.exec(`UPDATE menu_item SET archived_at=now() WHERE id=1`)
	f.call("GET", burger, "", "", 404)
	f.exec(`UPDATE menu_item SET archived_at=NULL WHERE id=1; UPDATE physical_restaurant SET cnpj_verified=false WHERE restaurant_id=1`)
	f.call("GET", burger, "", "", 404)
	f.call("GET", org, "", "", 404)
}

func TestREV012SalesExcludeNonPurchases(t *testing.T) {
	t.Parallel()
	f := setupReviews(t)
	b, c := f.assertSummary(burger, 0, 0), f.assertSummary(combo, 0, 0)
	if b.Summary.Sold == nil || *b.Summary.Sold != 6 || c.Summary.Sold == nil || *c.Summary.Sold != 1 {
		f.t.Fatalf("wrong sold counts: %+v %+v", b.Summary, c.Summary)
	}
	f.exec(`UPDATE "order" SET status='cancelled' WHERE id=4`)
	p := f.get(burger)
	if p.Summary.Sold == nil || *p.Summary.Sold != 4 {
		f.t.Fatalf("cancellation did not subtract units: %+v", p.Summary)
	}
}

func TestREV013RollupUpdateAndDelete(t *testing.T) {
	t.Parallel()
	f := setupReviews(t)
	f.post(itemPurchase, "buyer", `{"rating":3}`, 200)
	f.assertSummary(burger, 1, 3)
	f.exec(`UPDATE item_review SET rating=2 WHERE order_item_id=1`)
	f.assertSummary(burger, 1, 2)
	f.exec(`DELETE FROM item_review WHERE order_item_id=1`)
	p := f.assertSummary(burger, 0, 0)
	if len(p.Data) != 0 {
		f.t.Fatalf("deleted review still visible: %+v", p.Data)
	}
}

func TestREV014MultibyteCommentLimit(t *testing.T) {
	t.Parallel()
	f := setupReviews(t)
	comment := strings.Repeat("á", 2000)
	f.post(restaurantPurchase, "buyer", fmt.Sprintf(`{"rating":5,"comment":%q}`, comment), 200)
	f.post(restaurantPurchase, "buyer", fmt.Sprintf(`{"rating":2,"comment":%q}`, comment+"á"), 400)
	p := f.assertSummary(org, 1, 5)
	if len(p.Data) != 1 || p.Data[0].Comment == nil || *p.Data[0].Comment != comment {
		f.t.Fatal("overlong input changed the review or valid comment was truncated")
	}
}
