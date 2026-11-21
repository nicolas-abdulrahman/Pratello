package menu

import (
	"context"
	"errors"
	"fmt"
	"math"
	"pratello/server/internal/common"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"pratello/server/internal/user"
)

var (
	ErrInvalid      = errors.New("invalid menu input")
	ErrMissing      = errors.New("menu resource not found")
	ErrConflict     = errors.New("menu resource is in use or already exists")
	ErrForbidden    = errors.New("menu access forbidden")
	ErrUnavailable  = errors.New("menu database unavailable")
	ErrNotOrderable = errors.New("item cannot be ordered")
)

type layout struct {
	Slug       string     `json:"restaurant_slug"`
	Name       *string    `json:"name,omitempty"`
	Style      string     `json:"template_style"`
	Accent     *string    `json:"accent_color,omitempty"`
	Secondary  *string    `json:"secondary_color,omitempty"`
	Typography *string    `json:"typography,omitempty"`
	Updated    *time.Time `json:"updated_at,omitempty"`
}

type category struct {
	Image    *string   `json:"image_url,omitempty"`
	ID       int       `json:"id"`
	Slug     string    `json:"restaurant_slug"`
	Name     string    `json:"name"`
	Order    int       `json:"display_order"`
	Created  time.Time `json:"created_at"`
	Fallback bool      `json:"-"`
}

