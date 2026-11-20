package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
)

func isolatedDatabase(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = os.Getenv("AUTH_TEST_DATABASE_URL")
	}
	if url == "" {
		t.Skip("set DATABASE_URL or AUTH_TEST_DATABASE_URL for PostgreSQL integration")
	}
	db, err := Open(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)
	schema := pq.QuoteIdentifier(fmt.Sprintf("db_test_%d", time.Now().UnixNano()))
	if _, err = db.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(`DROP SCHEMA ` + schema + ` CASCADE`); err != nil {
			t.Error(err)
		}
	})
	if _, err = db.Exec(`SET search_path TO ` + schema); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestOpen_And_Migrate_Idempotent(t *testing.T) {
	db := isolatedDatabase(t)
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO restaurant(name_slug,name) VALUES('kept','Kept')`); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM restaurant WHERE name_slug='kept'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("baseline reapply: count=%d err=%v", count, err)
	}
}

func TestLegacySchemaRefusedWithoutModifyingData(t *testing.T) {
	db := isolatedDatabase(t)
	if _, err := db.Exec(`CREATE TABLE account (id int, password_hash text); INSERT INTO account VALUES(7,'preserve')`); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err == nil || !strings.Contains(err.Error(), "pre-DBML database layout") {
		t.Fatalf("expected baseline rejection, got %v", err)
	}
	var value string
	if err := db.QueryRow(`SELECT password_hash FROM account WHERE id=7`).Scan(&value); err != nil || value != "preserve" {
		t.Fatalf("legacy row changed: %q %v", value, err)
	}
}

func TestFullOrderLifecycleIntegration(t *testing.T) {
	db := isolatedDatabase(t)
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	id := func(q string, args ...any) int {
		t.Helper()
		var value int
		if err := db.QueryRow(q, args...).Scan(&value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.Exec(q, args...); err != nil {
			t.Fatal(err)
		}
	}
	owner := id(`INSERT INTO account(email,password,email_verified) VALUES('owner@example.test','hash',true) RETURNING id`)
	customer := id(`INSERT INTO account(email,password,email_verified) VALUES('customer@example.test','hash',true) RETURNING id`)
	org := id(`INSERT INTO restaurant(name_slug,name) VALUES('cafe','Café') RETURNING id`)
	exec(`INSERT INTO account_restaurant(account_id,restaurant_id,account_type) VALUES($1,$2,'OWNER')`, owner, org)
	exec(`INSERT INTO menu(restaurant_id) VALUES($1)`, org)
	address := id(`INSERT INTO "Address"("CEP",street,number) VALUES('12345678','Rua Bela','12A') RETURNING id`)
	branch := id(`INSERT INTO physical_restaurant(restaurant_id,address_id,name,cnpj,owner_cnpj,cnpj_verified) VALUES($1,$2,'Centro','12345678000100','12345678000199',true) RETURNING id`, org, address)
	exec(`INSERT INTO restaurant_phone(physical_restaurant_id,phone,phone_type) VALUES($1,'+5511999999999','whatsapp')`, branch)
	category := id(`INSERT INTO category(restaurant_id,name) VALUES($1,'Almoço') RETURNING id`, org)
	item := id(`INSERT INTO menu_item(restaurant_id,category_id,name,base_price_cents,base_cost_cents) VALUES($1,$2,'Massa',6500,2500) RETURNING id`, org, category)
	ingredient := id(`INSERT INTO ingredient(restaurant_id,name,default_unit) VALUES($1,'Cebola','g') RETURNING id`, org)
	exec(`INSERT INTO item_ingredient(item_id,ingredient_id,quantity,measure,is_removable) VALUES($1,$2,30,'g',true)`, item, ingredient)
	side := id(`INSERT INTO item_side(item_id,title,max_options) VALUES($1,'Adicionais',1) RETURNING id`, item)
	exec(`INSERT INTO item_side_option(item_side_id,name,additional_price_cents) VALUES($1,'Parmesão',800)`, side)
	session := id(`INSERT INTO menu_session(restaurant_id,physical_restaurant_id,account_id,channel,device_type) VALUES($1,$2,$3,'counter','mobile') RETURNING id`, org, branch, customer)
	order := id(`INSERT INTO "order"(physical_restaurant_id,account_id,order_type,subtotal_cents,total_cents) VALUES($1,$2,'local',7300,7300) RETURNING id`, branch, customer)
	exec(`INSERT INTO local_order(id,notes) VALUES($1,'Balcão')`, order)
	line := id(`INSERT INTO order_item(order_id,item_id,item_name,unit_price_cents,unit_cost_cents,line_total_cents,selected_options) VALUES($1,$2,'Massa',7300,2500,7300,'[{"name":"Parmesão"}]') RETURNING id`, order, item)
	exec(`UPDATE "order" SET status='delivered' WHERE id=$1`, order)
	exec(`UPDATE menu_session SET converted_order_id=$2 WHERE id=$1`, session, order)
	exec(`INSERT INTO ux_event(session_id,restaurant_id,physical_restaurant_id,event_type,item_id,cart_total_cents) VALUES($1,$2,$3,'order_submitted',$4,7300)`, session, org, branch, item)
	payment := id(`INSERT INTO transaction(order_id,transaction_type,amount_cents,status,created_by_account_id) VALUES($1,'credit_card',7300,'paid',$2) RETURNING id`, order, owner)
	exec(`INSERT INTO receipt(order_id,receipt_code,receipt) VALUES($1,'receipt-1','{"total_cents":7300}')`, order)
	exec(`INSERT INTO refund(transaction_id,amount_cents,reason,created_by_account_id) VALUES($1,800,'adicional indisponível',$2)`, payment, owner)
	exec(`UPDATE transaction SET refunded_cents=800,status='partially_refunded' WHERE id=$1`, payment)
	exec(`INSERT INTO item_review(item_id,order_item_id,account_id,rating) VALUES($1,$2,$3,3)`, item, line, customer)
	exec(`INSERT INTO restaurant_review(restaurant_id,account_id,rating) VALUES($1,$2,5)`, org, customer)
	exec(`UPDATE menu_item SET base_price_cents=7000,base_cost_cents=3000 WHERE id=$1`, item)
	exec(`INSERT INTO menu_item_price_history(restaurant_id,item_id,changed_by_account_id,old_price_cents,new_price_cents,reason) VALUES($1,$2,$3,6500,7000,'base_price_update')`, org, item, owner)
	var price, cost, net, rating int
	var status, option string
	err := db.QueryRow(`SELECT oi.unit_price_cents,oi.unit_cost_cents,p.amount_cents-p.refunded_cents,ir.rating,o.status,oi.selected_options->0->>'name'
		FROM order_item oi JOIN "order" o ON o.id=oi.order_id JOIN transaction p ON p.order_id=o.id JOIN item_review ir ON ir.order_item_id=oi.id
		JOIN menu_session ms ON ms.converted_order_id=o.id WHERE oi.id=$1 AND ms.id=$2`, line, session).Scan(&price, &cost, &net, &rating, &status, &option)
	if err != nil || price != 7300 || cost != 2500 || net != 6500 || rating != 3 || status != "delivered" || option != "Parmesão" {
		t.Fatalf("unexpected completed order: price=%d cost=%d net=%d rating=%d status=%s option=%s err=%v", price, cost, net, rating, status, option, err)
	}
	if _, err := db.Exec(`UPDATE transaction SET refunded_cents=7301 WHERE id=$1`, payment); err == nil {
		t.Fatal("over-refund accepted")
	}
	if _, err := db.Exec(`INSERT INTO "order"(physical_restaurant_id,account_id,customer_cpf,order_type) VALUES($1,$2,'12345678901','local')`, branch, customer); err == nil {
		t.Fatal("order accepted both account and anonymous CPF")
	}
}
