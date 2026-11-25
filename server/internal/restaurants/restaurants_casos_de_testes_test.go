package restaurants

import (
	"context"
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
	"pratello/server/internal/menu"
	"pratello/server/internal/user"
)

func openTENDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("AUTH_TEST_DATABASE_URL")
	if url == "" {
		url = os.Getenv("DATABASE_URL")
	}
	if url == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL or DATABASE_URL for PostgreSQL integration")
	}
	control, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = control.Close() })
	schemaName := fmt.Sprintf("restaurant_test_%d", time.Now().UnixNano())
	schema := pq.QuoteIdentifier(schemaName)
	if _, err = control.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := control.Exec(`DROP SCHEMA ` + schema + ` CASCADE`); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
	})
	if strings.HasPrefix(url, "postgres://") || strings.HasPrefix(url, "postgresql://") {
		url, err = pq.ParseURL(url)
		if err != nil {
			t.Fatal(err)
		}
	}
	// Set search_path for each new connection, including a replacement connection
	// after cancellation and the second connection in the contention test.
	conn, err := sql.Open("postgres", url+" search_path="+schemaName)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	conn.SetMaxOpenConns(1)
	if err = db.Migrate(conn); err != nil {
		t.Fatal(err)
	}
	return conn
}

func TestIDP019And023RestaurantPostgresLifecycle(t *testing.T) {
	conn := openTENDB(t)
	var err error
	identity := auth.NewService(auth.NewSQLStore(conn))
	mux := http.NewServeMux()
	organizationHandler := NewHandler(conn, identity)
	organizationHandler.Routes(mux)
	menu.NewHandler(menu.NewService(conn), identity.Identify).Routes(mux)
	user.NewHandler(conn, identity.Identify).Routes(mux)

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
	if _, err = conn.Exec(`INSERT INTO account_restaurant(account_id,restaurant_id,account_type) SELECT $1,id,'ANALYST' FROM restaurant WHERE name_slug='cafe'`, analystID); err != nil {
		t.Fatal(err)
	}
	call("PUT", "/restaurants/cafe", analyst, `{"name":"No"}`, 403)
	call("PUT", "/restaurants/cafe", owner, `{"name":"Café Updated"}`, 200)
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
	category := idFrom(call("POST", "/restaurants/cafe/categories", owner, `{"name":"Almoço"}`, 201))
	item := idFrom(call("POST", "/restaurants/cafe/items", owner, fmt.Sprintf(`{"name":"Massa","base_price_cents":1725,"category_id":%d}`, category), 201))
	var cents int
	if err := conn.QueryRow(`SELECT base_price_cents FROM menu_item WHERE id=$1`, item).Scan(&cents); err != nil || cents != 1725 {
		t.Fatalf("catalog price: cents=%d err=%v", cents, err)
	}
	call("GET", "/restaurants/cafe/menu", "", "", 200)
	call("GET", "/restaurants/cafe/dishes/menu", "", "", 404)
	call("POST", "/api/restaurants/cafe/dishes", owner, "{}", 404)
	call("POST", "/api/upload/image", owner, "", 404)
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

const tenAddress = `{"cep":"12345678","street":"Rua Bela","number":"12A"}`
const tenRestaurant = "/restaurants/cafe"
const tenBranches = tenRestaurant + "/branches"

type tenAPI struct {
	t        *testing.T
	db       *sql.DB
	identity *auth.Service
	mux      *http.ServeMux
}

type tenBranch struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	OwnerCNPJ string  `json:"owner_cnpj"`
	Verified  bool    `json:"cnpj_verified"`
	Address   address `json:"address"`
}

type tenPage[T any] struct {
	Data  []T `json:"data"`
	Page  int `json:"page"`
	Size  int `json:"page_size"`
	Total int `json:"total"`
}

func newTEN(t *testing.T) tenAPI {
	t.Helper()
	conn := openTENDB(t)
	identity := auth.NewService(auth.NewSQLStore(conn))
	mux := http.NewServeMux()
	NewHandler(conn, identity).Routes(mux)
	menu.NewHandler(menu.NewService(conn), identity.Identify).Routes(mux)
	user.NewHandler(conn, identity.Identify).Routes(mux)
	return tenAPI{t, conn, identity, mux}
}

