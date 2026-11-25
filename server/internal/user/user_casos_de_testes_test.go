package user_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"pratello/server/internal/auth"
	"pratello/server/internal/db"
	"pratello/server/internal/restaurants"
	"pratello/server/internal/user"
)

type fixture struct {
	t    *testing.T
	db   *sql.DB
	mux  *http.ServeMux
	auth *auth.Service
}

func setup(t *testing.T) fixture {
	t.Helper()
	dsn := os.Getenv("AUTH_TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL or DATABASE_URL for PostgreSQL integration")
	}
	c, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	c.SetMaxOpenConns(1)
	schema := pq.QuoteIdentifier(fmt.Sprintf("user_test_%d", time.Now().UnixNano()))
	if _, err = c.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := c.Exec(`DROP SCHEMA ` + schema + ` CASCADE`); err != nil {
			t.Error(err)
		}
	})
	if _, err = c.Exec(`SET search_path TO ` + schema); err != nil {
		t.Fatal(err)
	}
	if err = db.Migrate(c); err != nil {
		t.Fatal(err)
	}
	s := auth.NewService(auth.NewSQLStore(c))
	mux := http.NewServeMux()
	auth.NewHandler(s).Routes(mux)
	restaurants.NewHandler(c, s).Routes(mux)
	user.NewHandler(c, s.Identify).Routes(mux)
	return fixture{t, c, mux, s}
}

func (f fixture) account(name string, sysadmin bool) (int, string) {
	f.t.Helper()
	var id int
	email := name + "@users.test"
	if err := f.db.QueryRow(`INSERT INTO account(name,email,password,email_verified,is_sysadmin) VALUES($1,$2,'hash',true,$3) RETURNING id`, name, email, sysadmin).Scan(&id); err != nil {
		f.t.Fatal(err)
	}
	token, err := f.auth.GenerateToken(id, email)
	if err != nil {
		f.t.Fatal(err)
	}
	return id, token
}
func (f fixture) call(method, path, token, body string, want int) []byte {
	f.t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	f.mux.ServeHTTP(w, req)
	if w.Code != want {
		f.t.Fatalf("%s %s: got %d, want %d: %s", method, path, w.Code, want, w.Body.String())
	}
	return w.Body.Bytes()
}
func (f fixture) id(data []byte) int {
	f.t.Helper()
	var v struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(data, &v); err != nil || v.ID < 1 {
		f.t.Fatalf("invalid ID: %s (%v)", data, err)
	}
	return v.ID
}
func (f fixture) org(slug, token string) {
	f.t.Helper()
	f.call("POST", "/restaurants", token, fmt.Sprintf(`{"name":%q,"slug":%q}`, slug, slug), 201)
}
func (f fixture) branch(slug, token, cnpj string) (int, string) {
	f.t.Helper()
	base := "/restaurants/" + slug + "/branches"
	id := f.id(f.call("POST", base, token, fmt.Sprintf(`{"name":"Centro","cnpj":%q,"owner_cnpj":"12345678000199","address":{"cep":"12345678","street":"Rua Bela","number":"12"}}`, cnpj), 201))
	return id, fmt.Sprintf("%s/%d", base, id)
}

func TestIDP020SysadminRoleAuthority(t *testing.T) {
	f := setup(t)
	_, regular := f.account("regular", false)
	_, admin := f.account("sysadmin", true)
	target, _ := f.account("target", false)
	body := fmt.Sprintf(`{"account_id":%d}`, target)
	f.call("POST", "/sysadmins", "", body, 401)
	f.call("POST", "/sysadmins", regular, body, 403)
	f.call("POST", "/administrators", admin, body, 403)
	var promoted bool
	if err := f.db.QueryRow(`SELECT is_sysadmin OR is_administrator FROM account WHERE id=$1`, target).Scan(&promoted); err != nil || promoted {
		t.Fatalf("unauthorized global role granted: %t (%v)", promoted, err)
	}
	f.call("POST", "/sysadmins", admin, body, 201)
	if err := f.db.QueryRow(`SELECT is_sysadmin AND NOT is_administrator FROM account WHERE id=$1`, target).Scan(&promoted); err != nil || !promoted {
		t.Fatalf("sysadmin not granted independently: %t (%v)", promoted, err)
	}
}

