package order

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"pratello/server/internal/auth"
	"pratello/server/internal/db"
	"pratello/server/internal/menu"
	"pratello/server/internal/restaurants"
	"pratello/server/internal/sales"
	"pratello/server/internal/transaction"
)

func openOrderTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("AUTH_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL for PostgreSQL integration")
	}
	root, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	schema := fmt.Sprintf("orders_test_%d", time.Now().UnixNano())
	if _, err = root.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := root.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
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
	c, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if err = db.Migrate(c); err != nil {
		t.Fatal(err)
	}
	return c
}

func orderFixture(t *testing.T) (*sql.DB, *Service, *http.ServeMux) {
	t.Helper()
	c := openOrderTestDB(t)
	_, err := c.Exec(`INSERT INTO restaurant(name_slug,name) VALUES('cafe','Cafe'),('other','Other');
	INSERT INTO account(email,password,is_administrator) VALUES('admin@test.local','x',true),('customer@test.local','x',false);
	INSERT INTO "Address"("CEP",street,number) VALUES('12345678','Rua','1');
	INSERT INTO physical_restaurant(restaurant_id,address_id,name,cnpj,owner_cnpj,cnpj_verified) VALUES(1,1,'Cafe','12345678000100','12345678000100',true),(2,1,'Other','12345678000200','12345678000200',true);
	INSERT INTO dining_table(physical_restaurant_id,number) VALUES(1,7),(2,7);
	INSERT INTO category(restaurant_id,name) VALUES(1,'Lunch');
	INSERT INTO menu_item(restaurant_id,category_id,name,base_price_cents) VALUES(1,1,'Burger',1500);
	INSERT INTO coupon(restaurant_id,code,discount_type,discount_value,max_uses) VALUES(1,'SAVE','fixed_cents',300,1);`)
	if err != nil {
		t.Fatal(err)
	}
	identify := func(_ context.Context, key string) (int, error) {
		switch key {
		case "staff":
			return 1, nil
		case "customer":
			return 2, nil
		}
		return 0, errForbidden
	}
	mux := http.NewServeMux()
	svc := NewService(c)
	NewHandler(svc, identify, "https://orders.example.test").Routes(mux)
	menu.NewHandler(menu.NewService(c), identify).Routes(mux)
	sales.NewHandler(sales.NewService(c), identify).Routes(mux)
	restaurants.NewHandler(c, nil).Routes(mux)
	transaction.NewHandler(transaction.NewService(c, transaction.MockProvider{}), identify).Routes(mux)
	return c, svc, mux
}

