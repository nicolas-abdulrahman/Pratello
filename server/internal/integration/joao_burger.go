// Package integration provides reusable HTTP fixtures for development and tests.
package integration

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed media
var artwork embed.FS

// Media returns the fixture assets for mounting at /media/.
func Media() fs.FS {
	f, err := fs.Sub(artwork, "media")
	if err != nil {
		panic(err)
	}
	return f
}

const JoaoBurgerSlug = "joao-burger"
const restaurantMedia = "/media/joao_burger/"

type FixtureAccount struct {
	ID                                  int
	Name, Email, AvatarURL, AccessToken string
}

type JoaoBurger struct {
	Owner, Administrator, Analyst, Manager, Employee FixtureAccount
	Rogerio                                          FixtureAccount
	BranchID                                         int
	ItemIDs                                          map[string]int
	Password                                         string // Shared demo password.
}

type fixtureAPI struct {
	client *http.Client
	base   string
}

func (a fixtureAPI) request(ctx context.Context, method, path, token string, body, result any, want int) error {
	return a.requestWithKey(ctx, method, path, token, "", body, result, want)
}

func (a fixtureAPI) requestWithKey(ctx context.Context, method, path, token, key string, body, result any, want int) error {
	var input io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		input = bytes.NewReader(raw)
	}
	r, err := http.NewRequestWithContext(ctx, method, a.base+path, input)
	if err != nil {
		return err
	}
	if input != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	if key != "" {
		r.Header.Set("Idempotency-Key", key)
	}
	w, err := a.client.Do(r)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer w.Body.Close()
	if w.StatusCode != want {
		raw, _ := io.ReadAll(io.LimitReader(w.Body, 2048))
		return fmt.Errorf("%s %s: got %d, want %d: %s", method, path, w.StatusCode, want, raw)
	}
	if result != nil {
		if err := json.NewDecoder(w.Body).Decode(result); err != nil {
			return fmt.Errorf("%s %s: %w", method, path, err)
		}
	} else {
		_, _ = io.Copy(io.Discard, w.Body)
	}
	return nil
}

// SeedJoaoBurger creates the sample restaurant through HTTP on a clean fixture
// server. bootstrapToken is a platform administrator's access token. Mount Media at
// /media/ (or point the server at the fixture media directory) before seeding.
func SeedJoaoBurger(ctx context.Context, client *http.Client, baseURL, bootstrapToken string) (JoaoBurger, error) {
	if bootstrapToken == "" {
		return JoaoBurger{}, fmt.Errorf("João Burger fixture requires a platform administrator token")
	}
	return seedJoaoBurger(ctx, client, baseURL, bootstrapToken, nil)
}

// SeedJoaoBurgerFresh bootstraps the first administrator on a reset database.
// All fixture registrations and subsequent assignments still go through HTTP.
func SeedJoaoBurgerFresh(ctx context.Context, client *http.Client, baseURL string, bootstrap func(context.Context, int) error) (JoaoBurger, error) {
	if bootstrap == nil {
		return JoaoBurger{}, fmt.Errorf("missing administrator bootstrap")
	}
	return seedJoaoBurger(ctx, client, baseURL, "", bootstrap)
}

