package menu

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
	"pratello/server/internal/user"
)

type postgres struct{ db *sql.DB }

// CheckoutLines prices and validates selections within the caller's checkout transaction.
func CheckoutLines(ctx context.Context, tx *sql.Tx, org int, slug string, branch int, selections []Selection, at time.Time) ([]CheckoutLine, error) {
	u := sqlSession{tx: tx, org: org, slug: slug}
	c, err := u.load(ctx, branch)
	if err != nil {
		return nil, err
	}
	return c.checkoutLines(branch, selections, at)
}

type sqlSession struct {
	tx   *sql.Tx
	org  int
	slug string
}

func NewService(db *sql.DB) *Service { return newService(postgres{db}) }

func (p postgres) transact(ctx context.Context, slug string, write bool, f func(session) (any, error)) (any, error) {
	if p.db == nil {
		return nil, ErrUnavailable
	}
	options := &sql.TxOptions{ReadOnly: !write}
	if !write {
		options.Isolation = sql.LevelRepeatableRead
	}
	tx, err := p.db.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	u := sqlSession{tx: tx, slug: slug}
	q := `SELECT id FROM restaurant WHERE name_slug=$1 AND archived_at IS NULL`
	// Share the organization lock with role and branch writers.
	if write {
		q += ` FOR UPDATE`
	}
	if err := tx.QueryRowContext(ctx, q, slug).Scan(&u.org); err != nil {
		return nil, dbError(err)
	}
	v, err := f(u)
	if err != nil {
		return nil, err
	}
	return v, tx.Commit()
}

func dbError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrMissing
	}
	var pg *pq.Error
	if errors.As(err, &pg) {
		switch pg.Code {
		case "23505", "23503":
			return ErrConflict
		case "23514", "22001", "22P02", "22003", "22021":
			return ErrInvalid
		}
	}
	return err
}

func (u sqlSession) scope(ctx context.Context, actor, branch int) (user.Scope, error) {
	return user.ResolveScope(ctx, u.tx, actor, u.org, branch)
}

func (u sqlSession) publication(ctx context.Context, branch int) (bool, error) {
	if branch != 0 {
		return u.exists(ctx, `SELECT cnpj_verified AND is_active FROM physical_restaurant WHERE restaurant_id=$1 AND id=$2 AND archived_at IS NULL`, u.org, branch)
	}
	return u.exists(ctx, `SELECT EXISTS(SELECT 1 FROM physical_restaurant WHERE restaurant_id=$1 AND archived_at IS NULL AND is_active AND cnpj_verified)`, u.org)
}

func (u sqlSession) exists(ctx context.Context, q string, args ...any) (bool, error) {
	var ok bool
	err := u.tx.QueryRowContext(ctx, q, args...).Scan(&ok)
	return ok, dbError(err)
}

func (u sqlSession) ingredientUsed(ctx context.Context, id int) (bool, error) {
	return u.exists(ctx, `SELECT EXISTS(SELECT 1 FROM item_ingredient WHERE ingredient_id=$1)`, id)
}

func (u sqlSession) categoryUsed(ctx context.Context, id int) (bool, error) {
	return u.exists(ctx, `SELECT EXISTS(SELECT 1 FROM menu_item WHERE category_id=$1 AND restaurant_id=$2)`, id, u.org)
}