func orderRequest(mux *http.ServeMux, method, path, key, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Idempotency-Key", "payment-test")
	if key == "staff" || key == "customer" {
		r.Header.Set("Authorization", "Bearer "+key)
	} else if key != "" {
		r.Header.Set("X-Order-Token", key)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w
}

func orderCall(t *testing.T, mux *http.ServeMux) func(string, string, string, string, int) *httptest.ResponseRecorder {
	return func(method, path, key, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		w := orderRequest(mux, method, path, key, body)
		if w.Code != want {
			t.Fatalf("%s %s got %d want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		return w
	}
}

func decodeOrderResponse(t *testing.T, w *httptest.ResponseRecorder, dest any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), dest); err != nil {
		t.Fatal(err)
	}
}

func TestFIN013014016GuestCPFLoyaltyAndSignup(t *testing.T) {
	c, _, mux := orderFixture(t)
	call := orderCall(t, mux)
	const root = "/api/restaurants/cafe"
	const cpf = "52998224725"
	const draft = `{"customer_cpf":"52998224725","items":[{"item_id":1,"quantity":1}]}`
	coins := func(w *httptest.ResponseRecorder, want int) {
		t.Helper()
		var b struct {
			Coins int `json:"coins_amount"`
		}
		decodeOrderResponse(t, w, &b)
		if b.Coins != want {
			t.Fatalf("balance %d, want %d", b.Coins, want)
		}
	}
	lookup := func(want int) {
		t.Helper()
		coins(call("POST", root+"/loyalty/balance", "", `{"cpf":"52998224725"}`, 200), want)
	}
	open := func() (string, string) {
		t.Helper()
		w := call("POST", root+"/branches/1/orders/local", "", draft, 201)
		var v struct{ ID int }
		decodeOrderResponse(t, w, &v)
		return fmt.Sprintf("/api/orders/%d", v.ID), w.Header().Get("X-Order-Token")
	}
	lookup(0)
	path, key := open()
	// A reserved PIX amount is not money received and must not earn coins.
	p := call("POST", path+"/transactions", key, `{"payment_method":"pix","amount_cents":1500}`, 201)
	var tx struct{ ID int }
	decodeOrderResponse(t, p, &tx)
	lookup(0)
	confirm := fmt.Sprintf("/api/transactions/%d/pix/confirm", tx.ID)
	call("POST", confirm, key, "", 200)
	call("POST", confirm, key, "", 200)
	lookup(15)
	path, key = open()
	call("POST", path+"/transactions", key, `{"payment_method":"credit_card","amount_cents":1500}`, 201)
	call("POST", path+"/transactions", key, `{"payment_method":"credit_card","amount_cents":1500}`, 201)
	lookup(30)
	var shadow, entries int
	err := c.QueryRow(`SELECT a.id,(SELECT count(*) FROM wallet_transaction WHERE reason='paid_order_earn') FROM account a WHERE a.cpf=$1 AND a.email IS NULL`, cpf).Scan(&shadow, &entries)
	if err != nil || entries != 2 {
		t.Fatalf("loyalty identity and ledger: %d %v", entries, err)
	}
	identity := auth.NewService(auth.NewStore(c))
	user, err := identity.Register(context.Background(), auth.RegisterInput{Name: "Loyal buyer", Email: "cpf-buyer@test.local", Password: "strong-password-123", CPF: cpf})
	if err != nil || user.ID != shadow {
		t.Fatalf("signup did not preserve the shadow account: %+v %v", user, err)
	}
	lookup(30)
	signed := http.NewServeMux()
	sales.NewHandler(sales.NewService(c), func(context.Context, string) (int, error) { return user.ID, nil }).Routes(signed)
	coins(orderCall(t, signed)("GET", "/api/me/restaurants/cafe/loyalty/balance", "customer", "", 200), 30)
	const quote = root + "/branches/1/checkout/quote"
	const redemption = `{"items":[{"item_id":1,"quantity":1}],"coins":30}`
	call("POST", quote, "", redemption, 403)
	var preview struct {
		Payable int `json:"payable_cents"`
	}
	decodeOrderResponse(t, orderCall(t, signed)("POST", quote, "customer", redemption, 200), &preview)
	if preview.Payable != 1470 {
		t.Fatalf("sign-up balance not available for redemption: %+v", preview)
	}
	call("GET", "/api/me/restaurants/cafe/loyalty/balance", "", "", 401)
	call("POST", root+"/loyalty/balance", "", `{"cpf":"invalid"}`, 400)
}

func TestOrderQRCodeCheckoutAndAnalyticsPostgres(t *testing.T) {
	c, _, mux := orderFixture(t)
	call := orderCall(t, mux)
	read := func(w *httptest.ResponseRecorder, dest any) { t.Helper(); decodeOrderResponse(t, w, dest) }
	const branch = "/api/restaurants/cafe/branches/1"
	call("POST", branch+"/order-qr", "customer", `{}`, 403)
	call("POST", branch+"/order-qr", "staff", `{"table_id":2}`, 404)
	var qr struct {
		ID  int64
		URL string
		PNG string `json:"qr_code_base64"`
	}
	read(call("POST", branch+"/order-qr", "staff", `{"table_id":1}`, 201), &qr)
	link, e := url.Parse(qr.URL)
	png, decodeErr := base64.StdEncoding.DecodeString(qr.PNG)
	if e != nil || link.Host != "orders.example.test" || decodeErr != nil || !strings.HasPrefix(string(png), "\x89PNG") {
		t.Fatalf("bad QR: %s", qr.URL)
	}
	call("GET", link.Path, "", "", 200)
	call("GET", link.Path, "", "", 200)
	var count int
	if e = c.QueryRow(`SELECT count(*) FROM "order"`).Scan(&count); e != nil || count != 0 {
		t.Fatalf("landing GET created an order: %d %v", count, e)
	}
	type detail struct {
		ID        int
		QR        int64      `json:"qr_id"`
		Total     int64      `json:"total_cents"`
		Submitted *time.Time `json:"submitted_at"`
		Local     struct {
			Number *int `json:"table_number"`
		} `json:"local_info"`
		Items []struct {
			ID   int
			Name string `json:"item_name"`
		}
	}
	var first, second detail
	w := call("POST", link.Path+"/open", "", "", 201)
	key := w.Header().Get("X-Order-Token")
	if key == "" || strings.Contains(w.Body.String(), key) {
		t.Fatal("guest token missing or leaked into the body")
	}
	read(w, &first)
	w = call("POST", link.Path+"/open", "", "", 201)
	otherKey := w.Header().Get("X-Order-Token")
	read(w, &second)
	if first.ID == second.ID || key == "" || key == otherKey || first.QR != qr.ID || second.QR != qr.ID || first.Local.Number == nil || *first.Local.Number != 7 || first.Submitted != nil {
		t.Fatalf("incorrect scan attribution: %+v %+v", first, second)
	}
	path := fmt.Sprintf("/api/orders/%d", first.ID)
	var anonymous bool
	var hash []byte
	if err := c.QueryRow(`SELECT account_id IS NULL AND customer_cpf IS NULL,guest_token_hash FROM "order" WHERE id=$1`, first.ID).Scan(&anonymous, &hash); err != nil || !anonymous || string(hash) != string(digest(key)) {
		t.Fatalf("guest credential was not safely stored: %v", err)
	}
	call("GET", path, otherKey, "", 403)
	call("GET", path, "customer", "", 403)
	call("GET", "/api/orders/99999", key, "", 404)
	call("POST", path+"/items", otherKey, `{"item_id":1,"quantity":1}`, 403)
	call("PATCH", path+"/status", "staff", `{"status":"preparing"}`, 409)
	call("POST", path+"/transactions", key, `{"payment_method":"pix","amount_cents":100}`, 409)
	call("POST", path+"/submit", key, `{}`, 409)
	itemBody := `{"item_id":1,"quantity":1,"notes":"Well done"}`
	var removed struct{ ID int }
	read(call("POST", path+"/items", key, itemBody, 201), &removed)
	call("DELETE", fmt.Sprintf("/api/orders/%d/items/%d", second.ID, removed.ID), otherKey, "", 404)
	call("DELETE", fmt.Sprintf("%s/items/%d?reason=mistake", path, removed.ID), key, "", 204)
	var audited bool
	if err := c.QueryRow(`SELECT voided_at IS NOT NULL AND voided_reason='mistake' AND (SELECT total_cents=0 FROM "order" WHERE id=$2) FROM order_draft_item WHERE id=$1 AND order_id=$2`, removed.ID, first.ID).Scan(&audited); err != nil || !audited {
		t.Fatalf("void did not leave an audit entry and reset the total: %v", err)
	}
	call("POST", path+"/items", key, itemBody, 201)
	read(call("POST", path+"/submit", key, `{"coupon_codes":["save"]}`, 200), &first)
	if first.Total != 1200 || first.Submitted == nil || len(first.Items) != 1 || first.Items[0].Name != "Burger" {
		t.Fatalf("bad finalized order: %+v", first)
	}
	call("POST", path+"/submit", key, `{}`, 409)
	call("POST", path+"/items", key, itemBody, 409)
	call("DELETE", fmt.Sprintf("%s/items/%d", path, first.Items[0].ID), key, "", 409)
	var paid struct {
		ID     int
		Status string
	}
	read(call("POST", path+"/transactions", key, `{"payment_method":"pix","amount_cents":1200}`, 201), &paid)
	if paid.Status != "pending" {
		t.Fatalf("unexpected payment: %+v", paid)
	}
	call("POST", fmt.Sprintf("/api/transactions/%d/pix/confirm", paid.ID), key, "", 200)
	call("PATCH", path+"/status", "staff", `{"status":"delivered"}`, 409)
	call("PATCH", path+"/status", "staff", `{"status":"unknown"}`, 400)
	call("PATCH", path+"/status", "customer", `{"status":"preparing"}`, 403)
	call("PATCH", path+"/status", "staff", `{"status":"preparing"}`, 200)
	call("PATCH", path+"/status", "staff", `{"status":"ready"}`, 200)
	var ready struct{ Total int }
	read(call("GET", branch+"/orders?status=ready", "staff", "", 200), &ready)
	if ready.Total != 1 {
		t.Fatalf("wrong ready queue: %d", ready.Total)
	}
	call("PATCH", path+"/status", "staff", `{"status":"delivered"}`, 200)
	call("PATCH", path+"/status", "staff", `{"status":"cancelled"}`, 409)
	var events, conversion, uses int
	if e = c.QueryRow(`SELECT (SELECT count(*) FROM ux_event WHERE metadata->>'order_id'=$1::text AND metadata->>'qr_id'=$2::text),(SELECT count(*) FROM menu_session WHERE converted_order_id=$1::int AND dining_table_id=1 AND channel='qr_table'),(SELECT used_count FROM coupon)`, first.ID, qr.ID).Scan(&events, &conversion, &uses); e != nil || events != 5 || conversion != 1 || uses != 1 {
		t.Fatalf("events=%d conversion=%d uses=%d error=%v", events, conversion, uses, e)
	}
	call("PUT", branch+"/delivery-fee", "customer", `{"delivery_fee_cents":500}`, 403)
	call("PUT", branch+"/delivery-fee", "staff", `{"delivery_fee_cents":-1}`, 400)
	call("PUT", branch+"/delivery-fee", "staff", `{"delivery_fee_cents":500}`, 200)
	call("POST", branch+"/orders/local", "", `{"delivery_address":{}}`, 400)
	call("POST", branch+"/orders/remote", "", `{"table_id":1}`, 400)
	call("POST", branch+"/orders/remote", "", `{"phone":"bad","delivery_address":{}}`, 400)
	call("POST", branch+"/orders/remote", "", `{"freight_cents":0}`, 400)
	call("POST", branch+"/orders/local", "", `{"table_id":1,"table_number":7}`, 400)
	call("POST", branch+"/orders/local", "", `{"table_id":2}`, 404)
	if _, err := c.Exec(`UPDATE dining_table SET is_active=false WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	call("POST", branch+"/orders/local", "", `{"table_id":1}`, 404)
	call("POST", "/api/restaurants/cafe/branches/2/orders/local", "", `{}`, 404)
	call("POST", "/api/restaurants/cafe/branches/999/orders/local", "", `{}`, 404)
	var remote detail
	delivery := call("POST", branch+"/orders/remote", "", `{"phone":"81999990000","delivery_address":{"cep":"12345678","street":"Rua","number":"1","city":"Recife","state":"PE"},"items":[{"item_id":1,"quantity":1}]}`, 201)
	deliveryKey := delivery.Header().Get("X-Order-Token")
	read(delivery, &remote)
	if remote.Total != 2000 || remote.Submitted == nil {
		t.Fatalf("wrong delivery fee: %+v", remote)
	}
	remotePath := fmt.Sprintf("/api/orders/%d", remote.ID)
	if deliveryKey == "" {
		t.Fatal("delivery guest did not get a credential")
	}
	call("GET", remotePath, deliveryKey, "", 200)
	call("PATCH", remotePath+"/status", "staff", `{"status":"cancelled"}`, 200)
	call("POST", remotePath+"/transactions", deliveryKey, `{"payment_method":"pix","amount_cents":2000}`, 409)
	var counter detail
	read(call("POST", branch+"/orders/local", "", `{}`, 201), &counter)
	if counter.Local.Number != nil || counter.Submitted != nil {
		t.Fatalf("table unexpectedly required: %+v", counter)
	}
	var page struct{ Total int }
	if _, err := c.Exec(`INSERT INTO "order"(physical_restaurant_id,order_type) VALUES(2,'local')`); err != nil {
		t.Fatal(err)
	}
	read(call("GET", branch+"/orders", "staff", "", 200), &page)
	if page.Total != 2 {
		t.Fatalf("draft or foreign order appeared in staff queue: %d", page.Total)
	}
	call("GET", branch+"/orders", "customer", "", 403)
	read(call("POST", branch+"/order-qr", "staff", `{}`, 201), &qr)
	link, e = url.Parse(qr.URL)
	if e != nil {
		t.Fatal(e)
	}
	read(call("POST", link.Path+"/open", "", "", 201), &counter)
	if counter.QR != qr.ID || counter.Local.Number != nil {
		t.Fatalf("tableless QR attribution: %+v", counter)
	}
	if _, err := c.Exec(`UPDATE physical_restaurant SET cnpj_verified=false WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	call("POST", link.Path+"/open", "", "", 404)
}

func TestOrderCheckoutRevalidatesAndFreezesSnapshot(t *testing.T) {
	c, svc, mux := orderFixture(t)
	call := orderCall(t, mux)
	const branch = "/api/restaurants/cafe/branches/1"
	at := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return at }
	setup := `INSERT INTO item_sale(item_id,discount_type,discount_value,starts_at,ends_at) VALUES(1,'fixed_cents',200,'2026-01-01T11:00:00Z','2026-01-01T13:00:00Z');
	INSERT INTO coupon(restaurant_id,code,discount_type,discount_value,min_order_cents) VALUES(1,'MINIMUM','fixed_cents',200,9999),(2,'FOREIGN','fixed_cents',200,0);
	INSERT INTO item_side(item_id,title,min_options,max_options) VALUES(1,'Extras',0,1);
	INSERT INTO item_side_option(item_side_id,name,additional_price_cents) VALUES(1,'Cheese',100);
	INSERT INTO ingredient(restaurant_id,name) VALUES(1,'Onions');
	INSERT INTO item_ingredient(item_id,ingredient_id,is_removable) VALUES(1,1,true);`
	if _, err := c.Exec(setup); err != nil {
		t.Fatal(err)
	}
	type snapshot struct {
		ID       int
		Total    int64 `json:"total_cents"`
		Discount int64 `json:"discount_cents"`
		Local    struct {
			Number int `json:"table_number"`
		} `json:"local_info"`
		Items []struct {
			Name    string `json:"item_name"`
			Notes   string
			Unit    int64                   `json:"unit_price_cents"`
			Total   int64                   `json:"line_total_cents"`
			Options []struct{ Name string } `json:"selected_options"`
		}
	}
	read := func(w *httptest.ResponseRecorder) snapshot {
		t.Helper()
		var v snapshot
		decodeOrderResponse(t, w, &v)
		return v
	}
	first := read(call("POST", branch+"/orders/local", "customer", `{"table_number":7,"items":[{"item_id":1,"quantity":1}]}`, 201))
	if first.Local.Number != 7 || first.Total != 1300 || first.Discount != 200 || len(first.Items) != 1 || first.Items[0].Unit != 1500 || first.Items[0].Total != 1300 {
		t.Fatalf("wrong active sale snapshot: %+v", first)
	}
	v := read(call("POST", branch+"/orders/local", "customer", `{}`, 201))
	path := fmt.Sprintf("/api/orders/%d", v.ID)
	for _, bad := range []string{`{"item_id":1,"quantity":0}`, `{"item_id":1,"quantity":-1}`} {
		call("POST", path+"/items", "customer", bad, 400)
	}
	if v := read(call("GET", path, "customer", "", 200)); v.Total != 0 || len(v.Items) != 0 {
		t.Fatalf("invalid additions changed cart: %+v", v)
	}
	call("POST", path+"/items", "customer", `{"item_id":1,"quantity":1,"option_ids":[1],"remove_ingredient_ids":[1],"notes":"Well done"}`, 201)
	call("POST", path+"/items", "customer", `{"item_id":1,"quantity":2}`, 201)
	if _, err := c.Exec(`INSERT INTO branch_item_availability(branch_id,item_id,is_available) VALUES(1,1,false)`); err != nil {
		t.Fatal(err)
	}
	call("POST", path+"/submit", "customer", `{"coupon_codes":["SAVE"]}`, 409)
	if _, err := c.Exec(`UPDATE branch_item_availability SET is_available=true WHERE branch_id=1 AND item_id=1`); err != nil {
		t.Fatal(err)
	}
	call("POST", path+"/submit", "customer", `{"coupon_codes":["FOREIGN"]}`, 404)
	var untampered bool
	if err := c.QueryRow(`SELECT submitted_at IS NULL AND (SELECT count(*) FROM order_item WHERE order_id=o.id)=0 AND (SELECT converted_order_id IS NULL FROM menu_session WHERE id=o.menu_session_id) AND (SELECT sum(used_count) FROM coupon)=0 FROM "order" o WHERE id=$1`, v.ID).Scan(&untampered); err != nil || !untampered {
		t.Fatalf("failed checkout mutated state: %v", err)
	}
	at = at.Add(time.Hour)
	final := read(call("POST", path+"/submit", "customer", `{"coupon_codes":["MINIMUM"]}`, 200))
	if final.Total != 4600 || final.Discount != 0 || len(final.Items) != 2 || final.Items[0].Unit != 1600 || len(final.Items[0].Options) != 1 || final.Items[0].Options[0].Name != "Cheese" || final.Items[0].Notes != "Well done\nWithout: Onions" {
		t.Fatalf("wrong final customization snapshot: %+v", final)
	}
	if _, err := c.Exec(`UPDATE menu_item SET name='Renamed',base_price_cents=500 WHERE id=1; UPDATE item_side_option SET name='Changed' WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	after := read(call("GET", path, "customer", "", 200))
	if after.Items[0].Name != "Burger" || after.Items[0].Unit != 1600 || after.Items[0].Options[0].Name != "Cheese" || after.Total != 4600 {
		t.Fatalf("catalog edit rewrote order: %+v", after)
	}
}

func TestORD013ComboPreparationLines(t *testing.T) {
	c, _, mux := orderFixture(t)
	call := orderCall(t, mux)
	_, err := c.Exec(`INSERT INTO menu_item(restaurant_id,category_id,name,base_price_cents,is_combo) VALUES
	(1,1,'Fries',700,false),(1,1,'Lunch combo',2000,true);
	INSERT INTO combo_item(combo_id,included_item_id,quantity) VALUES(3,1,2),(3,2,1);
	INSERT INTO ingredient(restaurant_id,name) VALUES(1,'Onions');
	INSERT INTO item_ingredient(item_id,ingredient_id,is_removable) VALUES(1,1,true)`)
	if err != nil {
		t.Fatal(err)
	}
	var componentID int
	if err := c.QueryRow(`SELECT id FROM combo_item WHERE combo_id=3 AND included_item_id=1`).Scan(&componentID); err != nil {
		t.Fatal(err)
	}
	type detail struct {
		ID       int
		Subtotal int `json:"subtotal_cents"`
		Discount int `json:"discount_cents"`
		Total    int `json:"total_cents"`
		Items    []struct {
			ID, Quantity int
			Item         int    `json:"item_id"`
			Name         string `json:"item_name"`
			Parent       *int   `json:"parent_order_item_id"`
			Unit         int    `json:"unit_price_cents"`
			Total        int    `json:"line_total_cents"`
			Notes        string
		}
	}
	body := fmt.Sprintf(`{"coupon_codes":["SAVE"],"items":[{"item_id":3,"quantity":2,"component_removals":[{"combo_item_id":%d,"unit":1,"remove_ingredient_ids":[1]}]},{"item_id":1,"quantity":1}]}`, componentID)
	var v detail
	decodeOrderResponse(t, call("POST", "/api/restaurants/cafe/branches/1/orders/local", "customer", body, 201), &v)
	if v.Subtotal != 5500 || v.Discount != 300 || v.Total != 5200 || len(v.Items) != 5 {
		t.Fatalf("wrong combo charge or line count: %+v", v)
	}
	parent := v.Items[0]
	if parent.Parent != nil || parent.Item != 3 || parent.Quantity != 2 || parent.Unit != 2000 {
		t.Fatalf("invalid combo parent: %+v", parent)
	}
	var plain, customized, fries, billable, standalone int
	for _, line := range v.Items {
		billable += line.Total
		if line.Parent == nil {
			if line.Item == 1 && line.Quantity == 1 && line.Unit == 1500 {
				standalone++
			}
			continue
		}
		if *line.Parent != parent.ID || line.Unit != 0 || line.Total != 0 {
			t.Fatalf("component was charged or attached to the wrong parent: %+v", line)
		}
		switch {
		case line.Item == 1 && line.Name == "Burger" && line.Notes == "Without: Onions":
			customized += line.Quantity
		case line.Item == 1 && line.Name == "Burger" && line.Notes == "":
			plain += line.Quantity
		case line.Item == 2 && line.Name == "Fries" && line.Notes == "":
			fries += line.Quantity
		default:
			t.Fatalf("unexpected component: %+v", line)
		}
	}
	if plain != 2 || customized != 2 || fries != 2 || standalone != 1 || billable != 5200 {
		t.Fatalf("bad line breakdown: %+v", v.Items)
	}
	if _, err = c.Exec(`UPDATE menu_item SET name='Renamed' WHERE id=1; DELETE FROM combo_item WHERE combo_id=3`); err != nil {
		t.Fatal(err)
	}
	var persisted detail
	decodeOrderResponse(t, call("GET", fmt.Sprintf("/api/orders/%d", v.ID), "staff", "", 200), &persisted)
	if !reflect.DeepEqual(v, persisted) {
		t.Fatalf("catalog edit changed the submitted kitchen ticket: %+v", persisted)
	}
}

func TestOrderConcurrentWritesRecordOneConversion(t *testing.T) {
	c, _, mux := orderFixture(t)
	call := orderCall(t, mux)
	var draft struct{ ID int }
	decodeOrderResponse(t, call("POST", "/api/restaurants/cafe/branches/1/orders/local", "customer", `{}`, 201), &draft)
	path := fmt.Sprintf("/api/orders/%d", draft.ID)
	race := func(path, body string) map[int]int {
		var wg sync.WaitGroup
		gate := make(chan struct{})
		out := make(chan int, 2)
		for range 2 {
			wg.Go(func() { <-gate; out <- orderRequest(mux, "POST", path, "customer", body).Code })
		}
		close(gate)
		wg.Wait()
		close(out)
		counts := map[int]int{}
		for status := range out {
			counts[status]++
		}
		return counts
	}
	if got := race(path+"/items", `{"item_id":1,"quantity":1}`); got[201] != 2 {
		t.Fatalf("concurrent additions: %v", got)
	}
	if got := race(path+"/submit", `{"coupon_codes":["SAVE"]}`); got[200] != 1 || got[409] != 1 {
		t.Fatalf("concurrent submissions: %v", got)
	}
	var lines, total, uses, conversions, events, adds int
	err := c.QueryRow(`SELECT (SELECT count(*) FROM order_item WHERE order_id=$1),total_cents,(SELECT used_count FROM coupon),(SELECT count(*) FROM menu_session WHERE converted_order_id=$1),(SELECT count(*) FROM ux_event WHERE session_id=o.menu_session_id AND event_type='order_submitted'),(SELECT count(*) FROM ux_event WHERE session_id=o.menu_session_id AND event_type='cart_item_add' AND item_id=1 AND quantity=1) FROM "order" o WHERE id=$1`, draft.ID).Scan(&lines, &total, &uses, &conversions, &events, &adds)
	if err != nil || lines != 2 || total != 2700 || uses != 1 || conversions != 1 || events != 1 || adds != 2 {
		t.Fatalf("unexpected checkout effects: lines=%d total=%d uses=%d conversions=%d submissions=%d additions=%d err=%v", lines, total, uses, conversions, events, adds, err)
	}
}