func (a tenAPI) account(name string, sysadmin bool) (int, string) {
	a.t.Helper()
	var id int
	email := name + "@restaurant.test"
	err := a.db.QueryRow(`INSERT INTO account(email,password,email_verified,is_sysadmin) VALUES($1,'hash',true,$2) RETURNING id`, email, sysadmin).Scan(&id)
	if err != nil {
		a.t.Fatal(err)
	}
	token, err := a.identity.GenerateToken(id, email)
	if err != nil {
		a.t.Fatal(err)
	}
	return id, token
}

func (a tenAPI) call(method, path, token, body string, want int) []byte {
	a.t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	a.mux.ServeHTTP(w, r)
	if w.Code != want {
		a.t.Fatalf("%s %s: got %d, want %d: %s", method, path, w.Code, want, w.Body.String())
	}
	return w.Body.Bytes()
}

func tenDecode[T any](t *testing.T, data []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("decode %s: %v", data, err)
	}
	return v
}

func (a tenAPI) createOrg(token, slug string) {
	a.t.Helper()
	a.call("POST", "/restaurants", token, fmt.Sprintf(`{"name":"Café","slug":%q}`, slug), 201)
}

func tenBranchBody(cnpj, ownerCNPJ, addr string) string {
	return fmt.Sprintf(`{"name":"Centro","cnpj":%q,"owner_cnpj":%q,"address":%s}`, cnpj, ownerCNPJ, addr)
}

func (a tenAPI) createBranch(token, slug, cnpj string) (int, string) {
	a.t.Helper()
	path := "/restaurants/" + slug + "/branches"
	v := tenDecode[tenBranch](a.t, a.call("POST", path, token, tenBranchBody(cnpj, "12345678000199", tenAddress), 201))
	if v.ID <= 0 || v.Verified {
		a.t.Fatalf("unexpected new branch: %+v", v)
	}
	return v.ID, fmt.Sprintf("%s/%d", path, v.ID)
}

func TestTEN001OrganizationCreation(t *testing.T) {
	a := newTEN(t)
	id, owner := a.account("owner", false)
	_, outsider := a.account("outsider", false)
	a.call("POST", "/restaurants", "", `{"name":"Café","slug":"cafe"}`, 401)
	a.createOrg(owner, "cafe")
	var got int
	if err := a.db.QueryRow(`SELECT m.account_id FROM account_restaurant m JOIN restaurant r ON r.id=m.restaurant_id WHERE r.name_slug='cafe' AND m.account_type='OWNER'`).Scan(&got); err != nil || got != id {
		t.Fatalf("initial owner=%d, want %d: %v", got, id, err)
	}
	for _, tc := range []struct {
		token string
		want  int
	}{{owner, 1}, {outsider, 0}} {
		items := tenDecode[[]struct {
			Role       string `json:"role"`
			Restaurant struct {
				Slug string `json:"slug"`
			} `json:"restaurant"`
		}](t, a.call("GET", "/me/restaurants", tc.token, "", 200))
		if len(items) != tc.want || tc.want == 1 && (items[0].Role != "OWNER" || items[0].Restaurant.Slug != "cafe") {
			t.Fatalf("unexpected memberships: %+v", items)
		}
	}
}

func TestTEN002SlugValidationAndUniqueness(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	_, other := a.account("other", false)
	a.call("POST", "/restaurants", owner, `{"name":"Trattoria","slug":"Invalid slug"}`, 400)
	created := tenDecode[struct {
		Slug string `json:"slug"`
	}](t, a.call("POST", "/restaurants", owner, `{"name":"Trattoria & Cia"}`, 201))
	if created.Slug != "trattoria-cia" {
		t.Fatalf("generated slug: %q", created.Slug)
	}
	a.call("POST", "/restaurants", other, `{"name":"Duplicate","slug":"trattoria-cia"}`, 409)
	a.call("PUT", "/restaurants/trattoria-cia", owner, `{"name":"New","slug":"new-slug"}`, 400)
	a.call("GET", "/restaurants/trattoria-cia", "", "", 200)
	a.call("GET", "/restaurants/new-slug", "", "", 404)
}