func TestIDP027AnalystMembershipDoesNotGrantMutation(t *testing.T) {
	f := setup(t)
	_, owner := f.account("owner", false)
	id, analyst := f.account("analyst", false)
	f.org("cafe", owner)
	body := fmt.Sprintf(`{"account_id":%d}`, id)
	f.call("POST", "/restaurants/cafe/analysts", owner, body, 201)
	f.call("PUT", "/restaurants/cafe", analyst, `{"name":"Changed"}`, 403)
	f.call("POST", "/restaurants/cafe/owners", analyst, body, 403)
	if got := f.call("GET", "/me/restaurants", analyst, "", 200); !strings.Contains(string(got), `"role":"ANALYST"`) {
		t.Fatalf("analyst grant missing: %s", got)
	}
	f.call("DELETE", fmt.Sprintf("/restaurants/cafe/analysts/%d", id), owner, "", 204)
	if got := f.call("GET", "/me/restaurants", analyst, "", 200); strings.TrimSpace(string(got)) != "[]" {
		t.Fatalf("role survived revoke: %s", got)
	}
}

func TestTEN014EmployeeAndManagerMembershipsAreIndependent(t *testing.T) {
	f := setup(t)
	_, owner := f.account("owner", false)
	personID, person := f.account("person", false)
	for _, slug := range []string{"cafe", "bar", "diner"} {
		f.org(slug, owner)
	}
	firstID, first := f.branch("cafe", owner, "12345678000100")
	secondID, second := f.branch("cafe", owner, "12345678000200")
	managerID, managed := f.branch("bar", owner, "12345678000300")
	_, ordinary := f.branch("diner", owner, "12345678000400")
	body := fmt.Sprintf(`{"account_id":%d,"position":"WAITER"}`, personID)
	employee := f.id(f.call("POST", first+"/employees", owner, body, 201))
	f.call("POST", second+"/employees", owner, body, 201)
	f.call("POST", first+"/employees", owner, body, 409)
	f.call("PUT", fmt.Sprintf("%s/managers/%d", managed, personID), owner, "", 204)
	f.call("GET", first+"/tables", person, "", 200)
	f.call("GET", second+"/tables", person, "", 200)
	f.call("POST", first+"/tables", person, `{"number":1}`, 403)
	f.call("PUT", managed, person, `{"name":"Updated"}`, 200)
	f.call("PUT", "/restaurants/bar", person, `{"name":"No"}`, 403)
	f.call("GET", ordinary+"/tables", person, "", 403)
	var assignments []struct {
		Restaurant struct {
			Slug string `json:"slug"`
		} `json:"restaurant"`
		Role      string `json:"role"`
		BranchIDs []int  `json:"branch_ids"`
	}
	if err := json.Unmarshal(f.call("GET", "/me/restaurants", person, "", 200), &assignments); err != nil {
		t.Fatal(err)
	}
	if len(assignments) != 2 || assignments[0].Restaurant.Slug != "bar" || assignments[0].Role != "MANAGER" || fmt.Sprint(assignments[0].BranchIDs) != fmt.Sprint([]int{managerID}) || assignments[1].Restaurant.Slug != "cafe" || assignments[1].Role != "EMPLOYEE" || fmt.Sprint(assignments[1].BranchIDs) != fmt.Sprint([]int{firstID, secondID}) {
		t.Fatalf("wrong memberships: %+v", assignments)
	}
	f.call("DELETE", fmt.Sprintf("%s/employees/%d", second, employee), owner, "", 404)
	f.call("DELETE", fmt.Sprintf("%s/employees/%d", first, employee), owner, "", 204)
	f.call("GET", first+"/tables", person, "", 403)
	f.call("GET", second+"/tables", person, "", 200)
	f.call("PUT", managed, person, `{"name":"Still managed"}`, 200)
}