func seedJoaoBurger(ctx context.Context, client *http.Client, baseURL, bootstrapToken string, bootstrap func(context.Context, int) error) (JoaoBurger, error) {
	if client == nil {
		client = http.DefaultClient
	}
	a := fixtureAPI{client, strings.TrimRight(baseURL, "/")}
	f := JoaoBurger{ItemIDs: map[string]int{}, Password: "password"}
	for _, p := range []struct {
		account        *FixtureAccount
		name, key, img string
	}{
		{&f.Owner, "Geraldo do Rio", "geraldo-do-rio", "https://static.independent.co.uk/s3fs-public/thumbnails/image/2018/09/26/15/man-working-bored.jpg"},
		{&f.Administrator, "Mr Puzzles", "mr-puzzles", "https://i.redd.it/qj5ozbyqtpvc1.png"},
		{&f.Analyst, "Meg", "meg-from-splatoon", "https://static.wikia.nocookie.net/supermarioglitchy4/images/2/2f/Meggy_re-redesign_official_render.png"},
		{&f.Manager, "Petros", "petros-barreto", "https://hrnews.co.uk/wp-content/uploads/2024/03/Image-582-x-420-px-2024-03-14T094701.910.png"},
		{&f.Employee, "Bob", "bob", "https://thumbs.dreamstime.com/b/average-construction-worker-6828347.jpg"},
		{&f.Rogerio, "Rogério", "rogerio", "https://thumbs.dreamstime.com/b/average-construction-worker-6828347.jpg"},
	} {
		p.account.Name, p.account.Email, p.account.AvatarURL = p.name, p.key+"@joao-burger.test", p.img
		var tokens struct {
			Access string `json:"access_token"`
		}
		if err := a.request(ctx, "POST", "/api/auth/register/email/request-code", "", map[string]any{"email": p.account.Email}, nil, 202); err != nil {
			return f, err
		}
		if err := a.request(ctx, "POST", "/api/auth/register", "", map[string]any{"name": p.name, "email": p.account.Email, "password": f.Password, "code": "123456"}, &tokens, 201); err != nil {
			return f, err
		}
		p.account.AccessToken = tokens.Access
		if err := a.request(ctx, "PUT", "/api/me/avatar", tokens.Access, map[string]any{"avatar_url": p.account.AvatarURL}, p.account, 200); err != nil {
			return f, err
		}
	}
	post := func(path, token string, body any, result any) error {
		return a.request(ctx, "POST", path, token, body, result, 201)
	}
	owner := f.Owner.AccessToken
	if bootstrap != nil {
		if err := bootstrap(ctx, f.Administrator.ID); err != nil {
			return f, err
		}
	} else {
		if err := post("/api/administrators", bootstrapToken, map[string]int{"account_id": f.Administrator.ID}, nil); err != nil {
			return f, err
		}
	}
	if err := post("/api/administrators", f.Administrator.AccessToken, map[string]int{"account_id": f.Rogerio.ID}, nil); err != nil {
		return f, err
	}
	const joaoLogo = "https://static.vecteezy.com/system/resources/previews/048/115/352/non_2x/hamburger-food-icon-illustration-vector.jpg"
	if err := post("/api/restaurants", owner, map[string]any{"name": "João Burger", "slug": JoaoBurgerSlug, "description": "Smash burgers, fritas crocantes e combos para dividir.", "icon_url": joaoLogo, "cover_url": restaurantMedia + "brand/cover.png", "favicon_url": restaurantMedia + "brand/favicon.svg"}, nil); err != nil {
		return f, err
	}
	root := "/api/restaurants/" + JoaoBurgerSlug
	var branch struct{ ID int }
	if err := post(root+"/branches", owner, map[string]any{"name": "João Burger — Centro", "cnpj": "12345678000195", "owner_cnpj": "12345678000195", "address": map[string]string{"cep": "50010000", "street": "Rua do Bom Jesus", "number": "77", "city": "Recife", "state": "PE"}}, &branch); err != nil {
		return f, err
	}
	f.BranchID = branch.ID
	branchPath := fmt.Sprintf("%s/branches/%d", root, branch.ID)
	for _, r := range []struct {
		method, path, token string
		body                any
		status              int
	}{
		{"PUT", branchPath + "/verification", f.Administrator.AccessToken, nil, 200},
		{"POST", root + "/analysts", owner, map[string]int{"account_id": f.Analyst.ID}, 201},
		{"PUT", fmt.Sprintf("%s/managers/%d", branchPath, f.Manager.ID), owner, nil, 204},
		{"POST", branchPath + "/employees", f.Manager.AccessToken, map[string]any{"account_id": f.Employee.ID, "position": "WAITER"}, 201},
		{"PUT", root + "/menu", owner, map[string]any{"name": "João Burger — Cardápio", "accent_color": "#EA580C", "secondary_color": "#FDE68A", "display_config": joaoDisplay}, 200},
	} {
		if err := a.request(ctx, r.method, r.path, r.token, r.body, nil, r.status); err != nil {
			return f, err
		}
	}
	categories := map[string]int{}
	categoryPhotos := map[string]string{
		"Burgers": joaoFeatured["smash-classico"].photo, "Acompanhamentos": joaoFeatured["fritas"].photo,
		"Bebidas": joaoCatalogPhotos["suco-laranja"], "Sobremesas": joaoFeatured["brownie"].photo,
		"Combos": joaoFeatured["combo-pra-dois"].photo, "Pizzas": joaoCatalogPhotos["pizza-margherita"],
	}
	for n, name := range []string{"Burgers", "Acompanhamentos", "Bebidas", "Sobremesas", "Combos", "Pizzas"} {
		var v struct{ ID int }
		if err := post(root+"/categories", owner, map[string]any{"name": name, "display_order": n, "image_url": categoryPhotos[name]}, &v); err != nil {
			return f, err
		}
		categories[name] = v.ID
	}
	add := func(v itemSeed, combo bool) error {
		var result struct{ ID int }
		photo := joaoCatalogPhotos[v.key]
		if featured, ok := joaoFeatured[v.key]; ok {
			v.name, v.description, v.price, photo = featured.name, featured.description, featured.original, featured.photo
		}
		body := map[string]any{"category_id": categories[v.category], "name": v.name, "base_price_cents": v.price, "is_combo": combo, "photo_url": photo}
		if !combo && v.description != "" {
			body["description"] = v.description
		}
		if err := post(root+"/items", owner, body, &result); err != nil {
			return err
		}
		f.ItemIDs[v.key] = result.ID
		return nil
	}
	for _, v := range joaoItems {
		if err := add(v, false); err != nil {
			return f, err
		}
	}
	for _, v := range joaoCombos {
		if err := add(itemSeed{v.key, v.name, "Combos", v.description, v.price}, true); err != nil {
			return f, err
		}
		for _, part := range v.parts {
			if err := post(fmt.Sprintf("%s/items/%d/combo-items", root, f.ItemIDs[v.key]), owner, map[string]int{"sub_item_id": f.ItemIDs[part.key], "quantity": part.quantity}, nil); err != nil {
				return f, err
			}
		}
		if sale := joaoComboSales[v.key]; sale > 0 {
			if err := post(fmt.Sprintf("%s/items/%d/sales", root, f.ItemIDs[v.key]), owner, map[string]any{"discount_type": "fixed_cents", "discount_value": v.price - sale, "starts_at": "2020-01-01T00:00:00Z", "ends_at": "2100-01-01T00:00:00Z"}, nil); err != nil {
				return f, err
			}
		}
	}
	if err := a.seedJoaoFeatured(ctx, root, owner, f.ItemIDs); err != nil {
		return f, err
	}
	for _, v := range []struct {
		code, title, description, scope, target, kind, banner string
		value, minimum                                        int
	}{
		{"SMASH15", "DIA DO SMASH", "15% de desconto no Hambúrguer Duplo Especial.", "item", "smash-classico", "percent", "https://img.cdndsgni.com/preview/10705484-m.jpg", 15, 0},
		{"DUPLA8", "BORA DIVIDIR?", "R$ 8,00 de desconto no Combo Casal.", "combo", "combo-pra-dois", "fixed_cents", "https://t3.ftcdn.net/jpg/02/94/30/58/360_F_294305868_QTSSjWvyGvUCPfuH7bPuq6tBqF08hT0x.jpg", 800, 0},
		{"CHAPA10", "PEDIDO GRANDE, DESCONTO JUNTO", "R$ 10,00 de desconto a partir de R$ 80,00.", "order", "", "fixed_cents", "https://img.cdndsgni.com/preview/10024884.jpg", 1000, 8000},
	} {
		body := map[string]any{"code": v.code, "name": v.title, "title": v.title, "description": v.description, "target_scope": v.scope, "discount_type": v.kind, "discount_value": v.value, "min_order_cents": v.minimum, "is_banner": true, "background_image_url": v.banner}
		if v.target != "" {
			body["target_item_id"] = f.ItemIDs[v.target]
		}
		if err := post(root+"/coupons", owner, body, nil); err != nil {
			return f, err
		}
	}
	if err := a.seedJoaoPurchases(ctx, branchPath, f); err != nil {
		return f, err
	}
	return f, a.seedTanoshii(ctx, f.Administrator.AccessToken, f.Password)
}

