// Package integration provides reusable HTTP fixtures for development and tests.
package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
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
	BranchID                                         int
	ItemIDs                                          map[string]int
	Password                                         string // Shared, randomly generated fixture password.
}

type fixtureAPI struct {
	client *http.Client
	base   string
}

func (a fixtureAPI) request(ctx context.Context, method, path, token string, body, result any, want int) error {
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
	var secret [16]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return JoaoBurger{}, err
	}
	f := JoaoBurger{ItemIDs: map[string]int{}, Password: hex.EncodeToString(secret[:])}
	for _, p := range []struct {
		account   *FixtureAccount
		name, key string
	}{
		{&f.Owner, "Geraldo do Rio", "geraldo-do-rio"}, {&f.Administrator, "Mr Puzzles", "mr-puzzles"},
		{&f.Analyst, "Meg from Splatoon", "meg-from-splatoon"}, {&f.Manager, "Petros Barreto", "petros-barreto"}, {&f.Employee, "Bob", "bob"},
	} {
		p.account.Name, p.account.Email, p.account.AvatarURL = p.name, p.key+"@joao-burger.test", "/media/account/"+p.key+".png"
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
	if err := post("/api/restaurants", owner, map[string]any{"name": "João Burger", "slug": JoaoBurgerSlug, "description": "Smash burgers, fritas crocantes e combos para dividir.", "icon_url": restaurantMedia + "brand/icon.png", "cover_url": restaurantMedia + "brand/cover.png"}, nil); err != nil {
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
		{"PUT", root + "/menu", owner, map[string]string{"name": "João Burger — Cardápio", "accent_color": "#F46B36", "secondary_color": "#FFE135"}, 200},
	} {
		if err := a.request(ctx, r.method, r.path, r.token, r.body, nil, r.status); err != nil {
			return f, err
		}
	}
	categories := map[string]int{}
	for n, name := range []string{"Burgers", "Acompanhamentos", "Bebidas", "Sobremesas", "Combos"} {
		var v struct{ ID int }
		if err := post(root+"/categories", owner, map[string]any{"name": name, "display_order": n}, &v); err != nil {
			return f, err
		}
		categories[name] = v.ID
	}
	add := func(v itemSeed, combo bool) error {
		var result struct{ ID int }
		if err := post(root+"/items", owner, map[string]any{"category_id": categories[v.category], "name": v.name, "description": v.description, "base_price_cents": v.price, "is_combo": combo, "photo_url": restaurantMedia + "menu/" + v.key + ".png"}, &result); err != nil {
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
	}
	for _, v := range []struct {
		code, title, description, scope, target, kind string
		value, minimum                                int
	}{
		{"SMASH15", "DIA DO SMASH", "15% de desconto no Smash Clássico.", "item", "smash-classico", "percent", 15, 0},
		{"DUPLA8", "BORA DIVIDIR?", "R$ 8,00 de desconto no Combo Pra Dois.", "combo", "combo-pra-dois", "fixed_cents", 800, 0},
		{"CHAPA10", "PEDIDO GRANDE, DESCONTO JUNTO", "R$ 10,00 de desconto a partir de R$ 80,00.", "order", "", "fixed_cents", 1000, 8000},
	} {
		body := map[string]any{"code": v.code, "name": v.title, "title": v.title, "description": v.description, "target_scope": v.scope, "discount_type": v.kind, "discount_value": v.value, "min_order_cents": v.minimum, "is_banner": true, "background_image_url": restaurantMedia + "sales/" + strings.ToLower(v.code) + ".png"}
		if v.target != "" {
			body["target_item_id"] = f.ItemIDs[v.target]
		}
		if err := post(root+"/coupons", owner, body, nil); err != nil {
			return f, err
		}
	}
	return f, nil
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
	{"smash-classico", "Smash Clássico", "Burgers", "Smash de 90 g, cheddar e molho da casa no brioche.", 2490},
	{"duplo-da-chapa", "Duplo da Chapa", "Burgers", "Dois smashes de 90 g, cheddar duplo e picles.", 3490},
	{"bacon-bbq", "Bacon BBQ", "Burgers", "Smash, cheddar, bacon crocante e barbecue.", 3190},
	{"smash-apimentado", "Smash Apimentado", "Burgers", "Smash, cheddar, jalapeño e maionese picante.", 2990},
	{"frango-crocante", "Frango Crocante", "Burgers", "Frango empanado, alface, picles e maionese de limão.", 2790},
	{"veggie-da-casa", "Veggie da Casa", "Burgers", "Hambúrguer de grão-de-bico, alface, tomate e molho verde.", 2890},
	{"fritas", "Fritas da Chapa", "Acompanhamentos", "Batatas finas com sal da casa, 150 g.", 1200},
	{"fritas-cheddar-bacon", "Fritas Cheddar & Bacon", "Acompanhamentos", "Batatas com creme de cheddar e bacon, 200 g.", 1790},
	{"aneis-de-cebola", "Anéis de Cebola", "Acompanhamentos", "Oito anéis empanados com molho da casa.", 1400},
	{"tiras-de-frango", "Tiras de Frango", "Acompanhamentos", "Quatro tiras crocantes com barbecue.", 1890},
	{"cola", "Cola", "Bebidas", "Lata de 350 ml.", 750},
	{"cha-pessego", "Chá Gelado de Pêssego", "Bebidas", "Chá preto com pêssego, 350 ml.", 850},
	{"limonada", "Limonada da Casa", "Bebidas", "Limão fresco com hortelã, 400 ml.", 990},
	{"brownie", "Brownie da Chapa", "Sobremesas", "Brownie com chocolate meio amargo.", 1190},
	{"milkshake-baunilha", "Milkshake de Baunilha", "Sobremesas", "Batido com sorvete de baunilha, 350 ml.", 1590},
}

var joaoCombos = []comboSeed{
	{"combo-classico", "Combo Clássico", "Smash Clássico, fritas e cola.", 3890, []portion{{"smash-classico", 1}, {"fritas", 1}, {"cola", 1}}},
	{"combo-duplo", "Combo Duplo", "Duplo da Chapa, fritas e cola.", 4790, []portion{{"duplo-da-chapa", 1}, {"fritas", 1}, {"cola", 1}}},
	{"combo-bbq", "Combo BBQ", "Bacon BBQ, anéis de cebola e chá gelado.", 4690, []portion{{"bacon-bbq", 1}, {"aneis-de-cebola", 1}, {"cha-pessego", 1}}},
	{"combo-fogo", "Combo Fogo", "Smash apimentado, fritas cheddar e bacon, e cola.", 4790, []portion{{"smash-apimentado", 1}, {"fritas-cheddar-bacon", 1}, {"cola", 1}}},
	{"combo-crocante", "Combo Crocante", "Frango crocante, tiras de frango e limonada.", 4890, []portion{{"frango-crocante", 1}, {"tiras-de-frango", 1}, {"limonada", 1}}},
	{"combo-veggie", "Combo Veggie", "Veggie da Casa, fritas e limonada.", 4390, []portion{{"veggie-da-casa", 1}, {"fritas", 1}, {"limonada", 1}}},
	{"combo-pra-dois", "Combo Pra Dois", "Dois smashes, fritas e duas colas.", 6590, []portion{{"smash-classico", 2}, {"fritas", 1}, {"cola", 2}}},
	{"combo-dupla-bruta", "Combo Dupla Bruta", "Dois duplos, fritas cheddar e bacon, e duas colas.", 8990, []portion{{"duplo-da-chapa", 2}, {"fritas-cheddar-bacon", 1}, {"cola", 2}}},
	{"combo-smash-brownie", "Combo Smash & Brownie", "Smash Clássico, brownie e chá gelado.", 3990, []portion{{"smash-classico", 1}, {"brownie", 1}, {"cha-pessego", 1}}},
	{"combo-frango-shake", "Combo Frango & Shake", "Frango crocante, fritas e milkshake.", 4790, []portion{{"frango-crocante", 1}, {"fritas", 1}, {"milkshake-baunilha", 1}}},
}