// Read the normalized rows in one database round trip, within the same snapshot
// as the role and publication checks. Domain availability is evaluated in Go.
const loadSQL = `WITH items AS (
  SELECT id,category_id,name,description,base_price_cents,base_cost_cents,image_url AS photo_url,is_combo,is_available,allows_review,created_at,updated_at
  FROM menu_item WHERE restaurant_id=$1 AND archived_at IS NULL
), groups AS (
  SELECT g.* FROM item_side g JOIN items i ON i.id=g.item_id
)
SELECT jsonb_build_object(
  'Layout',(SELECT row_to_json(m) FROM (
    SELECT $3::text AS restaurant_slug,name,CASE style WHEN 'EXPERIENCIA' THEN 'experience' ELSE 'classic' END AS template_style,
      accent_color,secondary_color,typography,display_config,updated_at FROM menu WHERE restaurant_id=$1
  ) m),
  'Categories',(SELECT jsonb_agg(c ORDER BY c.display_order,c.id) FROM (
    SELECT id,$3::text AS restaurant_slug,name,display_order,created_at,is_fallback,image_url FROM category WHERE restaurant_id=$1 AND archived_at IS NULL
  ) c),
  'Items',(SELECT jsonb_agg(i ORDER BY i.id) FROM items i),
  'Ingredients',(SELECT jsonb_agg(g ORDER BY g.id) FROM (
    SELECT id,name,default_unit FROM ingredient WHERE restaurant_id=$1
  ) g),
  'Recipes',(SELECT jsonb_agg(r ORDER BY r.id) FROM (
    SELECT r.id,r.item_id,r.ingredient_id,g.name,r.quantity,r.measure AS unit,r.is_removable
    FROM item_ingredient r JOIN ingredient g ON g.id=r.ingredient_id JOIN items i ON i.id=r.item_id WHERE g.restaurant_id=$1
  ) r),
  'Sides',(SELECT jsonb_agg(s ORDER BY s.id) FROM (
    SELECT g.*, (SELECT jsonb_agg(o ORDER BY o.id) FROM (
      SELECT id,item_side_id AS side_id,name,additional_price_cents,is_available FROM item_side_option WHERE item_side_id=g.id
    ) o) AS options FROM groups g
  ) s),
  'Components',(SELECT jsonb_agg(c ORDER BY c.id) FROM (
    SELECT c.id,c.combo_id,c.included_item_id AS sub_item_id,m.name AS sub_item_name,c.quantity
    FROM combo_item c JOIN items i ON i.id=c.combo_id JOIN menu_item m ON m.id=c.included_item_id WHERE m.restaurant_id=$1
  ) c),
  'Sales',(SELECT jsonb_agg(s ORDER BY s.id) FROM item_sale s JOIN items i ON i.id=s.item_id
    WHERE s.is_active AND (s.branch_id IS NULL OR s.branch_id=$2)),
  'ItemStock',(SELECT jsonb_object_agg(s.item_id,s.is_available) FROM branch_item_availability s JOIN items i ON i.id=s.item_id WHERE s.branch_id=$2),
  'OptionStock',(SELECT jsonb_object_agg(s.option_id,s.is_available) FROM branch_option_availability s
    JOIN item_side_option o ON o.id=s.option_id JOIN groups g ON g.id=o.item_side_id WHERE s.branch_id=$2)
)`

func (u sqlSession) load(ctx context.Context, branch int) (catalog, error) {
	var raw []byte
	if err := u.tx.QueryRowContext(ctx, loadSQL, u.org, branch, u.slug).Scan(&raw); err != nil {
		return catalog{}, dbError(err)
	}
	var rows struct {
		Layout     *layout
		Categories []struct {
			category
			Fallback bool `json:"is_fallback"`
		}
		Items                  []item
		Ingredients            []ingredient
		Recipes                []recipe
		Sides                  []side
		Components             []component
		Sales                  []sale
		ItemStock, OptionStock map[int]bool
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return catalog{}, err
	}
	c := catalog{
		Layout: layout{Slug: u.slug, Style: "classic"}, Categories: make(map[int]category, len(rows.Categories)), Items: make(map[int]item, len(rows.Items)),
		Ingredients: rows.Ingredients, Recipes: map[int][]recipe{}, Sides: map[int][]side{}, Components: map[int][]component{}, Sales: map[int][]sale{},
		ItemStock: rows.ItemStock, OptionStock: rows.OptionStock,
	}
	if rows.Layout != nil {
		c.Layout = *rows.Layout
	}
	for _, v := range rows.Categories {
		v.category.Fallback = v.Fallback
		c.Categories[v.ID] = v.category
		c.CategoryOrder = append(c.CategoryOrder, v.ID)
	}
	for _, v := range rows.Items {
		c.Items[v.ID] = v
		c.ItemOrder = append(c.ItemOrder, v.ID)
	}
	for _, v := range rows.Recipes {
		c.Recipes[v.Item] = append(c.Recipes[v.Item], v)
	}
	for _, v := range rows.Sides {
		c.Sides[v.Item] = append(c.Sides[v.Item], v)
	}
	for _, v := range rows.Components {
		c.Components[v.Combo] = append(c.Components[v.Combo], v)
	}
	for _, v := range rows.Sales {
		c.Sales[v.Item] = append(c.Sales[v.Item], v)
	}
	return c, nil
}