type itemFeedback struct {
	key     string
	rating  int
	comment string
}

type purchaseFeedback struct {
	buyer   FixtureAccount
	stars   int
	comment string
	items   []itemFeedback
}

func (a fixtureAPI) seedJoaoPurchases(ctx context.Context, branchPath string, f JoaoBurger) error {
	for _, v := range []purchaseFeedback{
		{f.Rogerio, 5, "Atendimento excelente e tudo chegou quentinho. Voltarei com certeza!", []itemFeedback{{"smash-classico", 3, "Carne suculenta, bordas crocantes e molho na medida."}, {"fritas", 3, "Sequinhas e bem crocantes."}, {"brownie", 3, "Chocolate intenso e miolo macio."}}},
		{f.Owner, 5, "O pedido saiu rápido e veio bem montado.", []itemFeedback{{"combo-pra-dois", 3, "Porção na medida e hambúrguer no ponto."}, {"suco-laranja", 2, "Refrescante, mas prefiro com mais gelo."}}},
		{f.Administrator, 4, "Comida muito boa. No horário de pico, a espera foi um pouco maior.", []itemFeedback{{"bacon-bbq", 3, "Bacon crocante e barbecue equilibrado."}, {"aneis-de-cebola", 2, "Gostosos, mas poderiam estar mais crocantes."}}},
		{f.Analyst, 5, "O pedido chegou bem embalado e quentinho.", []itemFeedback{{"combo-pra-dois", 3, "Porção generosa para dividir."}, {"pizza-pepperoni", 3, "Massa leve e pepperoni bem temperado."}}},
		{f.Manager, 4, "O pessoal foi atencioso e conferiu o pedido antes de servir.", []itemFeedback{{"duplo-da-chapa", 3, "Dois smashes bem feitos, com bastante queijo."}, {"fritas", 2, "O sabor estava bom, mas chegaram mornas."}}},
		{f.Employee, 5, "Tudo certinho com o pedido. Gostei bastante do frango!", []itemFeedback{{"frango-crocante", 3, "Frango grelhado bem temperado e suculento."}, {"smash-classico", 2, "Estava gostoso, mas prefiro um pouco mais de molho."}}},
	} {
		if err := a.seedJoaoPurchase(ctx, branchPath, f, v); err != nil {
			return fmt.Errorf("demo purchase for %s: %w", v.buyer.Name, err)
		}
	}
	return nil
}

