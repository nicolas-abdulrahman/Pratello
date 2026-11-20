package sales

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"pratello/server/internal/db"
	"pratello/server/internal/menu"
)

func openSalesTestDB(t *testing.T) *sql.DB {
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
	schema := fmt.Sprintf("sales_test_%d", time.Now().UnixNano())
	if _, err = root.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = root.Exec("DROP SCHEMA " + schema + " CASCADE") })
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, err := url.Parse(dsn)
		if err != nil {
			t.Fatal(err)
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
	_, err = conn.Exec(`INSERT INTO restaurant(name_slug,name) VALUES('test','Test'),('other','Other');
	INSERT INTO account(email,password,is_administrator) VALUES('admin@test.local','hash',true),('customer@test.local','hash',false),('manager@test.local','hash',false);
	INSERT INTO "Address"("CEP",street,number) VALUES('12345678','test','1');
	INSERT INTO physical_restaurant(restaurant_id,address_id,name,cnpj,owner_cnpj,cnpj_verified) VALUES
	 (1,1,'Branch','12345678000100','12345678000100',true),(1,1,'Pending','12345678000200','12345678000100',false),(2,1,'Other','12345678000300','12345678000100',true);
	INSERT INTO branch_manager(account_id,branch_id) VALUES(3,1);
	INSERT INTO category(restaurant_id,name) VALUES(1,'Food'),(2,'Food');
	INSERT INTO menu_item(restaurant_id,category_id,name,base_price_cents) VALUES(1,1,'Burger',3000),(2,2,'Other burger',3000);
	INSERT INTO loyalty_wallet(account_id,restaurant_id,coins_amount) VALUES(2,1,100);
	INSERT INTO "order"(physical_restaurant_id,account_id,order_type) VALUES(1,2,'local'),(1,2,'local');`)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func TestPostgresCouponRedemption(t *testing.T) {
	conn := openSalesTestDB(t)
	s := NewService(conn)
	ctx := context.Background()
	code, kind, title, image, scope := "burger20", "percent", "Burger Day", "/media/promo.png", "item"
	value := int64(20)
	limit, id, banner := 1, 1, true
	create := request{op: "create", slug: "test", actor: 1, edit: input{Code: &code, Type: &kind, Value: &value, Max: optional[int]{true, &limit}, Banner: &banner, Title: &title, Image: &image, Scope: &scope, Target: optional[int]{true, &id}}}
	v, err := s.execute(ctx, create)
	if err != nil {
		t.Fatal(err)
	}
	coupon := v.(Coupon)
	if coupon.Code != "BURGER20" {
		t.Fatal(coupon.Code)
	}
	if _, err = s.execute(ctx, request{op: "create", slug: "test", actor: 2}); err != errForbidden {
		t.Fatal(err)
	}
	try := func(r request, want error) {
		t.Helper()
		if _, err := s.execute(ctx, r); err != want {
			t.Fatalf("%s: got %v, want %v", r.op, err, want)
		}
	}
	// FIN-009..012, FIN-031: normalization, tenant isolation and target type.
	code = "BuRgEr20"
	try(create, errConflict)
	code = "DENIED"
	create.actor = 3
	try(create, errForbidden)
	create.actor = 1
	foreign := 2
	create.edit.Target.value = &foreign
	try(create, errInvalid)
	create.edit.Target.value = &id
	scope = "combo"
	try(create, errInvalid)
	scope = "item"
	value = 101
	try(create, errInvalid)
	value = 20
	other := create
	other.slug = "other"
	other.edit.Target.value = &foreign
	code = "BURGER20"
	try(other, nil)
	// FIN-032: publication and claims are scoped to the storefront and account.
	try(request{op: "banners", slug: "test", branch: 2}, errMissing)
	try(request{op: "claim", slug: "test", branch: 2, id: coupon.ID}, errMissing)
	try(request{op: "claim", slug: "test", branch: 1, id: coupon.ID}, nil)
	var anonymousClaims int
	if err := conn.QueryRow(`SELECT count(*) FROM coupon_claim`).Scan(&anonymousClaims); err != nil || anonymousClaims != 0 {
		t.Fatalf("anonymous claim persisted: %d %v", anonymousClaims, err)
	}
	for range 2 {
		if _, err = s.execute(ctx, request{op: "claim", slug: "test", actor: 2, id: coupon.ID}); err != nil {
			t.Fatal(err)
		}
	}
	cart := Cart{Selections: []menu.Selection{{ItemID: 1, Quantity: 1}}, Saved: []int{coupon.ID}, Codes: []string{"BURGER20"}, Coins: 100}
	for actor, want := range map[int]int{2: 1, 3: 0} {
		mine, err := s.execute(ctx, request{op: "mine", slug: "test", actor: actor})
		if err != nil || len(mine.([]Coupon)) != want {
			t.Fatalf("wallet coupons for %d: %v %v", actor, mine, err)
		}
	}
	try(request{op: "quote", slug: "test", branch: 1, actor: 3, cart: cart}, errMissing)
	try(request{op: "quote", slug: "test", branch: 1, cart: Cart{Selections: cart.Selections, Coins: 1}}, errForbidden)
	try(request{op: "quote", slug: "other", branch: 3, actor: 2, cart: Cart{Selections: []menu.Selection{{ItemID: 2, Quantity: 1}}, Coins: 1}}, errConflict)
	try(request{op: "quote", slug: "test", branch: 1, actor: 2, cart: Cart{Selections: cart.Selections, Coins: 3001}}, errInvalid)
	if _, err := s.ApplyToOrder(ctx, "test", 1, 1, 2, Cart{Selections: cart.Selections, Codes: cart.Codes, Coins: 101}); err != errConflict {
		t.Fatalf("overdraw: %v", err)
	}
	preview, err := s.execute(ctx, request{op: "quote", slug: "test", branch: 1, actor: 2, cart: cart})
	if err != nil || preview.(Price).Payable != 2300 {
		t.Fatalf("quote: %+v %v", preview, err)
	}
	var uses, coins, claims int
	if err = conn.QueryRow(`SELECT used_count,(SELECT coins_amount FROM loyalty_wallet) FROM coupon WHERE id=$1`, coupon.ID).Scan(&uses, &coins); err != nil || uses != 0 || coins != 100 {
		t.Fatalf("quote changed state: uses=%d coins=%d %v", uses, coins, err)
	}
	price, err := s.ApplyToOrder(ctx, "test", 1, 1, 2, cart)
	if err != nil || price.Payable != 2300 {
		t.Fatalf("finalize: %+v %v", price, err)
	}
	if _, err = s.ApplyToOrder(ctx, "test", 1, 1, 2, cart); err != errConflict {
		t.Fatalf("duplicate settlement: %v", err)
	}
	cart.Coins = 0
	p, err := s.ApplyToOrder(ctx, "test", 1, 2, 2, cart)
	if err != nil || p.Payable != 3000 {
		t.Fatalf("spent coupon: %+v %v", p, err)
	}
	if err = conn.QueryRow(`SELECT (SELECT used_count FROM coupon WHERE id=$1),(SELECT coins_amount FROM loyalty_wallet),(SELECT count(*) FROM coupon_claim)`, coupon.ID).Scan(&uses, &coins, &claims); err != nil || uses != 1 || coins != 0 || claims != 1 {
		t.Fatalf("uses=%d coins=%d claims=%d err=%v", uses, coins, claims, err)
	}
	var recorded int
	if err = conn.QueryRow(`SELECT count(*) FROM order_coupon oc JOIN order_promotion op USING(order_id) JOIN wallet_transaction w USING(order_id) WHERE oc.order_id=1 AND oc.discount_cents=600 AND w.points_delta=-100`).Scan(&recorded); err != nil || recorded != 1 {
		t.Fatalf("incomplete checkout audit: %d %v", recorded, err)
	}
	validate := request{op: "validate", slug: "test", code: "burger20", subtotal: 3000}
	v, err = s.execute(ctx, validate)
	if err != nil || v.(validation).Valid || v.(validation).Reason != "MAX_USES_REACHED" {
		t.Fatalf("wrong public validation for exhausted coupon: %v %v", v, err)
	}
	try(request{op: "deactivate", slug: "test", actor: 1, code: "BURGER20"}, nil)
	v, err = s.execute(ctx, validate)
	if err != nil || v.(validation).Reason != "INACTIVE" {
		t.Fatalf("wrong public validation for deactivated coupon: %v %v", v, err)
	}
	for _, r := range []request{{op: "banners", slug: "test", branch: 1}, {op: "mine", slug: "test", actor: 2}} {
		v, err = s.execute(ctx, r)
		if err != nil || len(v.([]Coupon)) != 0 {
			t.Fatalf("deactivated coupon remains visible in %s: %v %v", r.op, v, err)
		}
	}
}

func TestFIN024And030ConcurrentRedemptions(t *testing.T) {
	for _, coins := range []int64{0, 75} {
		t.Run(fmt.Sprintf("coins_%d", coins), func(t *testing.T) {
			conn := openSalesTestDB(t)
			s := NewService(conn)
			ctx := context.Background()
			if _, err := conn.Exec(`INSERT INTO coupon(restaurant_id,code,discount_type,discount_value,max_uses) VALUES(1,'LAST','fixed_cents',300,1)`); err != nil {
				t.Fatal(err)
			}
			cart := Cart{Selections: []menu.Selection{{ItemID: 1, Quantity: 1}}, Codes: []string{"last"}, Coins: coins}
			type result struct {
				price Price
				err   error
			}
			out := make(chan result, 2)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for order := 1; order <= 2; order++ {
				wg.Go(func() { <-start; p, e := s.ApplyToOrder(ctx, "test", 1, order, 2, cart); out <- result{p, e} })
			}
			close(start)
			wg.Wait()
			close(out)
			var wins, full, conflicts int
			for r := range out {
				switch {
				case r.err == errConflict:
					conflicts++
				case r.err != nil:
					t.Fatal(r.err)
				case r.price.Payable == 2700-coins:
					wins++
				case r.price.Payable == 3000:
					full++
				default:
					t.Fatalf("unexpected payable: %d", r.price.Payable)
				}
			}
			var uses, settlements int
			var balance int64
			if err := conn.QueryRow(`SELECT (SELECT used_count FROM coupon),(SELECT coins_amount FROM loyalty_wallet),(SELECT count(*) FROM order_promotion)`).Scan(&uses, &balance, &settlements); err != nil {
				t.Fatal(err)
			}
			if uses != 1 || wins != 1 || balance != 100-coins || (coins == 0 && (full != 1 || settlements != 2)) || (coins > 0 && (conflicts != 1 || settlements != 1)) {
				t.Fatalf("uses=%d wins=%d balance=%d full=%d conflicts=%d settlements=%d", uses, wins, balance, full, conflicts, settlements)
			}
		})
	}
}
