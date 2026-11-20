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

func TestReviewPurchasesAndTotals(t *testing.T) {
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
	run := func(method, path, key, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		switch key {
		case "buyer", "other":
			r.Header.Set("Authorization", "Bearer "+key)
		case "":
		default:
			r.Header.Set("X-Order-Token", key)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	call := func(method, path, key, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		w := run(method, path, key, body)
		if w.Code != want {
			t.Fatalf("%s %s: got %d want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		return w
	}
	post := func(path, key, body string, want int) { t.Helper(); call("POST", path, key, body, want) }
	main := "/orders/1/items/1/review"
	post(main, "other", `{"rating":3}`, 403)
	post(main, guest, `{"rating":3}`, 403)
	post("/orders/1/items/3/review", "buyer", `{"rating":3}`, 404)
	post("/orders/4/items/5/review", "buyer", `{"rating":3}`, 409)
	post("/orders/2/items/8/review", "buyer", `{"rating":3}`, 409)
	post("/orders/2/items/9/review", "buyer", `{"rating":3}`, 409)
	post("/orders/1/items/10/review", "buyer", `{"rating":3}`, 409)
	post(main, "buyer", `{"rating":3,"comment":"Great"}`, 200)
	post(main, "buyer", `{"rating":1}`, 409)
	post("/orders/1/items/2/review", "buyer", `{"rating":1}`, 200)
	post("/orders/2/items/3/review", "buyer", `{"rating":2,"comment":"Another visit"}`, 200)
	// A racing replay of one anonymous purchase must not inflate the rollup.
	var wg sync.WaitGroup
	results := make(chan int, 2)
	for range 2 {
		wg.Go(func() { results <- run("POST", "/orders/3/items/4/review", guest, `{"rating":1}`).Code })
	}
	wg.Wait()
	close(results)
	counts := map[int]int{}
	for status := range results {
		counts[status]++
	}
	if counts[200] != 1 || counts[409] != 1 {
		t.Fatalf("guest replay: %v", counts)
	}
	post("/orders/1/review", "buyer", `{"rating":5,"comment":"First branch"}`, 200)
	post("/orders/2/review", "buyer", `{"rating":4,"comment":"Updated after another branch"}`, 200)
	post("/orders/3/review", guest, `{"rating":2}`, 200)
	post("/orders/3/review", guest, `{"rating":3}`, 409)
	get := func(path string) page {
		t.Helper()
		w := call("GET", path, "", "", 200)
		var p page
		if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(w.Body.String(), "order_item_id") || strings.Contains(w.Body.String(), "account_id") {
			t.Fatal("private identifiers in public review")
		}
		return p
	}
	burger := get("/api/restaurants/cafe/items/1/reviews?page_size=1")
	if burger.Summary.Count != 3 || burger.Summary.Sum != 6 || *burger.Summary.Average != 2 || *burger.Summary.Sold != 6 || len(burger.Data) != 1 {
		t.Fatalf("burger totals: %+v", burger)
	}
	combo := get("/restaurants/cafe/items/2/reviews")
	if combo.Summary.Count != 1 || combo.Summary.Sum != 1 || *combo.Summary.Sold != 1 {
		t.Fatalf("combo totals: %+v", combo)
	}
	org := get("/restaurants/cafe/reviews")
	if org.Summary.Count != 2 || org.Summary.Sum != 6 || *org.Summary.Average != 3 || org.Summary.Sold != nil {
		t.Fatalf("restaurant totals: %+v", org)
	}
	call("GET", "/restaurants/other/items/1/reviews", "", "", 404)
	// Editing and deleting at the persistence layer preserves exact totals.
	exec(`DELETE FROM item_review WHERE order_item_id=4`)
	if p := get("/restaurants/cafe/items/1/reviews"); p.Summary.Count != 2 || p.Summary.Sum != 5 {
		t.Fatalf("rollup not corrected: %+v", p)
	}
}