func (a fixtureAPI) seedJoaoPurchase(ctx context.Context, branchPath string, f JoaoBurger, v purchaseFeedback) error {
	bought := make([]map[string]int, 0, len(v.items))
	for _, item := range v.items {
		id := f.ItemIDs[item.key]
		if id == 0 {
			return fmt.Errorf("unknown fixture item %q", item.key)
		}
		quantity := 1
		switch item.key {
		case "smash-classico":
			quantity = 14
		case "fritas":
			quantity = 10
		case "combo-pra-dois":
			quantity = 7
		}
		bought = append(bought, map[string]int{"item_id": id, "quantity": quantity})
	}
	var order struct {
		ID    int
		Total int64 `json:"total_cents"`
		Items []struct {
			ID     int
			ItemID int  `json:"item_id"`
			Parent *int `json:"parent_order_item_id"`
		}
	}
	token := v.buyer.AccessToken
	if err := a.request(ctx, "POST", branchPath+"/orders/local", token, map[string]any{"items": bought}, &order, 201); err != nil {
		return err
	}
	path := fmt.Sprintf("/api/orders/%d", order.ID)
	var payment struct{ Status string }
	if err := a.requestWithKey(ctx, "POST", path+"/transactions", token, "joao-burger-demo-purchase", map[string]any{"payment_method": "credit_card", "amount_cents": order.Total}, &payment, 201); err != nil {
		return err
	}
	if payment.Status != "paid" {
		return fmt.Errorf("mock payment has status %q", payment.Status)
	}
	for _, status := range []string{"preparing", "ready", "delivered"} {
		if err := a.request(ctx, "PATCH", path+"/status", f.Owner.AccessToken, map[string]string{"status": status}, nil, 200); err != nil {
			return err
		}
	}
	if err := a.request(ctx, "POST", path+"/review", token, map[string]any{"rating": v.stars, "comment": v.comment}, nil, 200); err != nil {
		return err
	}
	lines := map[int]int{}
	for _, line := range order.Items {
		if line.Parent == nil {
			lines[line.ItemID] = line.ID
		}
	}
	for _, item := range v.items {
		id := lines[f.ItemIDs[item.key]]
		if id == 0 {
			return fmt.Errorf("missing purchased line for %q", item.key)
		}
		if err := a.request(ctx, "POST", fmt.Sprintf("%s/items/%d/review", path, id), token, map[string]any{"rating": item.rating, "comment": item.comment}, nil, 200); err != nil {
			return err
		}
	}
	return nil
}