func TestTEN003DiscoveryPagination(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	for _, slug := range []string{"gamma", "alpha", "beta"} {
		a.createOrg(owner, slug)
	}
	a.call("DELETE", "/restaurants/alpha", owner, "", 204)
	type item struct {
		Slug string `json:"slug"`
	}
	for i, slug := range []string{"beta", "gamma", ""} {
		page := tenDecode[tenPage[item]](t, a.call("GET", fmt.Sprintf("/restaurants?page=%d&page_size=1", i+1), "", "", 200))
		if page.Total != 2 || page.Page != i+1 || page.Size != 1 || page.Data == nil || (slug == "" && len(page.Data) != 0) || (slug != "" && (len(page.Data) != 1 || page.Data[0].Slug != slug)) {
			t.Fatalf("unexpected page: %+v", page)
		}
	}
	a.call("GET", "/restaurants?page_size=101", "", "", 400)
}

func TestTEN004BranchCreationValidationAndRollback(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	_, other := a.account("other", false)
	a.createOrg(owner, "cafe")
	const cnpj = "12345678000100"
	body := tenBranchBody(cnpj, "12345678000199", tenAddress)
	a.call("POST", tenBranches, other, body, 403)
	a.call("POST", tenBranches, owner, tenBranchBody("12.345.678/0001-00", "12345678000199", tenAddress), 400)
	a.call("POST", tenBranches, owner, tenBranchBody(cnpj, "", tenAddress), 400)
	a.call("POST", tenBranches, owner, tenBranchBody(cnpj, "12345678000199", `{"cep":"123","street":"Rua Bela","number":"12A"}`), 400)
	a.createBranch(owner, "cafe", cnpj)
	a.call("POST", tenBranches, owner, body, 409)
	var branches, addresses int
	if err := a.db.QueryRow(`SELECT (SELECT count(*) FROM physical_restaurant),(SELECT count(*) FROM "Address")`).Scan(&branches, &addresses); err != nil || branches != 1 || addresses != 1 {
		t.Fatalf("branches=%d addresses=%d: %v", branches, addresses, err)
	}
}