type item struct {
	ID          int       `json:"id"`
	Category    int       `json:"category_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Base        int64     `json:"base_price_cents"`
	Cost        *int64    `json:"base_cost_cents,omitempty"`
	Photo       *string   `json:"photo_url,omitempty"`
	Combo       bool      `json:"is_combo"`
	Available   bool      `json:"is_available"`
	Reviews     bool      `json:"allows_review"`
	Created     time.Time `json:"created_at"`
	Updated     time.Time `json:"updated_at"`
}

type itemView struct {
	item
	Effective int64 `json:"effective_price_cents"`
}

type itemDetail struct {
	itemView
	Ingredients []recipe    `json:"ingredients"`
	Sides       []side      `json:"sides"`
	Components  []component `json:"combo_items"`
}

type ingredient struct {
	ID   int     `json:"id"`
	Name string  `json:"name"`
	Unit *string `json:"default_unit,omitempty"`
}

type recipe struct {
	ID         int      `json:"id"`
	Item       int      `json:"item_id"`
	Ingredient int      `json:"ingredient_id"`
	Name       string   `json:"name"`
	Quantity   *float64 `json:"quantity,omitempty"`
	Unit       *string  `json:"unit,omitempty"`
	Removable  bool     `json:"is_removable"`
}

type side struct {
	ID      int      `json:"id"`
	Item    int      `json:"item_id"`
	Title   string   `json:"title"`
	Min     int      `json:"min_options"`
	Max     int      `json:"max_options"`
	Options []option `json:"options"`
}

type option struct {
	ID        int    `json:"id"`
	Side      int    `json:"side_id"`
	Name      string `json:"name"`
	Price     int64  `json:"additional_price_cents"`
	Available bool   `json:"is_available"`
}

type component struct {
	ID        int      `json:"id"`
	Combo     int      `json:"combo_id"`
	Item      int      `json:"sub_item_id"`
	Name      string   `json:"sub_item_name"`
	Quantity  int      `json:"quantity"`
	Removable []recipe `json:"removable_ingredients"`
}

type sale struct {
	ID     int       `json:"id"`
	Item   int       `json:"item_id"`
	Branch *int      `json:"branch_id,omitempty"`
	Type   string    `json:"discount_type"`
	Value  int64     `json:"discount_value"`
	Start  time.Time `json:"starts_at"`
	End    time.Time `json:"ends_at"`
}

type catalog struct {
	Layout        layout
	Categories    map[int]category
	CategoryOrder []int
	Items         map[int]item
	ItemOrder     []int
	Ingredients   []ingredient
	Recipes       map[int][]recipe
	Sides         map[int][]side
	Components    map[int][]component
	Sales         map[int][]sale
	ItemStock     map[int]bool
	OptionStock   map[int]bool
}

type operation string

const (
	getLayout       operation = "getLayout"
	putLayout       operation = "putLayout"
	listCategories  operation = "listCategories"
	addCategory     operation = "addCategory"
	putCategory     operation = "putCategory"
	dropCategory    operation = "dropCategory"
	listItems       operation = "listItems"
	getItem         operation = "getItem"
	addItem         operation = "addItem"
	putItem         operation = "putItem"
	patchItem       operation = "patchItem"
	archiveItem     operation = "archiveItem"
	stockItem       operation = "stockItem"
	listIngredients operation = "listIngredients"
	addIngredient   operation = "addIngredient"
	putIngredient   operation = "putIngredient"
	dropIngredient  operation = "dropIngredient"
	listRecipes     operation = "listRecipes"
	addRecipe       operation = "addRecipe"
	putRecipe       operation = "putRecipe"
	dropRecipe      operation = "dropRecipe"
	listSides       operation = "listSides"
	addSide         operation = "addSide"
	putSide         operation = "putSide"
	dropSide        operation = "dropSide"
	addOption       operation = "addOption"
	putOption       operation = "putOption"
	dropOption      operation = "dropOption"
	stockOption     operation = "stockOption"
	listComponents  operation = "listComponents"
	addComponent    operation = "addComponent"
	putComponent    operation = "putComponent"
	dropComponent   operation = "dropComponent"
	listSales       operation = "listSales"
	addSale         operation = "addSale"
	dropSale        operation = "dropSale"
)

// nullable distinguishes omission from an explicit null when clearing an item field.
type nullable[T any] struct {
	Set   bool
	Value *T
}

type layoutInput struct {
	Name       *string `json:"name"`
	Style      *string `json:"template_style"`
	Accent     *string `json:"accent_color"`
	Secondary  *string `json:"secondary_color"`
	Typography *string `json:"typography"`
}
type categoryInput struct {
	Image nullable[string] `json:"image_url"`
	Name  string           `json:"name"`
	Order *int32           `json:"display_order"`
}
type itemInput struct {
	Category    *int             `json:"category_id"`
	Name        *string          `json:"name"`
	Description nullable[string] `json:"description"`
	Base        *int64           `json:"base_price_cents"`
	Cost        nullable[int64]  `json:"base_cost_cents"`
	Photo       nullable[string] `json:"photo_url"`
	Combo       *bool            `json:"is_combo"`
	Available   *bool            `json:"is_available"`
	Reviews     *bool            `json:"allows_review"`
}
type ingredientInput struct {
	Name string  `json:"name"`
	Unit *string `json:"default_unit"`
}
type recipeInput struct {
	Ingredient *int     `json:"ingredient_id"`
	Quantity   *float64 `json:"quantity"`
	Unit       *string  `json:"unit"`
	Removable  *bool    `json:"is_removable"`
}
type sideInput struct {
	Title string `json:"title"`
	Min   *int32 `json:"min_options"`
	Max   *int32 `json:"max_options"`
}
type optionInput struct {
	Name      string `json:"name"`
	Price     *int64 `json:"additional_price_cents"`
	Available *bool  `json:"is_available"`
}
type componentInput struct {
	Item     *int   `json:"sub_item_id"`
	Quantity *int32 `json:"quantity"`
}
type stockInput struct {
	Available *bool `json:"is_available"`
}
type saleInput struct {
	Branch *int      `json:"branch_id"`
	Type   string    `json:"discount_type"`
	Value  int64     `json:"discount_value"`
	Start  time.Time `json:"starts_at"`
	End    time.Time `json:"ends_at"`
}

type request struct {
	op                                                  operation
	slug                                                string
	actor, branch, item, id, side, category, page, size int
	write, force                                        bool
	body                                                any
}

type session interface {
	scope(context.Context, int, int) (user.Scope, error)
	publication(context.Context, int) (bool, error)
	load(context.Context, int) (catalog, error)
	mutate(context.Context, request, any) (int, error)
	ingredientUsed(context.Context, int) (bool, error)
	categoryUsed(context.Context, int) (bool, error)
}

type store interface {
	transact(context.Context, string, bool, func(session) (any, error)) (any, error)
}

type Service struct {
	store store
	now   func() time.Time
}

func newService(s store) *Service { return &Service{s, time.Now} }

func (s *Service) execute(ctx context.Context, r request) (any, error) {
	if r.op == addSale {
		if b := r.body.(*saleInput).Branch; b != nil {
			if !validID(*b) {
				return nil, ErrInvalid
			}
			r.branch = *b
		}
	}
	return s.store.transact(ctx, r.slug, r.write, func(tx session) (any, error) {
		a, err := tx.scope(ctx, r.actor, r.branch)
		if err != nil {
			return nil, err
		}
		admin := a.Administrator || a.Organization == "OWNER"
		private := admin || a.Organization == "ANALYST"
		if r.write && !admin {
			if (r.op != stockItem && r.op != stockOption && r.op != addSale && r.op != dropSale) || !a.Manager || r.branch == 0 {
				return nil, ErrForbidden
			}
		}
		if err := authorizePublication(ctx, tx, r, private || a.Manager || a.Employee); err != nil {
			return nil, err
		}
		c, err := tx.load(ctx, r.branch)
		if err != nil {
			return nil, err
		}
		if r.item != 0 && c.Items[r.item].ID == 0 {
			return nil, ErrMissing
		}
		now := s.now()
		if !r.write {
			return c.read(r, private, now), nil
		}
		data, err := prepare(ctx, tx, c, r)
		if err != nil {
			return nil, err
		}
		id, err := tx.mutate(ctx, r, data)
		if err != nil {
			return nil, err
		}
		switch r.op {
		case dropCategory, archiveItem, dropIngredient, dropRecipe, dropSide, dropOption, dropComponent, dropSale, stockItem, stockOption:
			return nil, nil
		}
		c, err = tx.load(ctx, r.branch)
		if err != nil {
			return nil, err
		}
		return c.saved(r, id, now), nil
	})
}

func authorizePublication(ctx context.Context, tx session, r request, preview bool) error {
	if r.branch == 0 && (r.write || preview) {
		return nil
	}
	published, err := tx.publication(ctx, r.branch)
	if err != nil {
		return err
	}
	if !r.write && !preview && !published {
		return ErrMissing
	}
	return nil
}

func validID(v int) bool { return v > 0 && v <= math.MaxInt32 }
func name(value string, maxLength int) (string, bool) {
	v := strings.TrimSpace(value)
	return v, v != "" && utf8.RuneCountInString(v) <= maxLength && !strings.ContainsRune(v, 0)
}
func unit(v *string) bool {
	if v == nil {
		return true
	}
	switch *v {
	case "g", "kg", "ml", "l", "un", "cs", "colher":
		return true
	}
	return false
}
func assign[T any](dst *T, src *T) {
	if src != nil {
		*dst = *src
	}
}
func find[T any](xs []T, match func(T) bool) (value T) {
	for _, x := range xs {
		if match(x) {
			return x
		}
	}
	return value
}

var color = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func prepare(ctx context.Context, tx session, c catalog, r request) (any, error) {
	switch r.op {
	case putLayout:
		v, p := c.Layout, r.body.(*layoutInput)
		for _, entry := range []struct {
			field **string
			value *string
			max   int
		}{{&v.Name, p.Name, 255}, {&v.Typography, p.Typography, 50}} {
			if entry.value != nil {
				text, ok := name(*entry.value, entry.max)
				if !ok {
					return nil, ErrInvalid
				}
				*entry.field = &text
			}
		}
		if p.Style != nil {
			switch *p.Style {
			case "classic", "experience":
				v.Style = *p.Style
			default:
				return nil, ErrInvalid
			}
		}
		for _, col := range []*string{p.Accent, p.Secondary} {
			if col != nil && !color.MatchString(*col) {
				return nil, ErrInvalid
			}
		}
		if p.Accent != nil {
			v.Accent = p.Accent
		}
		if p.Secondary != nil {
			v.Secondary = p.Secondary
		}
		return v, nil
	case addCategory, putCategory:
		v, p := c.Categories[r.id], r.body.(*categoryInput)
		if r.op == putCategory && v.ID == 0 {
			return nil, ErrMissing
		}
		n, ok := name(p.Name, 100)
		if !ok {
			return nil, ErrInvalid
		}
		if v.Fallback && n != v.Name {
			return nil, ErrConflict
		}
		v.Name = n
		if p.Image.Set {
			v.Image = nil
			if p.Image.Value != nil && *p.Image.Value != "" {
				image, err := common.ImageURL(*p.Image.Value)
				if err != nil {
					return nil, ErrInvalid
				}
				v.Image = &image
			}
		}
		if p.Order != nil {
			v.Order = int(*p.Order)
		}
		return v, nil
	case dropCategory:
		v := c.Categories[r.id]
		if v.ID == 0 {
			return nil, ErrMissing
		}
		if v.Fallback {
			return nil, ErrConflict
		}
		used, err := tx.categoryUsed(ctx, r.id)
		if err != nil {
			return nil, err
		}
		if used && !r.force {
			return nil, ErrConflict
		}
		return used, nil
	case addItem, putItem, patchItem:
		return prepareItem(c, r, r.body.(*itemInput))
	case archiveItem:
		return nil, nil
	case stockItem, stockOption:
		if r.op == stockOption {
			v := find(c.Sides[r.item], func(v side) bool { return v.ID == r.side })
			if find(v.Options, func(v option) bool { return v.ID == r.id }).ID == 0 {
				return nil, ErrMissing
			}
		}
		p := r.body.(*stockInput)
		if p.Available == nil {
			return nil, ErrInvalid
		}
		return *p.Available, nil
	case addIngredient, putIngredient, dropIngredient:
		v := find(c.Ingredients, func(v ingredient) bool { return v.ID == r.id })
		if r.op != addIngredient && v.ID == 0 {
			return nil, ErrMissing
		}
		if r.op == dropIngredient {
			used, err := tx.ingredientUsed(ctx, r.id)
			if err != nil {
				return nil, err
			}
			if used {
				return nil, ErrConflict
			}
			return nil, nil
		}
		p := r.body.(*ingredientInput)
		n, ok := name(p.Name, 255)
		if !ok || !unit(p.Unit) {
			return nil, ErrInvalid
		}
		v.Name = n
		if p.Unit != nil {
			v.Unit = p.Unit
		}
		return v, nil
	case addRecipe, putRecipe, dropRecipe:
		v := find(c.Recipes[r.item], func(v recipe) bool { return v.ID == r.id })
		if r.op != addRecipe && v.ID == 0 {
			return nil, ErrMissing
		}
		if r.op == dropRecipe {
			return nil, nil
		}
		p := r.body.(*recipeInput)
		if r.op == addRecipe {
			if p.Ingredient == nil || !validID(*p.Ingredient) {
				return nil, ErrInvalid
			}
			i := find(c.Ingredients, func(v ingredient) bool { return v.ID == *p.Ingredient })
			if i.ID == 0 {
				return nil, ErrMissing
			}
			v.Ingredient, v.Unit = i.ID, i.Unit
		} else if p.Ingredient != nil {
			return nil, ErrInvalid
		}
		if !unit(p.Unit) || (p.Quantity != nil && (*p.Quantity < .001 || *p.Quantity > 9999999.999 || math.IsNaN(*p.Quantity))) {
			return nil, ErrInvalid
		}
		if p.Unit != nil {
			v.Unit = p.Unit
		}
		if p.Quantity != nil {
			v.Quantity = p.Quantity
		}
		assign(&v.Removable, p.Removable)
		return v, nil
	case addSide, putSide, dropSide:
		v := find(c.Sides[r.item], func(v side) bool { return v.ID == r.id })
		if r.op != addSide && v.ID == 0 {
			return nil, ErrMissing
		}
		if r.op == dropSide {
			return nil, nil
		}
		p := r.body.(*sideInput)
		title, ok := name(p.Title, 255)
		if !ok || (r.op == putSide && (p.Min == nil || p.Max == nil)) {
			return nil, ErrInvalid
		}
		v.Title = title
		if r.op == addSide {
			v.Max = 1
		}
		if p.Min != nil {
			v.Min = int(*p.Min)
		}
		if p.Max != nil {
			v.Max = int(*p.Max)
		}
		if v.Min < 0 || v.Max < 1 || v.Min > v.Max {
			return nil, ErrInvalid
		}
		return v, nil
	case addOption, putOption, dropOption:
		g := find(c.Sides[r.item], func(v side) bool { return v.ID == r.side })
		if g.ID == 0 {
			return nil, ErrMissing
		}
		v := find(g.Options, func(v option) bool { return v.ID == r.id })
		if r.op != addOption && v.ID == 0 {
			return nil, ErrMissing
		}
		if r.op == dropOption {
			return nil, nil
		}
		p := r.body.(*optionInput)
		n, ok := name(p.Name, 255)
		if !ok || p.Price == nil || *p.Price < 0 {
			return nil, ErrInvalid
		}
		if r.op == addOption {
			v.Available = true
		}
		v.Name, v.Price = n, *p.Price
		assign(&v.Available, p.Available)
		return v, nil
	case addComponent, putComponent, dropComponent:
		if !c.Items[r.item].Combo {
			return nil, ErrConflict
		}
		v := find(c.Components[r.item], func(v component) bool { return v.ID == r.id })
		if r.op != addComponent && v.ID == 0 {
			return nil, ErrMissing
		}
		if r.op == dropComponent {
			return nil, nil
		}
		p := r.body.(*componentInput)
		if p.Quantity == nil || *p.Quantity < 1 {
			return nil, ErrInvalid
		}
		if r.op == addComponent {
			if p.Item == nil || !validID(*p.Item) || *p.Item == r.item {
				return nil, ErrInvalid
			}
			i := c.Items[*p.Item]
			if i.ID == 0 {
				return nil, ErrMissing
			}
			if i.Combo {
				return nil, ErrConflict
			}
			v.Item = i.ID
		} else if p.Item != nil {
			return nil, ErrInvalid
		}
		v.Quantity = int(*p.Quantity)
		return v, nil
	case addSale:
		p := r.body.(*saleInput)
		if (p.Type != "percent" && p.Type != "fixed_cents") || p.Value < 1 ||
			(p.Type == "percent" && p.Value > 100) || (p.Type == "fixed_cents" && p.Value > c.Items[r.item].Base) ||
			p.Start.IsZero() || !p.End.After(p.Start) {
			return nil, ErrInvalid
		}
		return sale{Branch: p.Branch, Type: p.Type, Value: p.Value, Start: p.Start, End: p.End}, nil
	case dropSale:
		v := find(c.Sales[r.item], func(v sale) bool { return v.ID == r.id })
		if v.ID == 0 {
			return nil, ErrMissing
		}
		if r.branch != 0 && (v.Branch == nil || *v.Branch != r.branch) {
			return nil, ErrMissing
		}
		return nil, nil
	}
	return nil, ErrInvalid
}

func prepareItem(c catalog, r request, p *itemInput) (item, error) {
	v := c.Items[r.item]
	if r.op == addItem {
		v.Available, v.Reviews = true, true
	}
	if r.op != patchItem && (p.Category == nil || p.Name == nil || p.Base == nil) {
		return v, ErrInvalid
	}
	if r.op == putItem && (p.Combo == nil || p.Available == nil) {
		return v, ErrInvalid
	}
	if p.Category != nil {
		if !validID(*p.Category) {
			return v, ErrInvalid
		}
		if c.Categories[*p.Category].ID == 0 {
			return v, ErrMissing
		}
		v.Category = *p.Category
	}
	if p.Name != nil {
		n, ok := name(*p.Name, 255)
		if !ok {
			return v, ErrInvalid
		}
		v.Name = n
	}
	if (p.Base != nil && *p.Base < 0) || (p.Cost.Value != nil && *p.Cost.Value < 0) {
		return v, ErrInvalid
	}
	assign(&v.Base, p.Base)
	assign(&v.Reviews, p.Reviews)
	assign(&v.Available, p.Available)
	if p.Cost.Set {
		v.Cost = p.Cost.Value
	}
	if p.Description.Set || r.op == putItem {
		v.Description = p.Description.Value
	}
	if p.Photo.Set || r.op == putItem {
		v.Photo = nil
		if p.Photo.Value != nil {
			photo, err := common.ImageURL(*p.Photo.Value)
			if err != nil {
				return v, ErrInvalid
			}
			v.Photo = &photo
		}
	}
	if p.Combo != nil && *p.Combo != v.Combo {
		if *p.Combo {
			for _, children := range c.Components {
				if find(children, func(v component) bool { return v.Item == r.item }).ID != 0 {
					return v, ErrConflict
				}
			}
		} else if len(c.Components[r.item]) > 0 {
			return v, ErrConflict
		}
		v.Combo = *p.Combo
	}
	return v, nil
}

func array[T any](xs []T) []T {
	if xs == nil {
		return []T{}
	}
	return xs
}

func (c catalog) view(id, branch int, private bool, at time.Time) itemView {
	v := c.Items[id]
	if !private {
		v.Cost = nil
	}
	v.Available = c.orderable(v)
	return itemView{v, c.effective(v, branch, at)}
}

func (c catalog) sides(id int) []side {
	result := make([]side, 0, len(c.Sides[id]))
	for _, v := range c.Sides[id] {
		v.Options = append(make([]option, 0, len(v.Options)), v.Options...)
		for idx := range v.Options {
			opt := &v.Options[idx]
			if stock, ok := c.OptionStock[opt.ID]; ok {
				opt.Available = stock
			}
		}
		result = append(result, v)
	}
	return result
}

func (c catalog) components(id int) []component {
	result := make([]component, 0, len(c.Components[id]))
	for _, v := range c.Components[id] {
		v.Removable = []recipe{}
		for _, ingredient := range c.Recipes[v.Item] {
			if ingredient.Removable {
				v.Removable = append(v.Removable, ingredient)
			}
		}
		result = append(result, v)
	}
	return result
}

func (c catalog) orderable(v item) bool {
	if v.ID == 0 {
		return false
	}
	available := v.Available
	if override, ok := c.ItemStock[v.ID]; ok {
		available = override
	}
	if !available {
		return false
	}
	for _, g := range c.Sides[v.ID] {
		count := 0
		for _, o := range g.Options {
			available := o.Available
			if override, ok := c.OptionStock[o.ID]; ok {
				available = override
			}
			if available {
				count++
			}
		}
		if count < g.Min {
			return false
		}
	}
	if !v.Combo {
		return true
	}
	children := c.Components[v.ID]
	if len(children) == 0 {
		return false
	}
	for _, child := range children {
		i := c.Items[child.Item]
		if i.Combo || !c.orderable(i) {
			return false
		}
	}
	return true
}

func (c catalog) effective(v item, branch int, at time.Time) int64 {
	var best sale
	for _, p := range c.Sales[v.ID] {
		if at.Before(p.Start) || !at.Before(p.End) || (p.Branch != nil && *p.Branch != branch) {
			continue
		}
		sameScope := (p.Branch == nil) == (best.Branch == nil)
		if best.ID == 0 || (sameScope && p.ID > best.ID) || (p.Branch != nil && best.Branch == nil) {
			best = p
		}
	}
	switch best.Type {
	case "percent":
		return v.Base - (v.Base/100*best.Value + v.Base%100*best.Value/100)
	case "fixed_cents":
		return max(0, v.Base-best.Value)
	}
	return v.Base
}

type itemPage struct {
	Data  []itemView `json:"data"`
	Page  int        `json:"page"`
	Size  int        `json:"page_size"`
	Total int        `json:"total"`
}

func (c catalog) read(r request, private bool, at time.Time) any {
	switch r.op {
	case getLayout:
		return c.Layout
	case listCategories:
		out := make([]category, 0, len(c.CategoryOrder))
		for _, id := range c.CategoryOrder {
			out = append(out, c.Categories[id])
		}
		return out
	case listItems:
		out := itemPage{Data: []itemView{}, Page: r.page, Size: r.size}
		offset := int64(r.page-1) * int64(r.size)
		for _, id := range c.ItemOrder {
			if r.category != 0 && c.Items[id].Category != r.category {
				continue
			}
			if int64(out.Total) >= offset && len(out.Data) < r.size {
				out.Data = append(out.Data, c.view(id, r.branch, private, at))
			}
			out.Total++
		}
		return out
	case getItem:
		return itemDetail{c.view(r.item, r.branch, private, at), array(c.Recipes[r.item]), c.sides(r.item), c.components(r.item)}
	case listIngredients:
		return array(c.Ingredients)
	case listRecipes:
		return array(c.Recipes[r.item])
	case listSides:
		return c.sides(r.item)
	case listComponents:
		return c.components(r.item)
	case listSales:
		out := []sale{}
		for _, v := range c.Sales[r.item] {
			if (v.Branch == nil || *v.Branch == r.branch) && (private || !at.Before(v.Start) && at.Before(v.End)) {
				out = append(out, v)
			}
		}
		return out
	}
	return nil
}

func (c catalog) saved(r request, id int, at time.Time) any {
	switch r.op {
	case putLayout:
		return c.Layout
	case addCategory, putCategory:
		return c.Categories[id]
	case addItem, putItem, patchItem:
		return c.view(id, 0, true, at)
	case addIngredient, putIngredient:
		return find(c.Ingredients, func(v ingredient) bool { return v.ID == id })
	case addRecipe, putRecipe:
		return find(c.Recipes[r.item], func(v recipe) bool { return v.ID == id })
	case addSide, putSide:
		return find(c.sides(r.item), func(v side) bool { return v.ID == id })
	case addOption, putOption:
		g := find(c.sides(r.item), func(v side) bool { return v.ID == r.side })
		return find(g.Options, func(v option) bool { return v.ID == id })
	case addComponent, putComponent:
		return find(c.components(r.item), func(v component) bool { return v.ID == id })
	case addSale:
		return find(c.Sales[r.item], func(v sale) bool { return v.ID == id })
	}
	return nil
}

// Selection is the order domain's input for a configured cart line. A combo
// unit number identifies a specific serving of a component within the bundle.
type Selection struct {
	ItemID              int                `json:"item_id"`
	Quantity            int                `json:"quantity"`
	OptionIDs           []int              `json:"option_ids"`
	RemoveIngredientIDs []int              `json:"remove_ingredient_ids"`
	ComponentRemovals   []ComponentRemoval `json:"component_removals"`
	Notes               string             `json:"notes"`
}

type ComponentRemoval struct {
	ComboItemID         int   `json:"combo_item_id"`
	Unit                int   `json:"unit"` // 1-based.
	RemoveIngredientIDs []int `json:"remove_ingredient_ids"`
}

type CheckoutLine struct {
	LineSnapshot
	BasePrice, SaleSavings int64 // Per unit, before paid extras.
	Combo                  bool
}

type OptionSnapshot struct {
	Group                string `json:"group"`
	Name                 string `json:"name"`
	AdditionalPriceCents int64  `json:"additional_price_cents"`
}

// ComponentSnapshot is a non-billable preparation line. Servings with distinct
// removal instructions are kept separate from the unmodified servings.
type ComponentSnapshot struct {
	ItemID   int
	ItemName string
	Quantity int
	Notes    string
}

// LineSnapshot contains values captured at checkout, independent of subsequent
// catalog edits. Component exclusions are rendered as preparation notes.
type LineSnapshot struct {
	ItemID              int                 `json:"item_id"`
	Quantity            int                 `json:"quantity"`
	ItemName            string              `json:"item_name"`
	UnitPriceCents      int64               `json:"unit_price_cents"`
	LineTotalCents      int64               `json:"line_total_cents"`
	Options             []OptionSnapshot    `json:"options"`
	ExcludedIngredients []string            `json:"excluded_ingredients"`
	Notes               string              `json:"notes"`
	Components          []ComponentSnapshot `json:"-"`
}

func (c catalog) checkoutLines(branch int, selections []Selection, at time.Time) ([]CheckoutLine, error) {
	lines := make([]CheckoutLine, 0, len(selections))
	for _, selection := range selections {
		line, err := c.snapshot(branch, selection, at)
		if err != nil {
			return nil, err
		}
		v := c.Items[selection.ItemID]
		lines = append(lines, CheckoutLine{line, v.Base, v.Base - c.effective(v, branch, at), v.Combo})
	}
	return lines, nil
}

func (s *Service) Snapshot(ctx context.Context, slug string, branch int, selection Selection) (LineSnapshot, error) {
	if !validID(branch) || !validID(selection.ItemID) || !validID(selection.Quantity) {
		return LineSnapshot{}, ErrInvalid
	}
	out, err := s.store.transact(ctx, slug, false, func(tx session) (any, error) {
		ok, err := tx.publication(ctx, branch)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrMissing
		}
		c, err := tx.load(ctx, branch)
		if err != nil {
			return nil, err
		}
		return c.snapshot(branch, selection, s.now())
	})
	if err != nil {
		return LineSnapshot{}, err
	}
	return out.(LineSnapshot), nil
}

func (c catalog) snapshot(branch int, sel Selection, at time.Time) (LineSnapshot, error) {
	if !validID(sel.Quantity) {
		return LineSnapshot{}, ErrInvalid
	}
	v := c.Items[sel.ItemID]
	if !c.orderable(v) {
		return LineSnapshot{}, ErrNotOrderable
	}
	line := LineSnapshot{ItemID: v.ID, Quantity: sel.Quantity, ItemName: v.Name, UnitPriceCents: c.effective(v, branch, at), Options: []OptionSnapshot{}, Notes: sel.Notes}
	selected := make(map[int]bool, len(sel.OptionIDs))
	for _, id := range sel.OptionIDs {
		if selected[id] {
			return LineSnapshot{}, ErrInvalid
		}
		selected[id] = true
	}
	for _, g := range c.sides(v.ID) {
		count := 0
		for _, opt := range g.Options {
			if !selected[opt.ID] {
				continue
			}
			delete(selected, opt.ID)
			count++
			if !opt.Available {
				return LineSnapshot{}, ErrNotOrderable
			}
			if opt.Price > math.MaxInt64-line.UnitPriceCents {
				return LineSnapshot{}, ErrInvalid
			}
			line.UnitPriceCents += opt.Price
			line.Options = append(line.Options, OptionSnapshot{g.Title, opt.Name, opt.Price})
		}
		if count < g.Min || count > g.Max {
			return LineSnapshot{}, ErrInvalid
		}
	}
	if len(selected) != 0 || (line.UnitPriceCents > math.MaxInt64/int64(sel.Quantity)) {
		return LineSnapshot{}, ErrInvalid
	}
	line.LineTotalCents = line.UnitPriceCents * int64(sel.Quantity)
	var err error
	line.ExcludedIngredients, err = exclusions(c.Recipes[v.ID], sel.RemoveIngredientIDs)
	if err != nil {
		return LineSnapshot{}, err
	}
	seen := map[[2]int]bool{}
	customized := map[int][]ComponentSnapshot{}
	for _, choice := range sel.ComponentRemovals {
		part := find(c.Components[v.ID], func(p component) bool { return p.ID == choice.ComboItemID })
		key := [2]int{part.ID, choice.Unit}
		if !v.Combo || part.ID == 0 || choice.Unit < 1 || choice.Unit > part.Quantity || seen[key] {
			return LineSnapshot{}, ErrInvalid
		}
		seen[key] = true
		names, err := exclusions(c.Recipes[part.Item], choice.RemoveIngredientIDs)
		if err != nil {
			return LineSnapshot{}, err
		}
		if len(names) > 0 {
			customized[part.ID] = append(customized[part.ID], ComponentSnapshot{
				ItemID: part.Item, ItemName: part.Name, Quantity: sel.Quantity,
				Notes: "Without: " + strings.Join(names, ", "),
			})
			if line.Notes != "" {
				line.Notes += "\n"
			}
			line.Notes += fmt.Sprintf("%s #%d: without %s", part.Name, choice.Unit, strings.Join(names, ", "))
		}
	}
	if v.Combo {
		for _, part := range c.Components[v.ID] {
			if part.Quantity < 1 || sel.Quantity > math.MaxInt32/part.Quantity {
				return LineSnapshot{}, ErrInvalid
			}
			custom := customized[part.ID]
			line.Components = append(line.Components, custom...)
			if plain := part.Quantity - len(custom); plain > 0 {
				line.Components = append(line.Components, ComponentSnapshot{
					ItemID: part.Item, ItemName: part.Name, Quantity: plain * sel.Quantity,
				})
			}
		}
	}
	return line, nil
}

func exclusions(recipes []recipe, ids []int) ([]string, error) {
	out := make([]string, 0, len(ids))
	seen := map[int]bool{}
	for _, id := range ids {
		r := find(recipes, func(r recipe) bool { return r.ID == id })
		if r.ID == 0 || !r.Removable || seen[id] {
			return nil, ErrInvalid
		}
		seen[id] = true
		out = append(out, r.Name)
	}
	return out, nil
}
