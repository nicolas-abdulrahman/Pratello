package restaurants

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"net/http"
	"net/http/httptest"
	"os"
	"pratello/server/internal/auth"
	"pratello/server/internal/db"
	"strings"
	"testing"
	"time"
)

func TestRestaurantPostgresLifecycle(t *testing.T) {
	url := os.Getenv("AUTH_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL for PostgreSQL integration")
	}
	conn, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetMaxOpenConns(1)
	schema := pq.QuoteIdentifier(fmt.Sprintf("restaurant_test_%d", time.Now().UnixNano()))
	if _, err = conn.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatal(err)
	}
	defer conn.Exec(`DROP SCHEMA ` + schema + ` CASCADE`)
	if _, err = conn.Exec(`SET search_path TO ` + schema); err != nil {
		t.Fatal(err)
	}
	if err = db.Migrate(conn); err != nil {
		t.Fatal(err)
	}
	identity := auth.NewService(auth.NewSQLStore(conn))
	mux := http.NewServeMux()
	organizationHandler := NewHandler(conn, identity)
	organizationHandler.Routes(mux)

	account := func(email string) (int, string) {
		t.Helper()
		var id int
		if err := conn.QueryRow(`INSERT INTO account(email,password,email_verified) VALUES($1,'hash',true) RETURNING id`, email).Scan(&id); err != nil {
			t.Fatal(err)
		}
		token, err := identity.GenerateToken(id, email)
		if err != nil {
			t.Fatal(err)
		}
		return id, token
	}
	ownerID, owner := account("owner@restaurant.test")
	otherID, other := account("other@restaurant.test")
	managerID, manager := account("manager@restaurant.test")
	delegateID, delegate := account("delegate@restaurant.test")
	analystID, analyst := account("analyst@restaurant.test")
	adminID, admin := account("sysadmin@restaurant.test")
	if _, err = conn.Exec(`UPDATE account SET is_sysadmin=true WHERE id=$1`, adminID); err != nil {
		t.Fatal(err)
	}
	call := func(method, path, token, body string, want int) []byte {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != want {
			t.Fatalf("%s %s: got %d want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		return w.Body.Bytes()
	}
	idFrom := func(data []byte) int {
		t.Helper()
		var v struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(data, &v); err != nil || v.ID == 0 {
			t.Fatalf("missing ID in %s (%v)", data, err)
		}
		return v.ID
	}
	call("POST", "/api/restaurants", owner, `{"slug":"cafe","name":"Café"}`, 201)
	call("POST", "/restaurants", other, `{"slug":"elsewhere","name":"Elsewhere"}`, 201)
	call("POST", "/restaurants", other, `{"slug":"cafe","name":"Duplicate"}`, 409)
	call("GET", "/restaurants?page=2&page_size=1", "", "", 200)
	if data := call("GET", "/me/restaurants", owner, "", 200); !strings.Contains(string(data), `"OWNER"`) {
		t.Fatalf("owner membership missing: %s", data)
	}
	call("PUT", "/restaurants/cafe", other, `{"name":"No"}`, 403)
	managerBody := fmt.Sprintf(`{"account_id":%d}`, delegateID)
	call("POST", "/restaurants/cafe/managers", other, managerBody, 403)
	call("POST", "/restaurants/cafe/managers", owner, managerBody, 201)
	call("GET", "/restaurants/cafe/managers", owner, "", 200)
	if _, err = conn.Exec(`INSERT INTO account_restaurant(account_id,restaurant_id,account_type) SELECT $1,id,'ANALYST' FROM restaurant WHERE name_slug='cafe'`, analystID); err != nil {
		t.Fatal(err)
	}
	call("PUT", "/restaurants/cafe", analyst, `{"name":"No"}`, 403)
	call("PUT", "/restaurants/cafe", delegate, `{"name":"Café Updated"}`, 200)
	call("DELETE", "/restaurants/cafe", delegate, "", 403)
	call("DELETE", fmt.Sprintf("/restaurants/cafe/managers/%d", delegateID), owner, "", 204)
	call("PUT", "/restaurants/cafe", delegate, `{"name":"Revoked"}`, 403)
	const addr = `{"cep":"12345678","street":"Rua Bela","number":"12A"}`
	create := func(cnpj string) int {
		return idFrom(call("POST", "/restaurants/cafe/branches", owner, `{"cnpj":"`+cnpj+`","owner_cnpj":"12345678000199","name":"Filial","address":`+addr+`}`, 201))
	}
	id := create("12345678000100")
	second := create("12345678000200")
	branch := fmt.Sprintf("/restaurants/cafe/branches/%d", id)
	call("GET", branch, "", "", 404)
	if data := call("GET", "/restaurants/cafe/branches", "", "", 200); !strings.Contains(string(data), `"total":0`) {
		t.Fatalf("pending branches leaked: %s", data)
	}
	if data := call("GET", "/restaurants/cafe/branches", admin, "", 200); !strings.Contains(string(data), `"total":2`) {
		t.Fatalf("pending branches missing from review list: %s", data)
	}
	call("GET", branch, owner, "", 200)
	call("PUT", branch+"/verification", owner, "", 403)
	call("PUT", branch+"/verification", admin, "", 200)
	dish := idFrom(call("POST", "/restaurants/cafe/dishes", owner, `{"name":"Massa","price":17.25,"category":"Almoço"}`, 201))
	var cents int
	if err := conn.QueryRow(`SELECT base_price_cents FROM menu_item WHERE id=$1`, dish).Scan(&cents); err != nil || cents != 1725 {
		t.Fatalf("catalog price: cents=%d err=%v", cents, err)
	}
	call("GET", "/restaurants/cafe/menu", "", "", 200)
	if data := call("GET", branch, "", "", 200); strings.Contains(string(data), "owner_cnpj") {
		t.Fatalf("owner CNPJ leaked: %s", data)
	}
	call("GET", fmt.Sprintf("/restaurants/elsewhere/branches/%d", id), "", "", 404)
	managerPath := fmt.Sprintf("%s/managers/%d", branch, managerID)
	call("PUT", managerPath, other, "", 403)
	call("PUT", managerPath, owner, "", 204)
	call("PUT", fmt.Sprintf("/restaurants/cafe/branches/%d/address", second), manager, addr, 403)
	call("PUT", branch+"/address", manager, addr, 200)
	call("PUT", branch, manager, `{"name":"Novo"}`, 200)
	call("DELETE", branch, manager, "", 403)
	phone := idFrom(call("POST", branch+"/phones", manager, `{"phone_number":"+5511999999999"}`, 201))
	call("GET", branch+"/phones", manager, "", 200)
	call("DELETE", fmt.Sprintf("%s/phones/%d", branch, phone), manager, "", 204)
	call("PUT", branch+"/hours", manager, `[]`, 400)
	var week []string
	for day := 0; day < 7; day++ {
		week = append(week, fmt.Sprintf(`{"weekday":%d,"is_closed":true}`, day))
	}
	call("PUT", branch+"/hours", manager, "["+strings.Join(week, ",")+"]", 200)
	call("GET", branch+"/hours", "", "", 200)
	table := idFrom(call("POST", branch+"/tables", manager, `{"number":1,"seats":4}`, 201))
	call("POST", branch+"/tables", manager, `{"number":1}`, 409)
	tablePath := fmt.Sprintf("%s/tables/%d", branch, table)
	call("PUT", tablePath, manager, `{"number":2,"seats":6}`, 200)
	call("GET", branch+"/tables", manager, "", 200)
	call("DELETE", tablePath, manager, "", 204)
	call("DELETE", tablePath, manager, "", 404)
	call("POST", "/restaurants/cafe/owners", manager, fmt.Sprintf(`{"account_id":%d}`, otherID), 403)
	call("DELETE", managerPath, owner, "", 204)
	call("PUT", branch, manager, `{"name":"No access"}`, 403)
	call("DELETE", fmt.Sprintf("/restaurants/cafe/owners/%d", ownerID), owner, "", 409)
	call("POST", "/restaurants/cafe/owners", owner, fmt.Sprintf(`{"account_id":%d}`, otherID), 201)
	call("GET", "/restaurants/cafe/owners", other, "", 200)
	call("DELETE", fmt.Sprintf("/restaurants/cafe/owners/%d", ownerID), other, "", 204)
	call("DELETE", branch, other, "", 204)
	call("GET", branch, "", "", 404)
	if _, err = conn.Exec(`ALTER TABLE audit_entry ADD CONSTRAINT reject_archive_for_test CHECK (action <> 'restaurant.archive')`); err != nil {
		t.Fatal(err)
	}
	call("DELETE", "/restaurants/cafe", other, "", 500)
	call("GET", "/restaurants/cafe", "", "", 200)
	if _, err = conn.Exec(`ALTER TABLE audit_entry DROP CONSTRAINT reject_archive_for_test`); err != nil {
		t.Fatal(err)
	}
	call("DELETE", "/restaurants/cafe", other, "", 204)
	call("GET", "/restaurants/cafe", "", "", 404)
	var count int
	if err = conn.QueryRowContext(context.Background(), `SELECT count(*) FROM audit_entry WHERE actor_account_id=$1 AND branch_id=$2 AND action='branch.verify' AND actor_role='SYSADMIN'`, adminID, id).Scan(&count); err != nil || count != 1 {
		t.Fatalf("missing approval audit: %d (%v)", count, err)
	}
}
