package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"image/png"
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
	"pratello/server/internal/order"
	"pratello/server/internal/restaurants"
	"pratello/server/internal/review"
	"pratello/server/internal/sales"
	"pratello/server/internal/transaction"
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
	order.NewHandler(order.NewService(c), identity.Identify, "").Routes(mux)
	review.NewHandler(review.NewService(c), identity.Identify).Routes(mux)
	transaction.NewHandler(transaction.NewService(c, transaction.MockProvider{}), identity.Identify).Routes(mux)
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
	var sushi struct{ Name string }
	get("/api/restaurants/"+tanoshiiSlug, "", &sushi)
	if sushi.Name != "Tanoshii Sushi 楽しい" {
		t.Fatalf("wrong sushi restaurant name: %q", sushi.Name)
	}
	var featured struct {
		Name        string
		Price       int `json:"effective_price_cents"`
		Ingredients []struct {
			Removable bool `json:"is_removable"`
		}
	}
	get(fmt.Sprintf("%s/items/%d", base, f.ItemIDs["smash-classico"]), "", &featured)
	if featured.Name != "Hambúrguer Duplo Especial" || featured.Price != 2290 || len(featured.Ingredients) != 8 || featured.Ingredients[0].Removable || !featured.Ingredients[7].Removable {
		t.Fatalf("wrong featured burger: %+v", featured)
	}
	type reviewPage struct {
		Summary struct {
			Count int `json:"review_count"`
		} `json:"summary"`
		Data []struct {
			Rating  int    `json:"rating"`
			Comment string `json:"comment"`
		} `json:"data"`
	}
	var orgReviews reviewPage
	get(base+"/reviews", "", &orgReviews)
	if orgReviews.Summary.Count != 6 || len(orgReviews.Data) != 6 {
		t.Fatalf("wrong restaurant reviews: %+v", orgReviews)
	}
	stars := map[int]int{}
	for _, r := range orgReviews.Data {
		if r.Comment == "" {
			t.Fatal("missing demo restaurant comment")
		}
		stars[r.Rating]++
	}
	if stars[5] != 4 || stars[4] != 2 {
		t.Fatalf("wrong demo star distribution: %v", stars)
	}
	var smashReviews reviewPage
	get(fmt.Sprintf("%s/items/%d/reviews", base, f.ItemIDs["smash-classico"]), "", &smashReviews)
	faces := map[int]int{}
	for _, r := range smashReviews.Data {
		faces[r.Rating]++
	}
	if smashReviews.Summary.Count != 2 || faces[2] != 1 || faces[3] != 1 {
		t.Fatalf("incorrect standalone item reviews: %+v", smashReviews)
	}
	var paidReviews, itemReviews int
	err = c.QueryRowContext(ctx, `SELECT
		(SELECT count(*) FROM restaurant_review r JOIN "order" o ON o.id=r.order_id AND o.account_id=r.account_id
		 WHERE o.physical_restaurant_id=$1 AND o.status='delivered' AND EXISTS
		 (SELECT 1 FROM transaction t WHERE t.order_id=o.id AND t.status='paid' AND t.amount_cents=o.total_cents)),
		(SELECT count(*) FROM item_review r JOIN order_item i ON i.id=r.order_item_id
		 JOIN "order" o ON o.id=i.order_id AND o.account_id=r.account_id
		 WHERE o.physical_restaurant_id=$1 AND o.status='delivered' AND i.parent_order_item_id IS NULL)`, f.BranchID).Scan(&paidReviews, &itemReviews)
	if err != nil || paidReviews != 6 || itemReviews != 13 {
		t.Fatalf("fixture review purchases: restaurant=%d items=%d err=%v", paidReviews, itemReviews, err)
	}
	if err := api.request(ctx, "PUT", "/api/me/avatar", f.Employee.AccessToken,
		map[string]string{"avatar_url": "/account/bob.png"}, nil, 200); err != nil {
		t.Fatal(err)
	}
	if err := api.request(ctx, "PUT", "/api/me/avatar", f.Employee.AccessToken,
		map[string]string{"avatar_url": "/media/../secret.png"}, nil, 400); err != nil {
		t.Fatal(err)
	}
	var org struct {
		Name    string
		Icon    string `json:"icon_url"`
		Cover   string `json:"cover_url"`
		Favicon string `json:"favicon_url"`
	}
	get(base, "", &org)
	if org.Name != "João Burger" || org.Icon != restaurantMedia+"brand/icon.png" || org.Cover != restaurantMedia+"brand/cover.png" {
		t.Fatalf("wrong branding: %+v", org)
	}
	if org.Favicon != restaurantMedia+"brand/favicon.svg" {
		t.Fatalf("wrong favicon: %q", org.Favicon)
	}
	res, err := srv.Client().Get(srv.URL + org.Favicon)
	if err != nil {
		t.Fatal(err)
	}
	var svg struct {
		XMLName xml.Name `xml:"http://www.w3.org/2000/svg svg"`
		Width   int      `xml:"width,attr"`
		Height  int      `xml:"height,attr"`
	}
	err = xml.NewDecoder(res.Body).Decode(&svg)
	_ = res.Body.Close()
	if res.StatusCode != 200 || res.Header.Get("Content-Type") != "image/svg+xml" || err != nil || svg.Width != 48 || svg.Height != 48 {
		t.Fatalf("invalid favicon: status=%d size=%dx%d err=%v", res.StatusCode, svg.Width, svg.Height, err)
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
	for _, a := range []FixtureAccount{f.Owner, f.Administrator, f.Analyst, f.Manager, f.Employee, f.Rogerio} {
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
	get(base+"/items?page_size=100", "", &items)
	combos := 0
	for _, item := range items.Data {
		if !strings.HasPrefix(item.Photo, "https://images.unsplash.com/") {
			t.Fatalf("João Burger item %d has no external photo: %q", item.ID, item.Photo)
		}
		assets = append(assets, item.Photo)
		if item.Combo {
			combos++
			if !item.Available {
				t.Fatalf("combo %d is unavailable", item.ID)
			}
		}
	}
	if len(items.Data) != 32 || combos != 12 {
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
	for _, asset := range assets {
		if strings.HasPrefix(asset, "https://images.unsplash.com/") {
			continue
		}
		if !strings.HasPrefix(asset, "/media/") {
			t.Fatalf("unexpected asset URL: %q", asset)
		}
		resp, err := srv.Client().Get(srv.URL + asset)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		want, readErr := fs.ReadFile(Media(), strings.TrimPrefix(asset, "/media/"))
		if err != nil || readErr != nil || resp.StatusCode != 200 || resp.Header.Get("Content-Type") != "image/png" || !bytes.Equal(actual, want) {
			t.Fatalf("invalid image %s: HTTP %d (%v)", asset, resp.StatusCode, err)
		}
		img, err := png.Decode(bytes.NewReader(actual))
		if err != nil {
			t.Fatalf("invalid PNG %s: %v", asset, err)
		}
		bounds := img.Bounds()
		transparent, visible := false, false
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				_, _, _, alpha := img.At(x, y).RGBA()
				transparent = transparent || alpha == 0
				visible = visible || alpha > 0
			}
		}
		if !transparent || !visible {
			t.Fatalf("expected visible artwork on a transparent background: %s", asset)
		}
	}
}
