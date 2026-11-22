package integration

import (
	"context"
	"encoding/json"
	"fmt"
)

var joaoDisplay = json.RawMessage(`{
  "dark":false,
  "colors":{"bg":"#FFF9DB","primary_color":"#EA580C","secondary_color":"#FDE68A","bg_dark":"#431407","primary_color_dark":"#FACC15","secondary_color_dark":"#9A3412"},
  "menu":{"container":"horizontal"},
  "first_page":[
    {"id":"hero_joao_burger","type":"header1","dark":null,"bg":"bg","button1":{"bg":"primary_color","text":{"dark":true}},"welcome":{"text":"Ola!","color":"primary_color"},"title":{"text":"Joao burger"},"desc":{"text":"temos os melhores amburgueres","color":"text_muted"},"img":"https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&w=1000&q=80"},
    {"id":"sec_promocoes","type":"section","dark":null,"bg":"bg","border":"dotted","border_color":"primary_color","container":"grid","title":{"text":"Promoções","gradient":{"type":"rough","start_char":5,"first_color":"text","second_color":"primary_color"}},"display_items":{"by":"category","value":"sales"}},
    {"id":"promo_card_1","type":"promotion_banner","banner_id":"DUPLA8","dark":null},
    {"id":"sec_combos","type":"section","dark":null,"bg":"bg","border":"solid","container":"horizontal","title":{"text":"Combos"},"display_items":{"by":"category","value":"combos"}}
  ]
}`)

type featuredDish struct {
	name, description, photo string
	original, price          int
	ingredients              []string // Main ingredients come first and cannot be removed.
	fixed                    int
	extras                   []extraSeed
}

type extraSeed struct {
	name  string
	cents int
}

const photoQuery = "?auto=format&fit=crop&q=80&w=300&h=300"

var joaoFeatured = map[string]featuredDish{
	"smash-classico": {"Hambúrguer Duplo Especial", "Dois hambúrgueres, queijo derretido, bacon crocante e muito sabor!", "https://images.unsplash.com/photo-1568901346375-23c9450c58cd" + photoQuery, 3490, 2290,
		[]string{"Pão brioche artesanal", "Hambúrguer 180g (1º)", "Hambúrguer 180g (2º)", "Queijo cheddar inglês", "Bacon defumado", "Alface americana", "Tomate", "Molho secreto da casa"}, 3, []extraSeed{{"Bacon extra", 600}, {"Ovo frito", 350}, {"Cheddar duplo", 400}}},
	"pizza-pepperoni": {"Pizza Pepperoni", "Molho de tomate, muçarela e pepperoni.", "https://images.unsplash.com/photo-1628840042765-356cda07504e" + photoQuery, 4990, 2990,
		[]string{"Massa de fermentação natural (48h)", "Molho de tomate San Marzano", "Muçarela de búfala", "Pepperoni artesanal", "Manjericão fresco"}, 1, []extraSeed{{"Muçarela extra", 500}, {"Pepperoni extra", 650}, {"Azeitonas pretas", 300}}},
	"fritas": {"Batata Frita", "Porção grande de batata crocante.", "https://images.unsplash.com/photo-1573080496219-bb080dd4f877" + photoQuery, 1890, 1290,
		[]string{"Batata inglesa", "Sal marinho", "Alecrim"}, 1, []extraSeed{{"Cheddar cremoso", 450}, {"Bacon em cubos", 500}, {"Azeite trufado", 700}}},
	"frango-crocante": {"Sanduíche de Frango", "Frango grelhado, alface e molho especial da casa.", "https://images.unsplash.com/photo-1521390188846-e2a3a97453a0" + photoQuery, 2590, 1890,
		[]string{"Pão ciabatta", "Peito de frango grelhado", "Queijo suíço", "Alface crocante", "Tomate", "Maionese de limão siciliano"}, 2, []extraSeed{{"Queijo extra", 400}, {"Bacon", 550}}},
	"aneis-de-cebola": {"Onion Rings", "Anéis de cebola empanados.", "https://images.unsplash.com/photo-1639024471283-03518883512d" + photoQuery, 2200, 1590,
		[]string{"Cebola roxa", "Empanado panko"}, 2, []extraSeed{{"Molho ranch", 300}, {"Molho cheddar", 400}}},
	"combo-pra-dois": {"Combo Casal", "", "https://images.unsplash.com/photo-1594212699903-ec8a3eca50f5" + comboPhotoQuery, 7500, 5590, nil, 0, []extraSeed{{"Batata extra", 800}, {"Upgrade p/ refrigerante 2L", 600}, {"Brownie extra", 900}}},
	"brownie": {"Brownie com Sorvete", "Brownie quentinho acompanhado de sorvete de creme.", "https://images.unsplash.com/photo-1606313564200-e75d5e30476c" + photoQuery, 1800, 1250,
		[]string{"Brownie de chocolate belga", "Sorvete de creme", "Calda de chocolate"}, 1, []extraSeed{{"Bola extra de sorvete", 350}, {"Chantilly", 200}, {"Castanhas caramelizadas", 400}}},
}

func (a fixtureAPI) seedJoaoFeatured(ctx context.Context, root, owner string, ids map[string]int) error {
	post := func(path string, body, out any) error { return a.request(ctx, "POST", path, owner, body, out, 201) }
	ingredients := map[string]int{}
	for _, key := range []string{"smash-classico", "pizza-pepperoni", "fritas", "frango-crocante", "aneis-de-cebola", "combo-pra-dois", "brownie"} {
		v := joaoFeatured[key]
		path := fmt.Sprintf("%s/items/%d", root, ids[key])
		if err := post(path+"/sales", map[string]any{"discount_type": "fixed_cents", "discount_value": v.original - v.price, "starts_at": "2020-01-01T00:00:00Z", "ends_at": "2100-01-01T00:00:00Z"}, nil); err != nil {
			return err
		}
		for i, name := range v.ingredients {
			if ingredients[name] == 0 {
				var ing struct{ ID int }
				if err := post(root+"/ingredients", map[string]string{"name": name}, &ing); err != nil {
					return err
				}
				ingredients[name] = ing.ID
			}
			if err := post(path+"/ingredients", map[string]any{"ingredient_id": ingredients[name], "is_removable": i >= v.fixed}, nil); err != nil {
				return err
			}
		}
		var group struct{ ID int }
		if err := post(path+"/sides", map[string]any{"title": "Adicionais", "min_options": 0, "max_options": len(v.extras)}, &group); err != nil {
			return err
		}
		for _, extra := range v.extras {
			if err := post(fmt.Sprintf("%s/sides/%d/options", path, group.ID), map[string]any{"name": extra.name, "additional_price_cents": extra.cents}, nil); err != nil {
				return err
			}
		}
	}
	return nil
}
