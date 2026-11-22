package menu

import (
	"context"
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

	"github.com/lib/pq"
	"pratello/server/internal/auth"
	"pratello/server/internal/db"
	"pratello/server/internal/restaurants"
	"pratello/server/internal/user"
)

func openMenuDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("AUTH_TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL or DATABASE_URL for PostgreSQL integration")
	}
	root, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	schema := fmt.Sprintf("menu_test_%d", time.Now().UnixNano())
	quoted := pq.QuoteIdentifier(schema)
	if _, err := root.Exec("CREATE SCHEMA " + quoted); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := root.Exec("DROP SCHEMA " + quoted + " CASCADE"); err != nil {
			t.Error(err)
		}
	})
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
	if err := db.Migrate(conn); err != nil {
		t.Fatal(err)
	}
	return conn
}

func TestMenuPostgresLifecycle(t *testing.T) {
	conn := openMenuDB(t)
	svc := NewService(conn)
	svc.now = func() time.Time { return time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC) }
	identity := auth.NewService(auth.NewSQLStore(conn))
	mux := http.NewServeMux()
	NewHandler(svc, identity.Identify).Routes(mux)
	restaurants.NewHandler(conn, identity).Routes(mux)
	user.NewHandler(conn, identity.Identify).Routes(mux)
	call := func(method, path, token, body string, want int) []byte {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != want {
			t.Errorf("%s %s: got %d want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		return w.Body.Bytes()
	}
	decodeID := func(raw []byte) int {
		t.Helper()
		var v struct{ ID int }
		if err := json.Unmarshal(raw, &v); err != nil || v.ID == 0 {
			t.Fatalf("invalid resource: %s: %v", raw, err)
		}
		return v.ID
	}
	decodeItem := func(raw []byte) itemDetail {
		t.Helper()
		var v itemDetail
		if err := json.Unmarshal(raw, &v); err != nil {
			t.Fatalf("invalid item: %s: %v", raw, err)
		}
		return v
	}
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := conn.Exec(q, args...); err != nil {
			t.Fatal(err)
		}
	}
	id := func(q string, args ...any) int {
		t.Helper()
		var id int
		if err := conn.QueryRow(q, args...).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	org := id(`INSERT INTO restaurant(name_slug,name) VALUES('menu-test','Menu Test') RETURNING id`)
	other := id(`INSERT INTO restaurant(name_slug,name) VALUES('other','Other') RETURNING id`)
	addr := id(`INSERT INTO "Address"("CEP",street,number) VALUES('12345678','Rua','1') RETURNING id`)
	branch := func(org int, cnpj string) int {
		return id(`INSERT INTO physical_restaurant(restaurant_id,address_id,name,cnpj,owner_cnpj) VALUES($1,$2,'Branch',$3,'00000000000000') RETURNING id`, org, addr, cnpj)
	}
	b1, b2, foreignBranch := branch(org, "00000000000001"), branch(org, "00000000000002"), branch(other, "00000000000003")
	account := func(email string) (int, string) {
		i := id(`INSERT INTO account(email,email_verified,password) VALUES($1,true,'hash') RETURNING id`, email)
		token, err := identity.GenerateToken(i, email)
		if err != nil {
			t.Fatal(err)
		}
		return i, token
	}
	ownerID, owner := account("owner@example.test")
	managerID, manager := account("manager@example.test")
	analystID, analyst := account("analyst@example.test")
	adminID, admin := account("admin@example.test")
	employeeID, employee := account("employee@example.test")
	exec(`UPDATE account SET is_administrator=true WHERE id=$1`, adminID)
	exec(`INSERT INTO account_restaurant(account_id,restaurant_id,account_type) VALUES($1,$3,'OWNER'),($2,$3,'ANALYST')`, ownerID, analystID, org)
	exec(`INSERT INTO branch_manager(account_id,branch_id) VALUES($1,$2),($1,$3)`, managerID, b1, foreignBranch)
	exec(`INSERT INTO employee(account_id,physical_restaurant_id,name,position) VALUES($1,$2,'Employee','WAITER')`, employeeID, b1)

	const root = "/restaurants/menu-test"
	query := func(branch int) string { return fmt.Sprintf("?branch_id=%d", branch) }
	call("GET", root+"/menu", "", "", 404)
	call("GET", root+"/menu"+query(b1), manager, "", 200)
	call("GET", root+"/menu"+query(b2), manager, "", 404)
	call("PUT", root+"/menu", owner, `{"template_style":"modern"}`, 400)
	if body := call("PUT", root+"/menu", admin, `{"template_style":"experience","accent_color":"#123456","secondary_color":"#ABCDEF","name":"New Menu"}`, 200); !strings.Contains(string(body), `"secondary_color":"#ABCDEF"`) {
		t.Fatal(string(body))
	}
	exec(`UPDATE physical_restaurant SET cnpj_verified=true WHERE id=$1`, b1)
	call("GET", "/api"+root+"/menu", "", "", 200)
	call("GET", root+"/categories"+query(b2), "", "", 404)
	exec(`UPDATE physical_restaurant SET cnpj_verified=true WHERE id=$1`, b2)
	// CAT-048: verification of one location cannot publish another location.
	exec(`UPDATE physical_restaurant SET is_active=false WHERE id=$1`, b1)
	for _, path := range []string{"/menu", "/items", "/ingredients", "/categories"} {
		call("GET", root+path+query(b1), "", "", 404)
		call("GET", root+path+query(b2), "", "", 200)
		call("GET", root+path+query(b1), manager, "", 200)
	}
	exec(`UPDATE physical_restaurant SET is_active=true WHERE id=$1`, b1)

	foreignCat := id(`INSERT INTO category(restaurant_id,name) VALUES($1,'Other') RETURNING id`, other)
	cat := decodeID(call("POST", root+"/categories", owner, `{"name":"Lunch","display_order":8}`, 201))
	first := decodeID(call("POST", root+"/categories", owner, `{"name":"Specials","display_order":-1}`, 201))
	var categories []category
	if err := json.Unmarshal(call("GET", root+"/categories", "", "", 200), &categories); err != nil || len(categories) != 2 || categories[0].ID != first {
		t.Fatalf("category order: %+v %v", categories, err)
	}
	create := func(cid int, name, extra string) string {
		return fmt.Sprintf(`{"category_id":%d,"name":%q,"base_price_cents":1500%s}`, cid, name, extra)
	}
	call("POST", root+"/items", owner, create(foreignCat, "Other", ""), 404)
	call("POST", root+"/items", analyst, create(cat, "Forbidden", ""), 403)
	call("POST", root+"/items", owner, fmt.Sprintf(`{"category_id":%d,"name":"Missing price"}`, cat), 400)
	burger := decodeID(call("POST", root+"/items", owner, create(cat, "Burger", `,"base_cost_cents":300,"allows_review":false`), 201))
	combo := decodeID(call("POST", root+"/items", owner, create(cat, "Combo", `,"is_combo":true`), 201))
	burgerURL, comboURL := fmt.Sprintf("%s/items/%d", root, burger), fmt.Sprintf("%s/items/%d", root, combo)
	config := `{"display_config":{"dark":true,"colors":{"bg":"#ffffff","primary_color":"#ea580c","secondary_color":"#fde68a","bg_dark":"#431407","primary_color_dark":"#facc15","secondary_color_dark":"#9a3412"},"menu":{"container":"horizontal"},"first_page":[{"id":"hero","type":"header1","img":"/media/hero.png","welcome":{"text":"Olá"},"title":{"text":"Menu","dark":false},"desc":{"text":""},"button1":{"bg":"primary_color","text":{"dark":false}}},{"id":"sales","type":"section","container":"grid","title":{"text":"Promoções"},"display_items":{"by":"category","value":"sales"},"border_color":"primary_color","item_style":{"bg":"secondary_color","dark":false}}]}}`
	call("PUT", root+"/menu", owner, config, 200)
	var savedLayout layout
	if err := json.Unmarshal(call("GET", root+"/menu", "", "", 200), &savedLayout); err != nil {
		t.Fatal(err)
	}
	hero, section := savedLayout.Display.FirstPage[0], savedLayout.Display.FirstPage[1]
	if hero.Button1.BG != "primary_color" || hero.Button1.Text.Dark == nil || *hero.Button1.Text.Dark || hero.Title.Dark == nil || *hero.Title.Dark || section.BorderColor != "primary_color" || section.ItemStyle.Dark == nil || *section.ItemStyle.Dark {
		t.Fatalf("display styles did not round-trip: %+v", savedLayout.Display)
	}
	call("PUT", root+"/menu", owner, strings.Replace(config, `"border_color":"primary_color"`, `"border_color":"#ffffff"`, 1), 400)
	call("PATCH", burgerURL, owner, `{"display_style":{"bg":"secondary_color","title_color":"text","price_color":"primary_color","dark":false}}`, 200)
	call("PATCH", burgerURL, owner, `{"display_style":{"price_color":"lighter_primary_color"}}`, 400)
	for _, path := range []string{burgerURL, root + "/items"} {
		body := string(call("GET", path, "", "", 200))
		if !strings.Contains(body, `"display_style":{"dark":false,"bg":"secondary_color","title_color":"text","price_color":"primary_color"}`) {
			t.Fatalf("style missing from %s: %s", path, body)
		}
	}
	call("PATCH", burgerURL, owner, `{"display_style":null}`, 200)
	if v := decodeItem(call("GET", burgerURL, "", "", 200)); v.DisplayStyle != nil {
		t.Fatal("style was not cleared")
	}
	for _, tok := range []string{"", manager, employee, owner, analyst, admin} {
		detail := decodeItem(call("GET", burgerURL+query(b1), tok, "", 200))
		private := tok == owner || tok == analyst || tok == admin
		if (detail.Cost != nil) != private || detail.Reviews {
			t.Fatalf("wrong private/reviews state: %+v", detail)
		}
		list := call("GET", root+"/items"+query(b1), tok, "", 200)
		if strings.Contains(string(list), `"base_cost_cents"`) != private {
			t.Fatalf("wrong list cost visibility: %s", list)
		}
	}
	// CAT-047: membership revocation takes effect even with an existing token.
	exec(`DELETE FROM account_restaurant WHERE account_id=$1 AND restaurant_id=$2`, analystID, org)
	if v := decodeItem(call("GET", burgerURL, analyst, "", 200)); v.Cost != nil {
		t.Fatal("revoked analyst can still see cost")
	}
	if v := decodeItem(call("PATCH", burgerURL, owner, `{"base_cost_cents":0}`, 200)); v.Cost == nil || *v.Cost != 0 {
		t.Fatal("zero cost was treated as missing")
	}
	call("PATCH", burgerURL, owner, `{"photo_url":"not-a-url"}`, 400)
	call("PUT", burgerURL, owner, create(cat, "Incomplete", ""), 400)
	call("PATCH", burgerURL, owner, `{"base_cost_cents":null}`, 200)
	if detail := decodeItem(call("GET", burgerURL, owner, "", 200)); detail.Cost != nil {
		t.Fatal("cost not cleared")
	}

	ing := decodeID(call("POST", root+"/ingredients", owner, `{"name":"Onions","default_unit":"g"}`, 201))
	// CAT-055: a racing retry must not create a duplicate recipe relation.
	var recipeID int
	statuses := make(chan *httptest.ResponseRecorder, 2)
	var recipeRace sync.WaitGroup
	for range 2 {
		recipeRace.Go(func() {
			r := httptest.NewRequest("POST", burgerURL+"/ingredients", strings.NewReader(fmt.Sprintf(`{"ingredient_id":%d,"quantity":10,"is_removable":true}`, ing)))
			r.Header.Set("Authorization", "Bearer "+owner)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			statuses <- w
		})
	}
	recipeRace.Wait()
	close(statuses)
	counts := map[int]int{}
	for w := range statuses {
		counts[w.Code]++
		if w.Code == 201 {
			recipeID = decodeID(w.Body.Bytes())
		}
	}
	if counts[201] != 1 || counts[409] != 1 || id(`SELECT count(*) FROM item_ingredient WHERE item_id=$1 AND ingredient_id=$2`, burger, ing) != 1 {
		t.Fatalf("recipe race produced wrong status or data: %v", counts)
	}
	call("DELETE", fmt.Sprintf("%s/ingredients/%d", root, ing), owner, "", 409)
	if v := decodeItem(call("GET", comboURL, "", "", 200)); v.Available {
		t.Fatal("empty combo is orderable")
	}
	call("POST", comboURL+"/combo-items", owner, fmt.Sprintf(`{"sub_item_id":%d,"quantity":1}`, combo), 400)
	part := decodeID(call("POST", comboURL+"/combo-items", owner, fmt.Sprintf(`{"sub_item_id":%d,"quantity":2}`, burger), 201))
	// CAT-052: duplicate components cannot silently replace quantity; no nesting.
	call("POST", comboURL+"/combo-items", owner, fmt.Sprintf(`{"sub_item_id":%d,"quantity":9}`, burger), 409)
	nested := decodeID(call("POST", root+"/items", owner, create(first, "Nested", `,"is_combo":true`), 201))
	call("POST", comboURL+"/combo-items", owner, fmt.Sprintf(`{"sub_item_id":%d,"quantity":1}`, nested), 409)
	if v := decodeItem(call("GET", comboURL, "", "", 200)); len(v.Components) != 1 || v.Components[0].Quantity != 2 {
		t.Fatal("failed additions changed combo composition")
	}
	call("PATCH", burgerURL, owner, `{"is_combo":true}`, 409)
	call("POST", comboURL+"/sides", owner, `{"title":"Drink","min_options":2,"max_options":1}`, 400)
	group := decodeID(call("POST", comboURL+"/sides", owner, `{"title":"Drink","min_options":1,"max_options":1}`, 201))
	groupURL := fmt.Sprintf("%s/sides/%d", comboURL, group)
	optionID := decodeID(call("POST", groupURL+"/options", owner, `{"name":"Juice","additional_price_cents":200}`, 201))
	stockURL := func(b, i int) string { return fmt.Sprintf("%s/branches/%d/items/%d/availability", root, b, i) }
	call("PUT", stockURL(b2, burger), manager, `{"is_available":false}`, 403)
	call("PUT", stockURL(foreignBranch, burger), manager, `{"is_available":false}`, 404)
	call("PUT", stockURL(b1, burger), employee, `{"is_available":false}`, 403)
	call("PUT", stockURL(b1, burger), manager, `{"is_available":false}`, 200)
	if decodeItem(call("GET", comboURL+query(b1), "", "", 200)).Available || !decodeItem(call("GET", comboURL+query(b2), "", "", 200)).Available {
		t.Fatal("component availability leaked across branches")
	}
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() { call("PUT", stockURL(b1, burger), manager, `{"is_available":true}`, 200) })
	}
	wg.Wait()
	optionStock := fmt.Sprintf("%s/branches/%d/items/%d/sides/%d/options/%d/availability", root, b1, combo, group, optionID)
	call("PUT", optionStock, manager, `{"is_available":false}`, 200)
	if v := decodeItem(call("GET", comboURL+query(b1), "", "", 200)); v.Available || len(v.Sides) != 1 || v.Sides[0].Options[0].Available {
		t.Fatalf("unfulfillable group: %+v", v)
	}
	call("PUT", optionStock, manager, `{"is_available":true}`, 200)

	const saleBody = `"discount_type":"fixed_cents","discount_value":500,"starts_at":"2026-01-01T00:00:00Z","ends_at":"2026-01-05T00:00:00Z"`
	call("POST", comboURL+"/sales", manager, `{`+saleBody+`}`, 403)
	call("POST", comboURL+"/sales", owner, `{`+saleBody+`}`, 201)
	saleID := decodeID(call("POST", comboURL+"/sales", manager, fmt.Sprintf(`{"branch_id":%d,%s}`, b1, strings.Replace(saleBody, "500", "200", 1)), 201))
	saleURL := fmt.Sprintf("%s/sales/%d", comboURL, saleID)
	// CAT-046/CAT-053: check ancestry and scope all the way down to child IDs.
	foreignItem := id(`INSERT INTO menu_item(restaurant_id,category_id,name,base_price_cents) VALUES($1,$2,'Foreign',100) RETURNING id`, other, foreignCat)
	call("PATCH", fmt.Sprintf("%s/items/%d", root, foreignItem), owner, `{"base_price_cents":1}`, 404)
	call("DELETE", fmt.Sprintf("%s/ingredients/%d", comboURL, recipeID), owner, "", 404)
	call("DELETE", fmt.Sprintf("%s/sides/%d", burgerURL, group), owner, "", 404)
	call("PUT", fmt.Sprintf("%s/sides/%d/options/%d", burgerURL, group, optionID), owner, `{"name":"Moved","additional_price_cents":0}`, 404)
	call("DELETE", fmt.Sprintf("%s/items/%d/combo-items/%d", root, nested, part), owner, "", 404)
	call("DELETE", fmt.Sprintf("%s/sales/%d", burgerURL, saleID)+query(b1), owner, "", 404)
	call("DELETE", saleURL+query(b2), manager, "", 403)
	if price := id(`SELECT base_price_cents FROM menu_item WHERE id=$1`, foreignItem); price != 100 {
		t.Fatalf("foreign item price changed: %d", price)
	}
	for _, invalid := range []string{
		strings.Replace(saleBody, "500", "1501", 1),                                              // CAT-032: fixed discount exceeds base price.
		strings.Replace(strings.Replace(saleBody, "fixed_cents", "percent", 1), "500", "101", 1), // CAT-031.
		strings.Replace(saleBody, "2026-01-05", "2026-01-01", 1),                                 // CAT-022: zero-length validity.
	} {
		call("POST", comboURL+"/sales", owner, "{"+invalid+"}", 400)
	}
	if count := id(`SELECT count(*) FROM item_sale WHERE item_id=$1`, combo); count != 2 {
		t.Fatalf("invalid promotions changed the schedule: got %d records", count)
	}
	if v := decodeItem(call("GET", comboURL+query(b1), "", "", 200)); v.Effective != 1300 {
		t.Fatalf("branch price: %+v", v)
	}
	if v := decodeItem(call("GET", comboURL+query(b2), "", "", 200)); v.Effective != 1000 {
		t.Fatalf("organization price: %+v", v)
	}
	snapshot, err := svc.Snapshot(context.Background(), "menu-test", b1, Selection{ItemID: combo, Quantity: 2, OptionIDs: []int{optionID}, ComponentRemovals: []ComponentRemoval{{ComboItemID: part, Unit: 2, RemoveIngredientIDs: []int{recipeID}}}})
	if err != nil || snapshot.LineTotalCents != 3000 || snapshot.Notes != "Burger #2: without Onions" {
		t.Fatalf("snapshot: %+v %v", snapshot, err)
	}
	if _, err := svc.Snapshot(context.Background(), "menu-test", b1, Selection{ItemID: combo, Quantity: 1, OptionIDs: []int{optionID}, ComponentRemovals: []ComponentRemoval{{ComboItemID: part, Unit: 3, RemoveIngredientIDs: []int{recipeID}}}}); err != ErrInvalid {
		t.Fatalf("accepted a removal for a non-existent combo serving: %v", err)
	}
	call("DELETE", saleURL+query(b1), manager, "", 204)
	if v := decodeItem(call("GET", comboURL+query(b1), "", "", 200)); v.Effective != 1000 {
		t.Fatalf("cancel did not restore org price: %+v", v)
	}

	categoryURL := fmt.Sprintf("%s/categories/%d", root, cat)
	// CAT-049: archived records move as well, and all deletions share one fallback.
	archived := decodeID(call("POST", root+"/items", owner, create(first, "Archived", ""), 201))
	archivedURL := fmt.Sprintf("%s/items/%d", root, archived)
	call("DELETE", archivedURL, owner, "", 204)
	call("DELETE", categoryURL, owner, "", 409)
	if v := decodeItem(call("GET", comboURL, "", "", 200)); v.Category != cat {
		t.Fatal("failed removal moved the item")
	}
	call("DELETE", categoryURL+"?force=true", owner, "", 204)
	fallback := id(`SELECT id FROM category WHERE restaurant_id=$1 AND is_fallback`, org)
	call("DELETE", fmt.Sprintf("%s/categories/%d?force=true", root, first), owner, "", 204)
	call("DELETE", fmt.Sprintf("%s/categories/%d?force=true", root, fallback), owner, "", 409)
	call("GET", archivedURL, owner, "", 404)
	if v := decodeItem(call("GET", comboURL, "", "", 200)); v.Category != fallback || id(`SELECT category_id FROM menu_item WHERE id=$1`, archived) != fallback || id(`SELECT count(*) FROM category WHERE restaurant_id=$1 AND is_fallback`, org) != 1 {
		t.Fatal("items were not moved to the same fallback")
	}
	call("DELETE", groupURL, owner, "", 204) // Also removes branch option overrides.
	call("PATCH", comboURL, owner, `{"name":"Edited combo","base_price_cents":2500}`, 200)
	second, err := svc.Snapshot(context.Background(), "menu-test", b1, Selection{ItemID: combo, Quantity: 2})
	if err != nil || second.ItemName != "Edited combo" || second.LineTotalCents != 4000 || len(second.Options) != 0 {
		t.Fatalf("new snapshot did not reflect edited catalog: %+v %v", second, err)
	}
	call("DELETE", burgerURL, owner, "", 204)
	call("GET", burgerURL, "", "", 404)
	if v := decodeItem(call("GET", comboURL, "", "", 200)); v.Available {
		t.Fatal("archived item leaves bundle available")
	}
	if snapshot.ItemName != "Combo" || snapshot.LineTotalCents != 3000 {
		t.Fatal("prior checkout snapshot changed")
	}
}
