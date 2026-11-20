package integration

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"pratello/server/internal/auth"
	"pratello/server/internal/common"
	"pratello/server/internal/db"
	"pratello/server/internal/menu"
	"pratello/server/internal/restaurants"
	"pratello/server/internal/sales"
	"pratello/server/internal/user"
)

func TestJoaoBurgerHTTPFixture(t *testing.T) {
	dsn := os.Getenv("AUTH_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL for PostgreSQL integration")
	}
	redisURL := os.Getenv("AUTH_TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("set AUTH_TEST_REDIS_URL for verification")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	root, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	schema := fmt.Sprintf("joao_fixture_%d", time.Now().UnixNano())
	if _, err = root.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
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
	c, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if err = db.Migrate(c); err != nil {
		t.Fatal(err)
	}
	identity := auth.NewService(auth.NewStore(c))
	closeVerification, err := identity.EnableFixedCodeVerification(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeVerification() })
	mux := http.NewServeMux()
	auth.NewHandler(identity).Routes(mux)
	user.NewHandler(c, identity.Identify).Routes(mux)
	restaurants.NewHandler(c, identity).Routes(mux)
	menu.NewHandler(menu.NewService(c), identity.Identify).Routes(mux)
	sales.NewHandler(sales.NewService(c), identity.Identify).Routes(mux)
	srv := httptest.NewServer(common.ImageFS(Media()).Wrap(mux))
	defer srv.Close()
	api := fixtureAPI{srv.Client(), srv.URL}
	var tokens struct {
		Access string `json:"access_token"`
	}
	if err := api.request(ctx, "POST", "/api/auth/register/email/request-code", "", map[string]any{"email": "bootstrap@joao-burger.test"}, nil, 202); err != nil {
		t.Fatal(err)
	}
	if err := api.request(ctx, "POST", "/api/auth/register", "", map[string]any{"email": "bootstrap@joao-burger.test", "password": "bootstrap-test-password", "code": "123456"}, &tokens, 201); err != nil {
		t.Fatal(err)
	}
	// Seed only the first platform operator directly in this disposable schema.
	// All sample accounts and their role assignments use the real API.
	if _, err := c.ExecContext(ctx, `UPDATE account SET is_administrator=true WHERE email='bootstrap@joao-burger.test'`); err != nil {
		t.Fatal(err)
	}
	f, err := SeedJoaoBurger(ctx, srv.Client(), srv.URL, tokens.Access)
	if err != nil {
		t.Fatal(err)
	}
	get := func(path, token string, dst any) {
		t.Helper()
		if err := api.request(ctx, "GET", path, token, nil, dst, 200); err != nil {
			t.Fatal(err)
		}
	}
	const base = "/api/restaurants/" + JoaoBurgerSlug
	if err := api.request(ctx, "PUT", "/api/me/avatar", f.Employee.AccessToken,
		map[string]string{"avatar_url": "/account/bob.png"}, nil, 200); err != nil {
		t.Fatal(err)
	}
	if err := api.request(ctx, "PUT", "/api/me/avatar", f.Employee.AccessToken,
		map[string]string{"avatar_url": "/media/../secret.png"}, nil, 400); err != nil {
		t.Fatal(err)
	}
	var org struct {
		Name  string
		Icon  string `json:"icon_url"`
		Cover string `json:"cover_url"`
	}
	get(base, "", &org)
	if org.Name != "João Burger" || org.Icon != restaurantMedia+"brand/icon.png" || org.Cover != restaurantMedia+"brand/cover.png" {
		t.Fatalf("wrong branding: %+v", org)
	}
	var layout struct {
		Style     string `json:"template_style"`
		Accent    string `json:"accent_color"`
		Secondary string `json:"secondary_color"`
	}
	get(base+"/menu", "", &layout)
	if layout.Style != "classic" || layout.Accent != "#F46B36" || layout.Secondary != "#FFE135" {
		t.Fatalf("wrong menu layout: %+v", layout)
	}
	var assets []string
	assets = append(assets, org.Icon, org.Cover)
	for _, a := range []FixtureAccount{f.Owner, f.Administrator, f.Analyst, f.Manager, f.Employee} {
		var account struct {
			ID     int
			Name   string
			Avatar string `json:"avatar_url"`
		}
		get("/api/auth/me", a.AccessToken, &account)
		if account.ID != a.ID || account.Name != a.Name || account.Avatar != a.AvatarURL {
			t.Fatalf("wrong account: %+v", account)
		}
		assets = append(assets, account.Avatar)
	}
	var items struct {
		Data []struct {
			ID        int
			Photo     string `json:"photo_url"`
			Combo     bool   `json:"is_combo"`
			Available bool   `json:"is_available"`
		}
	}
	get(base+"/items?page_size=30", "", &items)
	combos := 0
	for _, item := range items.Data {
		assets = append(assets, item.Photo)
		if item.Combo {
			combos++
			if !item.Available {
				t.Fatalf("combo %d is unavailable", item.ID)
			}
		}
	}
	if len(items.Data) != 25 || combos != 10 {
		t.Fatalf("got %d products and %d combos", len(items.Data), combos)
	}
	var banners []struct {
		Image string `json:"background_image_url"`
	}
	get(base+"/menu/banners", "", &banners)
	if len(banners) != 3 {
		t.Fatalf("got %d banners", len(banners))
	}
	for _, b := range banners {
		assets = append(assets, b.Image)
	}
	defaultPNG, err := fs.ReadFile(Media(), "default.png")
	if err != nil {
		t.Fatal(err)
	}
	for _, asset := range assets {
		if !strings.HasPrefix(asset, "/media/") {
			t.Fatalf("unexpected asset URL: %q", asset)
		}
		resp, err := srv.Client().Get(srv.URL + asset)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil || resp.StatusCode != 200 || !bytes.Equal(actual, defaultPNG) {
			t.Fatalf("invalid image %s: HTTP %d (%v)", asset, resp.StatusCode, err)
		}
	}
}