func TestTEN005ApprovalAndPublicVisibility(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	adminID, admin := a.account("admin", true)
	a.createOrg(owner, "cafe")
	id, path := a.createBranch(owner, "cafe", "12345678000100")
	a.call("GET", path, "", "", 404)
	a.call("GET", path+"/hours", "", "", 404)
	a.call("GET", tenRestaurant+"/menu", "", "", 404)
	private := tenDecode[tenBranch](t, a.call("GET", path, owner, "", 200))
	if private.OwnerCNPJ != "12345678000199" || private.Verified {
		t.Fatalf("unexpected pending branch: %+v", private)
	}
	a.call("PUT", path+"/verification", owner, "", 403)
	list := tenDecode[tenPage[tenBranch]](t, a.call("GET", tenBranches, "", "", 200))
	if list.Total != 0 || len(list.Data) != 0 {
		t.Fatalf("pending branch leaked: %+v", list)
	}
	a.call("PUT", path+"/verification", admin, "", 200)
	public := tenDecode[map[string]json.RawMessage](t, a.call("GET", path, "", "", 200))
	if _, leaked := public["owner_cnpj"]; leaked || string(public["cnpj_verified"]) != "true" {
		t.Fatalf("public data: %v", public)
	}
	list = tenDecode[tenPage[tenBranch]](t, a.call("GET", tenBranches, "", "", 200))
	if list.Total != 1 || len(list.Data) != 1 || list.Data[0].ID != id || list.Data[0].OwnerCNPJ != "" {
		t.Fatalf("public list: %+v", list)
	}
	a.call("GET", tenRestaurant+"/menu", "", "", 200)
	var audited bool
	if err := a.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM audit_entry WHERE actor_account_id=$1 AND branch_id=$2 AND action='branch.verify' AND actor_role='SYSADMIN')`, adminID, id).Scan(&audited); err != nil || !audited {
		t.Fatalf("missing approval audit: %v", err)
	}
}

func TestTEN006GlobalAdministratorAndRevocation(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	_, second := a.account("second", false)
	rootID, root := a.account("root", false)
	adminID, admin := a.account("admin", false)
	if _, err := a.db.Exec(`UPDATE account SET is_administrator=true WHERE id=$1`, rootID); err != nil {
		t.Fatal(err)
	}
	a.createOrg(owner, "cafe")
	a.createOrg(second, "other")
	grant := fmt.Sprintf(`{"account_id":%d}`, adminID)
	a.call("POST", "/administrators", owner, grant, 403)
	a.call("DELETE", fmt.Sprintf("/administrators/%d", rootID), root, "", 409)
	a.call("POST", "/administrators", root, grant, 201)
	a.call("PUT", tenRestaurant, admin, `{"name":"Atualizado"}`, 200)
	a.call("PUT", "/restaurants/other", admin, `{"name":"Também administrado"}`, 200)
	_, branch := a.createBranch(admin, "other", "12345678000122")
	a.call("PUT", branch+"/verification", admin, "", 200)
	a.call("DELETE", fmt.Sprintf("/administrators/%d", adminID), root, "", 204)
	a.call("PUT", tenRestaurant, admin, `{"name":"Negado"}`, 403)
	got := tenDecode[struct {
		Name string `json:"name"`
	}](t, a.call("GET", tenRestaurant, "", "", 200))
	if got.Name != "Atualizado" {
		t.Fatalf("unexpected restaurant name %q", got.Name)
	}
}

func TestTEN007BranchManagerIsolation(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	managerID, manager := a.account("manager", false)
	a.createOrg(owner, "cafe")
	_, first := a.createBranch(owner, "cafe", "12345678000100")
	_, second := a.createBranch(owner, "cafe", "12345678000200")
	assignment := fmt.Sprintf("%s/managers/%d", first, managerID)
	a.call("PUT", assignment, owner, "", 204)
	a.call("PUT", first, manager, `{"name":"Permitido"}`, 200)
	a.call("PUT", second, manager, `{"name":"Negado"}`, 403)
	a.call("DELETE", first, manager, "", 403)
	a.call("DELETE", assignment, owner, "", 204)
	a.call("PUT", first, manager, `{"name":"Revogado"}`, 403)
	if got := tenDecode[tenBranch](t, a.call("GET", second, owner, "", 200)); got.Name != "Centro" {
		t.Fatalf("other branch modified: %+v", got)
	}
}

func TestTEN008CrossTenantAndChildIDs(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	_, other := a.account("other", false)
	a.createOrg(owner, "cafe")
	a.createOrg(other, "other")
	_, cafe := a.createBranch(owner, "cafe", "12345678000100")
	otherID, foreign := a.createBranch(other, "other", "12345678000200")
	a.call("PUT", fmt.Sprintf("%s/%d", tenBranches, otherID), owner, `{"name":"Wrong slug"}`, 404)
	a.call("PUT", foreign, owner, `{"name":"Wrong account"}`, 403)
	table := tenDecode[struct {
		ID int `json:"id"`
	}](t, a.call("POST", foreign+"/tables", other, `{"number":7}`, 201))
	a.call("DELETE", fmt.Sprintf("%s/tables/%d", cafe, table.ID), owner, "", 404)
	listed := tenDecode[[]struct {
		ID int `json:"id"`
	}](t, a.call("GET", foreign+"/tables", other, "", 200))
	if table.ID <= 0 || len(listed) != 1 || listed[0].ID != table.ID {
		t.Fatalf("foreign table was altered: %+v", listed)
	}
}

func TestTEN009BranchAddressReplacement(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	a.createOrg(owner, "cafe")
	_, path := a.createBranch(owner, "cafe", "12345678000100")
	a.call("PUT", path+"/address", owner, `{"cep":"87654321","street":"Nova Rua","number":"42B","complement":"Fundos"}`, 200)
	a.call("PUT", path+"/address", owner, `{"cep":"999","street":"Errada","number":"1"}`, 400)
	got := tenDecode[tenBranch](t, a.call("GET", path, owner, "", 200))
	if got.Address.CEP != "87654321" || got.Address.Street != "Nova Rua" || got.Address.Number != "42B" || got.Address.Complement != "Fundos" {
		t.Fatalf("unexpected address: %+v", got.Address)
	}
}

func TestTEN010BranchContactLifecycle(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	a.createOrg(owner, "cafe")
	_, branch := a.createBranch(owner, "cafe", "12345678000100")
	path := branch + "/phones"
	a.call("POST", path, owner, `{"phone_number":"+5511999999999","phone_type":"fax"}`, 400)
	a.call("POST", path, owner, `{"phone_number":"abc"}`, 400)
	type phone struct {
		ID     int    `json:"id"`
		Number string `json:"phone_number"`
		Type   string `json:"phone_type"`
	}
	added := tenDecode[phone](t, a.call("POST", path, owner, `{"phone_number":"+5511999999999"}`, 201))
	list := tenDecode[[]phone](t, a.call("GET", path, owner, "", 200))
	if added.ID <= 0 || added.Number != "+5511999999999" || added.Type != "mobile" || len(list) != 1 || list[0] != added {
		t.Fatalf("contact: %+v list: %+v", added, list)
	}
	a.call("GET", path, "", "", 401)
	a.call("DELETE", fmt.Sprintf("%s/%d", path, added.ID), owner, "", 204)
	if list = tenDecode[[]phone](t, a.call("GET", path, owner, "", 200)); len(list) != 0 {
		t.Fatalf("contact not removed: %+v", list)
	}
}

func TestTEN011HoursReplacement(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	a.createOrg(owner, "cafe")
	_, branch := a.createBranch(owner, "cafe", "12345678000100")
	path := branch + "/hours"
	week := []string{`{"weekday":0,"opens_at":"18:00","closes_at":"02:00","is_closed":false}`}
	for day := 1; day < 7; day++ {
		week = append(week, fmt.Sprintf(`{"weekday":%d,"is_closed":true}`, day))
	}
	valid := "[" + strings.Join(week, ",") + "]"
	a.call("PUT", path, owner, valid, 200)
	for _, bad := range []string{"[]", strings.Replace(valid, `"weekday":6`, `"weekday":5`, 1), strings.Replace(valid, `"18:00"`, `"25:00"`, 1), strings.Replace(valid, `"is_closed":false`, `"is_closed":true`, 1)} {
		a.call("PUT", path, owner, bad, 400)
	}
	got := tenDecode[[]hour](t, a.call("GET", path, owner, "", 200))
	if len(got) != 7 || got[0].Opens == nil || *got[0].Opens != "18:00" || got[0].Closes == nil || *got[0].Closes != "02:00" || got[6].Closed == nil || !*got[6].Closed {
		t.Fatalf("hours not preserved: %+v", got)
	}
}

func TestTEN012TableUniquenessAndHistory(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	a.createOrg(owner, "cafe")
	branchID, branch := a.createBranch(owner, "cafe", "12345678000100")
	_, other := a.createBranch(owner, "cafe", "12345678000200")
	path := branch + "/tables"
	type table struct {
		ID     int `json:"id"`
		Number int `json:"number"`
		Seats  int `json:"seats"`
	}
	created := tenDecode[table](t, a.call("POST", path, owner, `{"number":2,"seats":4}`, 201))
	a.call("POST", path, owner, `{"number":2}`, 409)
	a.call("POST", other+"/tables", owner, `{"number":2}`, 201)
	entry := fmt.Sprintf("%s/%d", path, created.ID)
	updated := tenDecode[table](t, a.call("PUT", entry, owner, `{"number":2,"seats":6}`, 200))
	if updated.ID != created.ID || updated.Number != 2 || updated.Seats != 6 {
		t.Fatalf("table update: %+v", updated)
	}
	if _, err := a.db.Exec(`WITH placed AS (INSERT INTO "order"(physical_restaurant_id,order_type) VALUES($1,'local') RETURNING id) INSERT INTO local_order(id,table_id,table_number) SELECT id,$2,2 FROM placed`, branchID, created.ID); err != nil {
		t.Fatal(err)
	}
	a.call("DELETE", entry, owner, "", 204)
	if live := tenDecode[[]table](t, a.call("GET", path, owner, "", 200)); len(live) != 0 {
		t.Fatalf("archived table still listed: %+v", live)
	}
	var number int
	if err := a.db.QueryRow(`SELECT table_number FROM local_order WHERE table_id=$1`, created.ID).Scan(&number); err != nil || number != 2 {
		t.Fatalf("historical table: %d (%v)", number, err)
	}
}

func TestTEN013ArchiveRollbackAndHistory(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	a.createOrg(owner, "cafe")
	branchID, branch := a.createBranch(owner, "cafe", "12345678000100")
	var orderID int
	if err := a.db.QueryRow(`INSERT INTO "order"(physical_restaurant_id,order_type) VALUES($1,'local') RETURNING id`, branchID).Scan(&orderID); err != nil {
		t.Fatal(err)
	}
	if _, err := a.db.Exec(`ALTER TABLE audit_entry ADD CONSTRAINT test_reject_archive CHECK (action <> 'restaurant.archive')`); err != nil {
		t.Fatal(err)
	}
	a.call("DELETE", tenRestaurant, owner, "", 500)
	a.call("GET", tenRestaurant, "", "", 200)
	if _, err := a.db.Exec(`ALTER TABLE audit_entry DROP CONSTRAINT test_reject_archive`); err != nil {
		t.Fatal(err)
	}
	a.call("DELETE", branch, owner, "", 204)
	a.call("GET", branch, owner, "", 404)
	a.call("DELETE", tenRestaurant, owner, "", 204)
	a.call("GET", tenRestaurant, "", "", 404)
	var historicalBranch int
	if err := a.db.QueryRow(`SELECT physical_restaurant_id FROM "order" WHERE id=$1`, orderID).Scan(&historicalBranch); err != nil || historicalBranch != branchID {
		t.Fatalf("order history: branch=%d err=%v", historicalBranch, err)
	}
}

func TestTEN015PoolCancellationAndRecovery(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	a.createOrg(owner, "cafe")
	reserved, err := a.db.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer reserved.Close()
	waits := a.db.Stats().WaitCount
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	w := httptest.NewRecorder()
	a.mux.ServeHTTP(w, httptest.NewRequestWithContext(ctx, http.MethodGet, tenRestaurant, nil))
	if ctx.Err() != context.DeadlineExceeded || w.Code < 400 || a.db.Stats().WaitCount <= waits {
		t.Fatalf("request did not time out in the pool: status=%d context=%v stats=%+v", w.Code, ctx.Err(), a.db.Stats())
	}
	if err := reserved.Close(); err != nil {
		t.Fatal(err)
	}
	fresh, stop := context.WithTimeout(t.Context(), 5*time.Second)
	defer stop()
	w = httptest.NewRecorder()
	a.mux.ServeHTTP(w, httptest.NewRequestWithContext(fresh, http.MethodGet, tenRestaurant, nil))
	if w.Code != http.StatusOK || a.db.Stats().InUse != 0 {
		t.Fatalf("pool did not recover: status=%d body=%s stats=%+v", w.Code, w.Body, a.db.Stats())
	}
}

func TestTEN016OversizedRequestDoesNotPersist(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	body, err := json.Marshal(map[string]string{"name": "Café", "slug": "large", "description": strings.Repeat("x", 32<<10)})
	if err != nil {
		t.Fatal(err)
	}
	data := a.call("POST", "/restaurants", owner, string(body), http.StatusRequestEntityTooLarge)
	if got := tenDecode[map[string]string](t, data)["error"]; got != "request body too large" {
		t.Fatalf("unexpected response: %s", data)
	}
	var count int
	if err := a.db.QueryRow(`SELECT count(*) FROM restaurant WHERE name_slug='large'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("oversized request persisted: count=%d err=%v", count, err)
	}
	a.call("GET", "/restaurants/large", "", "", http.StatusNotFound)
}

