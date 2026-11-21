package integration

import (
	"context"
	"fmt"
)

const tanoshiiSlug = "tanoshii-sushi"

func (a fixtureAPI) seedTanoshii(ctx context.Context, admin, password string) error {
	const root = "/api/restaurants/" + tanoshiiSlug
	post := func(path, token string, body, result any) error {
		return a.request(ctx, "POST", path, token, body, result, 201)
	}
	staff := make([]FixtureAccount, 0, 5)
	for _, person := range []struct{ name, key, avatar string }{
		{"Yabuki Joe", "yabuki-joe", "/media/account/yabuki-joe.svg"},
		{"Nina", "nina", ""},
		{"Mario aquele que virou revolucionario", "mario", "/media/account/mario.svg"},
		{"Greg", "greg", ""}, {"Yamamoto", "yamamoto", ""},
	} {
		email := person.key + "@tanoshii-sushi.test"
		if err := a.request(ctx, "POST", "/api/auth/register/email/request-code", "", map[string]string{"email": email}, nil, 202); err != nil {
			return err
		}
		var tokens struct {
			Access string `json:"access_token"`
		}
		if err := post("/api/auth/register", "", map[string]string{"name": person.name, "email": email, "password": password, "code": "123456"}, &tokens); err != nil {
			return err
		}
		account := FixtureAccount{Name: person.name, Email: email, AccessToken: tokens.Access, AvatarURL: person.avatar}
		if person.avatar != "" {
			if err := a.request(ctx, "PUT", "/api/me/avatar", tokens.Access, map[string]string{"avatar_url": person.avatar}, &account, 200); err != nil {
				return err
			}
		} else if err := a.request(ctx, "GET", "/api/auth/me", tokens.Access, nil, &account, 200); err != nil {
			return err
		}
		staff = append(staff, account)
	}
	owner, analyst, administrator, manager, employee := staff[0], staff[1], staff[2], staff[3], staff[4]
	if err := post("/api/administrators", admin, map[string]int{"account_id": administrator.ID}, nil); err != nil {
		return err
	}
	if err := post("/api/restaurants", owner.AccessToken, map[string]string{"slug": tanoshiiSlug, "name": "Tanoshii Sushi 楽しい", "description": "Sushis, sashimis e temakis preparados na hora."}, nil); err != nil {
		return err
	}
	var branch struct{ ID int }
	if err := post(root+"/branches", owner.AccessToken, map[string]any{"name": "Tanoshii Sushi — Recife", "cnpj": "11444777000161", "owner_cnpj": "11444777000161", "address": map[string]string{"cep": "50030000", "street": "Rua da Aurora", "number": "48", "city": "Recife", "state": "PE"}}, &branch); err != nil {
		return err
	}
	path := fmt.Sprintf("%s/branches/%d", root, branch.ID)
	if err := a.request(ctx, "PUT", path+"/verification", administrator.AccessToken, nil, nil, 200); err != nil {
		return err
	}
	if err := post(root+"/analysts", owner.AccessToken, map[string]int{"account_id": analyst.ID}, nil); err != nil {
		return err
	}
	if err := a.request(ctx, "PUT", fmt.Sprintf("%s/managers/%d", path, manager.ID), owner.AccessToken, nil, nil, 204); err != nil {
		return err
	}
	if err := post(path+"/employees", manager.AccessToken, map[string]any{"account_id": employee.ID, "position": "WAITER"}, nil); err != nil {
		return err
	}
	var category struct{ ID int }
	if err := post(root+"/categories", owner.AccessToken, map[string]string{"name": "Sushis e temakis"}, &category); err != nil {
		return err
	}
	for _, dish := range []struct {
		name, description string
		cents             int
		recipe            []string
	}{
		{"Nigiri de Salmão", "Duas peças de salmão fresco sobre arroz de sushi.", 1890, []string{"Arroz de sushi", "Salmão fresco", "Wasabi"}},
		{"Temaki de Salmão", "Cone de nori com salmão, arroz e cebolinha.", 3290, []string{"Nori", "Arroz de sushi", "Salmão fresco", "Cebolinha"}},
		{"Uramaki Califórnia", "Oito peças com kani, pepino, manga e gergelim.", 2790, []string{"Arroz de sushi", "Kani", "Pepino", "Manga", "Gergelim"}},
	} {
		var item struct{ ID int }
		if err := post(root+"/items", owner.AccessToken, map[string]any{"category_id": category.ID, "name": dish.name, "description": dish.description, "base_price_cents": dish.cents, "photo_url": "/media/default.png"}, &item); err != nil {
			return err
		}
		if err := a.seedSushiRecipe(ctx, root, owner.AccessToken, item.ID, dish.recipe); err != nil {
			return err
		}
	}
	return nil
}

func (a fixtureAPI) seedSushiRecipe(ctx context.Context, root, token string, item int, names []string) error {
	var existing []struct {
		ID   int
		Name string
	}
	if err := a.request(ctx, "GET", root+"/ingredients", token, nil, &existing, 200); err != nil {
		return err
	}
	ids := map[string]int{}
	for _, ing := range existing {
		ids[ing.Name] = ing.ID
	}
	for i, name := range names {
		id := ids[name]
		if id == 0 {
			var ing struct{ ID int }
			if err := a.request(ctx, "POST", root+"/ingredients", token, map[string]string{"name": name}, &ing, 201); err != nil {
				return err
			}
			id = ing.ID
		}
		if err := a.request(ctx, "POST", fmt.Sprintf("%s/items/%d/ingredients", root, item), token, map[string]any{"ingredient_id": id, "is_removable": i == len(names)-1}, nil, 201); err != nil {
			return err
		}
	}
	return nil
}