func (u sqlSession) returningID(ctx context.Context, q string, args ...any) (int, error) {
	var id int
	err := u.tx.QueryRowContext(ctx, q, args...).Scan(&id)
	return id, dbError(err)
}

func (u sqlSession) mutate(ctx context.Context, r request, data any) (int, error) {
	switch r.op {
	case putLayout:
		v := data.(layout)
		var display any
		if v.Display != nil {
			raw, err := json.Marshal(v.Display)
			if err != nil {
				return 0, err
			}
			display = string(raw)
		}
		return u.returningID(ctx, `INSERT INTO menu(restaurant_id,name,style,accent_color,secondary_color,typography,display_config)
		VALUES($1,$2,CASE $3 WHEN 'experience' THEN 'EXPERIENCIA'::menu_style_enum ELSE 'CLASSICO'::menu_style_enum END,$4,$5,$6,$7)
		ON CONFLICT(restaurant_id) DO UPDATE SET name=EXCLUDED.name,style=EXCLUDED.style,accent_color=EXCLUDED.accent_color,
		secondary_color=EXCLUDED.secondary_color,typography=EXCLUDED.typography,display_config=EXCLUDED.display_config RETURNING id`, u.org, v.Name, v.Style, v.Accent, v.Secondary, v.Typography, display)
	case addCategory:
		v := data.(category)
		return u.returningID(ctx, `INSERT INTO category(restaurant_id,name,display_order,image_url) VALUES($1,$2,$3,$4) RETURNING id`, u.org, v.Name, v.Order, v.Image)
	case putCategory:
		v := data.(category)
		return u.returningID(ctx, `UPDATE category SET name=$1,display_order=$2,image_url=$5 WHERE id=$3 AND restaurant_id=$4 RETURNING id`, v.Name, v.Order, r.id, u.org, v.Image)
	case dropCategory:
		if data.(bool) {
			fallback, err := u.returningID(ctx, `INSERT INTO category(restaurant_id,name,display_order,is_fallback) VALUES($1,'Uncategorized',2147483647,TRUE)
			ON CONFLICT(restaurant_id) WHERE is_fallback DO UPDATE SET is_fallback=TRUE RETURNING id`, u.org)
			if err != nil {
				return 0, err
			}
			if _, err := u.tx.ExecContext(ctx, `UPDATE menu_item SET category_id=$1 WHERE category_id=$2 AND restaurant_id=$3`, fallback, r.id, u.org); err != nil {
				return 0, dbError(err)
			}
		}
		return u.returningID(ctx, `UPDATE category SET archived_at=now() WHERE id=$1 AND restaurant_id=$2 RETURNING id`, r.id, u.org)
	case addItem:
		v := data.(item)
		return u.returningID(ctx, `INSERT INTO menu_item(restaurant_id,category_id,name,description,base_price_cents,base_cost_cents,image_url,is_combo,is_available,allows_review)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`, u.org, v.Category, v.Name, v.Description, v.Base, v.Cost, v.Photo, v.Combo, v.Available, v.Reviews)
	case putItem, patchItem:
		v := data.(item)
		return u.returningID(ctx, `UPDATE menu_item SET category_id=$1,name=$2,description=$3,base_price_cents=$4,base_cost_cents=$5,image_url=$6,
		is_combo=$7,is_available=$8,allows_review=$9 WHERE id=$10 AND restaurant_id=$11 RETURNING id`, v.Category, v.Name, v.Description, v.Base, v.Cost, v.Photo, v.Combo, v.Available, v.Reviews, r.item, u.org)
	case archiveItem:
		return u.returningID(ctx, `UPDATE menu_item SET archived_at=now(),is_available=FALSE WHERE id=$1 AND restaurant_id=$2 RETURNING id`, r.item, u.org)
	case stockItem:
		return u.returningID(ctx, `INSERT INTO branch_item_availability(branch_id,item_id,is_available) VALUES($1,$2,$3)
		ON CONFLICT(branch_id,item_id) DO UPDATE SET is_available=EXCLUDED.is_available RETURNING item_id`, r.branch, r.item, data)
	case stockOption:
		return u.returningID(ctx, `INSERT INTO branch_option_availability(branch_id,option_id,is_available) VALUES($1,$2,$3)
		ON CONFLICT(branch_id,option_id) DO UPDATE SET is_available=EXCLUDED.is_available RETURNING option_id`, r.branch, r.id, data)
	case addIngredient:
		v := data.(ingredient)
		return u.returningID(ctx, `INSERT INTO ingredient(restaurant_id,name,default_unit) VALUES($1,$2,$3) RETURNING id`, u.org, v.Name, v.Unit)
	case putIngredient:
		v := data.(ingredient)
		return u.returningID(ctx, `UPDATE ingredient SET name=$1,default_unit=$2 WHERE id=$3 AND restaurant_id=$4 RETURNING id`, v.Name, v.Unit, r.id, u.org)
	case dropIngredient:
		return u.returningID(ctx, `DELETE FROM ingredient WHERE id=$1 AND restaurant_id=$2 RETURNING id`, r.id, u.org)
	case addRecipe:
		v := data.(recipe)
		return u.returningID(ctx, `INSERT INTO item_ingredient(item_id,ingredient_id,quantity,measure,is_removable) VALUES($1,$2,$3,$4,$5) RETURNING id`, r.item, v.Ingredient, v.Quantity, v.Unit, v.Removable)
	case putRecipe:
		v := data.(recipe)
		return u.returningID(ctx, `UPDATE item_ingredient SET quantity=$1,measure=$2,is_removable=$3 WHERE id=$4 AND item_id=$5 RETURNING id`, v.Quantity, v.Unit, v.Removable, r.id, r.item)
	case dropRecipe:
		return u.returningID(ctx, `DELETE FROM item_ingredient WHERE id=$1 AND item_id=$2 RETURNING id`, r.id, r.item)
	case addSide:
		v := data.(side)
		return u.returningID(ctx, `INSERT INTO item_side(item_id,title,min_options,max_options) VALUES($1,$2,$3,$4) RETURNING id`, r.item, v.Title, v.Min, v.Max)
	case putSide:
		v := data.(side)
		return u.returningID(ctx, `UPDATE item_side SET title=$1,min_options=$2,max_options=$3 WHERE id=$4 AND item_id=$5 RETURNING id`, v.Title, v.Min, v.Max, r.id, r.item)
	case dropSide:
		if _, err := u.tx.ExecContext(ctx, `DELETE FROM item_side_option WHERE item_side_id=$1`, r.id); err != nil {
			return 0, dbError(err)
		}
		return u.returningID(ctx, `DELETE FROM item_side WHERE id=$1 AND item_id=$2 RETURNING id`, r.id, r.item)
	case addOption:
		v := data.(option)
		return u.returningID(ctx, `INSERT INTO item_side_option(item_side_id,name,additional_price_cents,is_available) VALUES($1,$2,$3,$4) RETURNING id`, r.side, v.Name, v.Price, v.Available)
	case putOption:
		v := data.(option)
		return u.returningID(ctx, `UPDATE item_side_option SET name=$1,additional_price_cents=$2,is_available=$3 WHERE id=$4 AND item_side_id=$5 RETURNING id`, v.Name, v.Price, v.Available, r.id, r.side)
	case dropOption:
		return u.returningID(ctx, `DELETE FROM item_side_option WHERE id=$1 AND item_side_id=$2 RETURNING id`, r.id, r.side)
	case addComponent:
		v := data.(component)
		return u.returningID(ctx, `INSERT INTO combo_item(combo_id,included_item_id,quantity) VALUES($1,$2,$3) RETURNING id`, r.item, v.Item, v.Quantity)
	case putComponent:
		v := data.(component)
		return u.returningID(ctx, `UPDATE combo_item SET quantity=$1 WHERE id=$2 AND combo_id=$3 RETURNING id`, v.Quantity, r.id, r.item)
	case dropComponent:
		return u.returningID(ctx, `DELETE FROM combo_item WHERE id=$1 AND combo_id=$2 RETURNING id`, r.id, r.item)
	case addSale:
		v := data.(sale)
		return u.returningID(ctx, `INSERT INTO item_sale(item_id,branch_id,discount_type,discount_value,starts_at,ends_at) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, r.item, v.Branch, v.Type, v.Value, v.Start, v.End)
	case dropSale:
		return u.returningID(ctx, `UPDATE item_sale SET is_active=FALSE WHERE id=$1 AND item_id=$2 AND branch_id IS NOT DISTINCT FROM NULLIF($3,0) RETURNING id`, r.id, r.item, r.branch)
	}
	return 0, ErrInvalid
}