func TestTEN017DatabaseErrorDoesNotLeakDetails(t *testing.T) {
	pool, err := sql.Open("postgres", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := pool.Close(); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	NewHandler(pool, nil).Routes(mux)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/restaurants", nil))
	if w.Code != http.StatusInternalServerError || w.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected response: %d %s %s", w.Code, w.Header().Get("Content-Type"), w.Body)
	}
	if got := tenDecode[map[string]string](t, w.Body.Bytes()); len(got) != 1 || got["error"] != "internal server error" {
		t.Fatalf("error response exposed internal details: %s", w.Body)
	}
}

func TestTEN018SchemaReapplyPreservesOrganization(t *testing.T) {
	a := newTEN(t)
	ownerID, owner := a.account("owner", false)
	a.createOrg(owner, "cafe")
	var before int
	if err := a.db.QueryRow(`SELECT id FROM restaurant WHERE name_slug='cafe'`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(a.db); err != nil {
		t.Fatal(err)
	}
	var restaurantID, count int
	err := a.db.QueryRow(`SELECT r.id,count(m.id) FROM restaurant r LEFT JOIN account_restaurant m ON m.restaurant_id=r.id AND m.account_id=$1 AND m.account_type='OWNER' WHERE r.name_slug='cafe' GROUP BY r.id`, ownerID).Scan(&restaurantID, &count)
	if err != nil || restaurantID != before || count != 1 {
		t.Fatalf("initialization changed the organization or owner: id=%d want=%d count=%d err=%v", restaurantID, before, count, err)
	}
	a.call("POST", "/restaurants", owner, `{"name":"Duplicado","slug":"cafe"}`, http.StatusConflict)
}

func TestTEN019ConcurrentTableCreationIsAtomic(t *testing.T) {
	a := newTEN(t)
	_, owner := a.account("owner", false)
	a.createOrg(owner, "cafe")
	branchID, path := a.createBranch(owner, "cafe", "12345678000100")
	a.db.SetMaxOpenConns(2)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	start := make(chan struct{})
	results := make(chan *httptest.ResponseRecorder, 2)
	for range 2 {
		go func() {
			<-start
			r := httptest.NewRequestWithContext(ctx, http.MethodPost, path+"/tables", strings.NewReader(`{"number":7}`))
			r.Header.Set("Authorization", "Bearer "+owner)
			w := httptest.NewRecorder()
			a.mux.ServeHTTP(w, r)
			results <- w
		}()
	}
	close(start)
	statuses := map[int]int{}
	for range 2 {
		select {
		case result := <-results:
			if result.Code != http.StatusCreated && result.Code != http.StatusConflict {
				t.Errorf("unexpected concurrent response: %d %s", result.Code, result.Body)
			}
			statuses[result.Code]++
		case <-ctx.Done():
			t.Fatalf("concurrent table creation did not finish: %v", ctx.Err())
		}
	}
	if statuses[http.StatusCreated] != 1 || statuses[http.StatusConflict] != 1 {
		t.Fatalf("expected one creation and one conflict, got %v", statuses)
	}
	var tables, audits int
	err := a.db.QueryRowContext(ctx, `SELECT
		(SELECT count(*) FROM dining_table WHERE physical_restaurant_id=$1 AND number=7),
		(SELECT count(*) FROM audit_entry WHERE branch_id=$1 AND action='branch.table.add')`, branchID).Scan(&tables, &audits)
	if err != nil || tables != 1 || audits != 1 {
		t.Fatalf("unexpected persisted effects: tables=%d audit=%d err=%v", tables, audits, err)
	}
}