type itemSeed struct {
	key, name, category, description string
	price                            int
}
type portion struct {
	key      string
	quantity int
}
type comboSeed struct {
	key, name, description string
	price                  int
	parts                  []portion
}

var joaoItems = []itemSeed{
	{"cola-2l", "Coca-Cola 2L", "Bebidas", "", 1390},
	{"agua", "Água Mineral", "Bebidas", "Garrafa de 500 ml.", 450},
	{"salada-caesar", "Salada Caesar", "Acompanhamentos", "Alface, croutons, parmesão e molho Caesar.", 1590},
	{"frango-teriyaki", "Sanduíche de Frango Teriyaki", "Burgers", "Frango grelhado com molho teriyaki.", 2490},
	{"batata-rustica", "Batata Rústica", "Acompanhamentos", "Batatas rústicas douradas.", 1390},
	{"hamburguer-kids", "Hambúrguer Pequeno", "Burgers", "Hambúrguer com queijo em pão pequeno.", 1590},
	{"fritas-pequenas", "Batata Frita Pequena", "Acompanhamentos", "Porção pequena de batatas crocantes.", 890},
	{"fritas-medias", "Batata Frita Média", "Acompanhamentos", "Porção média de batatas crocantes.", 1190},
	{"smash-classico", "Smash Clássico", "Burgers", "Smash de 90 g, cheddar e molho da casa no brioche.", 2490},
	{"duplo-da-chapa", "Cheese Burger", "Burgers", "Pão brioche, carne 180g, queijo cheddar e molho especial.", 2490},
	{"bacon-bbq", "Bacon Burger", "Burgers", "Carne 180g, muito bacon crocante e queijo prato.", 2990},
	{"frango-crocante", "Frango Crocante", "Burgers", "Frango empanado, alface, picles e maionese de limão.", 2790},
	{"veggie-da-casa", "Veggie Burger", "Burgers", "Hambúrguer de grão de bico, alface e tomate.", 2290},
	{"fritas", "Fritas da Chapa", "Acompanhamentos", "Batatas finas com sal da casa, 150 g.", 1200},
	{"aneis-de-cebola", "Anéis de Cebola", "Acompanhamentos", "Oito anéis empanados com molho da casa.", 1400},
	{"cola", "Coca-Cola Lata", "Bebidas", "", 600},
	{"suco-laranja", "Suco de Laranja", "Bebidas", "Suco de laranja espremido na hora.", 990},
	{"brownie", "Brownie da Chapa", "Sobremesas", "Brownie com chocolate meio amargo.", 1190},
	{"pizza-pepperoni", "Pizza Pepperoni", "Pizzas", "Massa artesanal, molho de tomate e pepperoni selecionado.", 3990},
	{"pizza-margherita", "Pizza Margherita", "Pizzas", "Molho de tomate, muçarela e manjericão.", 4500},
}

var joaoCombos = []comboSeed{
	{"combo-pra-dois", "Combo Casal", "", 7500, []portion{{"smash-classico", 2}, {"fritas", 1}, {"cola", 2}}},
	{"combo-solo", "Combo Solo", "", 4200, []portion{{"duplo-da-chapa", 1}, {"fritas-medias", 1}, {"cola", 1}}},
	{"combo-familia", "Combo Família", "", 13500, []portion{{"duplo-da-chapa", 4}, {"fritas", 2}, {"cola-2l", 1}}},
	{"combo-veggie", "Combo Veggie", "", 3890, []portion{{"veggie-da-casa", 1}, {"fritas", 1}, {"suco-laranja", 1}}},
	{"combo-kids", "Combo Kids", "", 2590, []portion{{"hamburguer-kids", 1}, {"fritas-pequenas", 1}, {"suco-laranja", 1}}},
	{"combo-fitness", "Combo Fitness", "", 2990, []portion{{"frango-crocante", 1}, {"salada-caesar", 1}, {"agua", 1}}},
	{"combo-duplo", "Combo Duplo", "", 5800, []portion{{"duplo-da-chapa", 2}, {"fritas-medias", 1}, {"cola", 1}}},
	{"combo-bacon", "Combo Bacon Lover", "", 5200, []portion{{"bacon-bbq", 1}, {"fritas", 1}, {"cola", 1}}},
	{"combo-doce", "Combo Doce Final", "", 3690, []portion{{"duplo-da-chapa", 1}, {"brownie", 1}, {"cola", 1}}},
	{"combo-asiatico", "Combo Asiático", "", 4000, []portion{{"frango-teriyaki", 1}, {"batata-rustica", 1}}},
	{"combo-executivo", "Combo Executivo", "", 3390, []portion{{"duplo-da-chapa", 1}, {"salada-caesar", 1}, {"suco-laranja", 1}}},
	{"combo-master", "Combo Master", "", 11000, []portion{{"duplo-da-chapa", 3}, {"fritas", 2}, {"cola", 3}}},
}

