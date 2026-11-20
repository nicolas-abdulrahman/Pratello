package transaction

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"pratello/server/internal/db"
)

type testProvider struct {
	calls atomic.Int32
	fail  atomic.Bool
}

func (p *testProvider) MakePayment(context.Context, Payment) error {
	p.calls.Add(1)
	if p.fail.Load() {
		return errors.New("provider failure")
	}
	return nil
}

func TestPostgresPaymentLifecycle(t *testing.T) {
	dsn := os.Getenv("AUTH_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL for PostgreSQL")
	}
	root, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	schema := fmt.Sprintf("txn_test_%d", time.Now().UnixNano())
	if _, err = root.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatal(err)
	}
	defer root.Exec("DROP SCHEMA " + schema + " CASCADE")
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
	defer conn.Close()
	if err = db.Migrate(conn); err != nil {
		t.Fatal(err)
	}
	_, err = conn.Exec(`INSERT INTO restaurant(name_slug,name) VALUES('test','Test');
	INSERT INTO account(email,password,is_administrator) VALUES('admin@test.local','hash',true),('customer@test.local','hash',false),('other@test.local','hash',false);
	INSERT INTO "Address"("CEP",street,number) VALUES('12345678','test','1');
	INSERT INTO physical_restaurant(restaurant_id,address_id,name,cnpj,owner_cnpj) VALUES(1,1,'Branch','12345678000100','12345678000100');
	INSERT INTO "order"(physical_restaurant_id,account_id,order_type,total_cents) VALUES(1,2,'local',1000),(1,2,'local',1000),(1,2,'local',1000),(1,2,'local',1000);`)
	if err != nil {
		t.Fatal(err)
	}
	provider := &testProvider{}
	s := NewService(conn, provider)
	at := time.Now()
	s.now = func() time.Time { return at }
	ctx := context.Background()
	start := request{op: "start", order: 1, actor: 2, key: "pix-start", start: startInput{"pix", 1000}}
	var wg sync.WaitGroup
	results := make(chan int, 2)
	for range 2 {
		wg.Go(func() {
			v, e := s.execute(ctx, start)
			if e != nil {
				t.Error(e)
				return
			}
			results <- v.(Transaction).ID
		})
	}
	wg.Wait()
	close(results)
	var id, count int
	for got := range results {
		count++
		if id != 0 && id != got {
			t.Fatal("duplicate payment")
		}
		id = got
	}
	if count != 2 || provider.calls.Load() != 0 {
		t.Fatal("PIX creation must not settle funds")
	}
	changed := start
	changed.start.Amount--
	if _, e := s.execute(ctx, changed); e != errConflict {
		t.Fatalf("same key accepted a changed amount: %v", e)
	}
	if _, e := s.execute(ctx, request{op: "get", id: id, actor: 3}); e != errForbidden {
		t.Fatalf("foreign account accessed payment: %v", e)
	}
	if _, e := s.execute(ctx, request{op: "pixGet", id: id, actor: 3}); e != errForbidden {
		t.Fatalf("foreign account read PIX: %v", e)
	}
	v, e := s.execute(ctx, request{op: "pixGet", id: id, actor: 2})
	if e != nil {
		t.Fatal(e)
	}
	pix := v.(Pix)
	png, e := base64.StdEncoding.DecodeString(pix.QR)
	if e != nil || !strings.HasPrefix(string(png), "\x89PNG") || pix.Paid != nil || !strings.HasPrefix(pix.Code, "PRATELLO-MOCK-PIX:") {
		t.Fatal("invalid mock PIX")
	}
	card := request{op: "start", order: 1, actor: 2, key: "card", start: startInput{"credit_card", 1000}}
	if _, e = s.execute(ctx, card); e != errConflict {
		t.Fatalf("pending PIX failed to reserve balance: %v", e)
	}
	confirm := request{op: "confirm", id: id, actor: 2}
	at = pix.Expires // FIN-040: expiration is exclusive.
	if _, e = s.execute(ctx, confirm); e != errConflict {
		t.Fatalf("expired PIX confirmed: %v", e)
	}
	provider.fail.Store(true)
	if _, e = s.execute(ctx, card); e == nil {
		t.Fatal("provider failure was ignored")
	}
	provider.fail.Store(false)
	start.key = "second-pix"
	v, e = s.execute(ctx, start)
	if e != nil {
		t.Fatal(e)
	}
	confirm.id = v.(Transaction).ID
	for range 2 {
		v, e = s.execute(ctx, confirm)
		if e != nil || v.(Transaction).Status != "paid" {
			t.Fatalf("confirm: %v %v", v, e)
		}
	}
	if provider.calls.Load() != 2 {
		t.Fatalf("unexpected provider calls: %d", provider.calls.Load())
	}
	amount := int64(300)
	refund := request{op: "refund", id: confirm.id, actor: 2, key: "refund", refund: refundInput{Amount: &amount}}
	if _, e = s.execute(ctx, refund); e != errForbidden {
		t.Fatalf("customer refunded self: %v", e)
	}
	refund.actor = 1
	for range 2 {
		v, e = s.execute(ctx, refund)
		if e != nil || v.(Transaction).Refunded != 300 {
			t.Fatalf("refund: %v %v", v, e)
		}
	}
	amount = 701
	refund.key = "over-refund"
	if _, e = s.execute(ctx, refund); e != errConflict {
		t.Fatalf("over-refund was accepted: %v", e)
	}
	if provider.calls.Load() != 3 {
		t.Fatalf("unexpected provider calls: %d", provider.calls.Load())
	}
	refund.key = "remaining"
	refund.refund.Amount = nil
	v, e = s.execute(ctx, refund)
	if e != nil || v.(Transaction).Refunded != 1000 || v.(Transaction).Status != "refunded" {
		t.Fatalf("full refund: %v %v", v, e)
	}
	var refunded int
	if e = conn.QueryRow(`SELECT sum(amount_cents) FROM refund WHERE transaction_id=$1`, confirm.id).Scan(&refunded); e != nil || refunded != 1000 || provider.calls.Load() != 4 {
		t.Fatalf("incorrect refunds persisted: %d %v", refunded, e)
	}

	// FIN-037: separate keys cannot charge the same remaining balance twice.
	out := make(chan error, 2)
	gate := make(chan struct{})
	for _, key := range []string{"charge-a", "charge-b"} {
		wg.Go(func() {
			<-gate
			_, err := s.execute(ctx, request{op: "start", order: 2, actor: 2, key: key, start: startInput{"credit_card", 1000}})
			out <- err
		})
	}
	close(gate)
	wg.Wait()
	close(out)
	ok, conflict := 0, 0
	for e := range out {
		switch e {
		case nil:
			ok++
		case errConflict:
			conflict++
		default:
			t.Fatal(e)
		}
	}
	if ok != 1 || conflict != 1 || provider.calls.Load() != 5 {
		t.Fatalf("distinct-key race: ok=%d conflicts=%d provider=%d", ok, conflict, provider.calls.Load())
	}

	// FIN-001..003, FIN-020, FIN-023, FIN-029: valid split and rejected charges.
	charge := func(r request, want error) {
		t.Helper()
		if _, err := s.execute(ctx, r); err != want {
			t.Fatalf("charge %+v: got %v want %v", r, err, want)
		}
	}
	charge(request{op: "start", order: 3, actor: 2, key: "cash", start: startInput{"cash", 400}}, errForbidden)
	charge(request{op: "start", order: 3, actor: 1, key: "cash", start: startInput{"cash", 400}}, nil)
	charge(request{op: "start", order: 3, actor: 2, key: "debit", start: startInput{"debit_card", 600}}, nil)
	charge(request{op: "start", order: 3, actor: 2, key: "over", start: startInput{"debit_card", 1}}, errConflict)
	charge(request{op: "start", order: 3, actor: 2, key: "zero", start: startInput{"debit_card", 0}}, errInvalid)
	charge(request{op: "start", order: 3, actor: 2, key: "negative", start: startInput{"debit_card", -1}}, errInvalid)
	charge(request{op: "start", order: 9999, actor: 2, key: "missing", start: startInput{"pix", 1}}, errMissing)
	v, e = s.execute(ctx, request{op: "start", order: 4, actor: 2, key: "cancel-pix", start: startInput{"pix", 1000}})
	if e != nil {
		t.Fatal(e)
	}
	cancelID := v.(Transaction).ID
	if _, e = conn.Exec(`UPDATE "order" SET status='cancelled' WHERE id=4`); e != nil {
		t.Fatal(e)
	}
	charge(request{op: "start", order: 4, actor: 2, key: "cancel-debit", start: startInput{"debit_card", 1}}, errConflict)
	if _, e = s.execute(ctx, request{op: "confirm", id: cancelID, actor: 2}); e != errConflict || provider.calls.Load() != 7 {
		t.Fatalf("cancelled order reached provider: %v calls=%d", e, provider.calls.Load())
	}
}