var joaoComboSales = map[string]int{"combo-solo": 3290, "combo-familia": 9990, "combo-duplo": 4590, "combo-bacon": 4190, "combo-asiatico": 3490, "combo-master": 8590}

const comboPhotoQuery = "?auto=format&fit=crop&q=80&w=400&h=300"

var joaoCatalogPhotos = map[string]string{
	"combo-solo":       "https://images.unsplash.com/photo-1561758033-d89a9ad46330" + comboPhotoQuery,
	"combo-familia":    "https://images.unsplash.com/photo-1550547660-d9450f859349" + comboPhotoQuery,
	"combo-veggie":     "https://images.unsplash.com/photo-1520072959219-c595dc870360" + comboPhotoQuery,
	"combo-kids":       "https://images.unsplash.com/photo-1568901346375-23c9450c58cd" + comboPhotoQuery,
	"combo-fitness":    "https://images.unsplash.com/photo-1521390188846-e2a3a97453a0" + comboPhotoQuery,
	"combo-duplo":      "https://images.unsplash.com/photo-1568901346375-23c9450c58cd" + comboPhotoQuery,
	"combo-bacon":      "https://images.unsplash.com/photo-1553979459-d2229ba7433b" + comboPhotoQuery,
	"combo-doce":       "https://images.unsplash.com/photo-1568901346375-23c9450c58cd" + comboPhotoQuery,
	"combo-asiatico":   "https://images.unsplash.com/photo-1521390188846-e2a3a97453a0" + comboPhotoQuery,
	"combo-executivo":  "https://images.unsplash.com/photo-1594212699903-ec8a3eca50f5" + comboPhotoQuery,
	"combo-master":     "https://images.unsplash.com/photo-1550547660-d9450f859349" + comboPhotoQuery,
	"cola-2l":          "https://images.unsplash.com/photo-1622483767028-3f66f32aef97" + photoQuery,
	"agua":             "https://images.unsplash.com/photo-1548839140-29a749e1cf4d" + photoQuery,
	"salada-caesar":    "https://images.unsplash.com/photo-1540420773420-3366772f4999" + photoQuery,
	"frango-teriyaki":  "https://images.unsplash.com/photo-1521390188846-e2a3a97453a0" + photoQuery,
	"batata-rustica":   "https://images.unsplash.com/photo-1573080496219-bb080dd4f877" + photoQuery,
	"fritas-pequenas":  "https://images.unsplash.com/photo-1573080496219-bb080dd4f877" + photoQuery,
	"fritas-medias":    "https://images.unsplash.com/photo-1573080496219-bb080dd4f877" + photoQuery,
	"hamburguer-kids":  "https://images.unsplash.com/photo-1568901346375-23c9450c58cd" + photoQuery,
	"duplo-da-chapa":   "https://images.unsplash.com/photo-1568901346375-23c9450c58cd" + photoQuery,
	"bacon-bbq":        "https://images.unsplash.com/photo-1594212699903-ec8a3eca50f5" + photoQuery,
	"veggie-da-casa":   "https://images.unsplash.com/photo-1520072959219-c595dc870360" + photoQuery,
	"pizza-margherita": "https://images.unsplash.com/photo-1574071318508-1cdbab80d002" + photoQuery,
	"cola":             "https://images.unsplash.com/photo-1622483767028-3f66f32aef97" + photoQuery,
	"suco-laranja":     "https://images.unsplash.com/photo-1613478223719-2ab802602423" + photoQuery,
}
