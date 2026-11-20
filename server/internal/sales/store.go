package sales

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
	"pratello/server/internal/menu"
	"pratello/server/internal/user"
)

type postgres struct{ db *sql.DB }

func (u session) loyaltyBalance(ctx context.Context, actor int, cpf string) (balance, error) {
	var b balance
	err := u.tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT w.coins_amount FROM loyalty_wallet w JOIN account a ON a.id=w.account_id WHERE w.restaurant_id=$1 AND (($2<>0 AND a.id=$2) OR ($2=0 AND a.cpf=$3))),0)`, u.org, actor, cpf).Scan(&b.Coins)
	return b, err
}

// FinalizeCheckout uses the order workflow's transaction. The caller has
// already authorized and locked the draft and its restaurant.
func FinalizeCheckout(ctx context.Context, tx *sql.Tx, org int, slug string, branch, order, actor int, cart Cart, at time.Time, freight int64) (Price, error) {
	u := session{tx: tx, org: org, slug: slug}
	all, err := u.coupons(ctx)
	if err != nil {
		return Price{}, err
	}
	byCode, byID := map[string]Coupon{}, map[int]Coupon{}
	for _, c := range all {
		byCode[normalize(c.Code)], byID[c.ID] = c, c
	}
	return checkout(ctx, u, request{op: "finalize", branch: branch, order: order, actor: actor, cart: cart}, byCode, byID, at, freight)
}

type session struct {
	tx   *sql.Tx
	org  int
	slug string
}

func NewService(db *sql.DB) *Service { return &Service{postgres{db}, time.Now} }

func dbError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return errMissing
	}
	var pg *pq.Error
	if errors.As(err, &pg) {
		switch pg.Code {
		case "23505", "23503":
			return errConflict
		case "23514", "22001", "22021", "22P02", "22003":
			return errInvalid
		}
	}
	return err
}
func (p postgres) transact(ctx context.Context, slug string, write bool, fn func(unit) (any, error)) (any, error) {
	if p.db == nil {
		return nil, errUnavailable
	}
	opts := &sql.TxOptions{ReadOnly: !write}
	if !write {
		opts.Isolation = sql.LevelRepeatableRead
	}
	tx, err := p.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	u := session{tx: tx, slug: slug}
	q := `SELECT id FROM restaurant WHERE name_slug=$1 AND archived_at IS NULL`
	if write {
		q += ` FOR UPDATE`
	}
	if err = tx.QueryRowContext(ctx, q, slug).Scan(&u.org); err != nil {
		return nil, dbError(err)
	}
	v, err := fn(u)
	if err != nil {
		return nil, err
	}
	return v, dbError(tx.Commit())
}
func (u session) owner(ctx context.Context, actor int) (bool, error) {
	s, err := user.ResolveScope(ctx, u.tx, actor, u.org, 0)
	return s.Administrator || s.Organization == "OWNER", err
}
func (u session) published(ctx context.Context, branch int) (bool, error) {
	var v bool
	err := u.tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM physical_restaurant WHERE restaurant_id=$1 AND ($2=0 OR id=$2) AND archived_at IS NULL AND is_active AND cnpj_verified)`, u.org, branch).Scan(&v)
	return v, err
}

const fields = `id,code,COALESCE(name,''),COALESCE(title,''),COALESCE(description,''),discount_type,discount_value,min_order_cents,expires_at,max_uses,used_count,is_active,target_scope,target_item_id,is_banner,COALESCE(background_image_url,''),created_at`

func (u session) scan(row interface{ Scan(...any) error }) (Coupon, error) {
	c := Coupon{Slug: u.slug}
	err := row.Scan(&c.ID, &c.Code, &c.Name, &c.Title, &c.Description, &c.Type, &c.Value, &c.Minimum, &c.Expires, &c.Max, &c.Used, &c.Active, &c.Scope, &c.Target, &c.Banner, &c.Image, &c.Created)
	return c, dbError(err)
}
func (u session) coupons(ctx context.Context) ([]Coupon, error) {
	rows, err := u.tx.QueryContext(ctx, `SELECT `+fields+` FROM coupon WHERE restaurant_id=$1 ORDER BY id`, u.org)
	if err != nil {
		return nil, dbError(err)
	}
	defer rows.Close()
	var out []Coupon
	for rows.Next() {
		v, err := u.scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (u session) target(ctx context.Context, id int) (bool, error) {
	var combo bool
	err := u.tx.QueryRowContext(ctx, `SELECT is_combo FROM menu_item WHERE id=$1 AND restaurant_id=$2 AND archived_at IS NULL`, id, u.org).Scan(&combo)
	return combo, dbError(err)
}
func (u session) save(ctx context.Context, c Coupon) (Coupon, error) {
	args := []any{u.org, c.Code, c.Name, c.Title, c.Description, c.Type, c.Value, c.Minimum, c.Expires, c.Max, c.Active, c.Scope, c.Target, c.Banner, c.Image}
	q := `INSERT INTO coupon(restaurant_id,code,name,title,description,discount_type,discount_value,min_order_cents,expires_at,max_uses,is_active,target_scope,target_item_id,is_banner,background_image_url) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`
	if c.ID != 0 {
		args = append(args, c.ID)
		q = `UPDATE coupon SET code=$2,name=$3,title=$4,description=$5,discount_type=$6,discount_value=$7,min_order_cents=$8,expires_at=$9,max_uses=$10,is_active=$11,target_scope=$12,target_item_id=$13,is_banner=$14,background_image_url=$15 WHERE restaurant_id=$1 AND id=$16`
	}
	return u.scan(u.tx.QueryRowContext(ctx, q+` RETURNING `+fields, args...))
}
func (u session) claim(ctx context.Context, actor, id int) error {
	_, err := u.tx.ExecContext(ctx, `INSERT INTO coupon_claim(account_id,coupon_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, actor, id)
	return dbError(err)
}
func (u session) claims(ctx context.Context, actor int) (map[int]bool, error) {
	rows, err := u.tx.QueryContext(ctx, `SELECT cc.coupon_id FROM coupon_claim cc JOIN coupon c ON c.id=cc.coupon_id WHERE cc.account_id=$1 AND c.restaurant_id=$2`, actor, u.org)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[int]bool{}
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		m[id] = true
	}
	return m, rows.Err()
}
func (u session) lines(ctx context.Context, branch int, items []menu.Selection, at time.Time) ([]menu.CheckoutLine, error) {
	return menu.CheckoutLines(ctx, u.tx, u.org, u.slug, branch, items, at)
}
func (u session) wallet(ctx context.Context, actor int, lock bool) (int, int64, error) {
	var id int
	var balance int64
	q := `SELECT id,coins_amount FROM loyalty_wallet WHERE restaurant_id=$1 AND account_id=$2`
	if lock {
		q += ` FOR UPDATE`
	}
	err := u.tx.QueryRowContext(ctx, q, u.org, actor).Scan(&id, &balance)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, nil
	}
	return id, balance, err
}
func (u session) draft(ctx context.Context, order, branch, actor int) (int64, error) {
	var freight int64
	var status string
	err := u.tx.QueryRowContext(ctx, `SELECT o.freight_cents,o.status FROM "order" o JOIN physical_restaurant b ON b.id=o.physical_restaurant_id WHERE o.id=$1 AND b.restaurant_id=$2 AND b.id=$3 AND o.account_id IS NOT DISTINCT FROM NULLIF($4,0) FOR UPDATE OF o`, order, u.org, branch, actor).Scan(&freight, &status)
	if err != nil {
		return 0, dbError(err)
	}
	var locked bool
	err = u.tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM order_promotion WHERE order_id=$1 UNION ALL SELECT 1 FROM transaction WHERE order_id=$1 UNION ALL SELECT 1 FROM order_item WHERE order_id=$1)`, order).Scan(&locked)
	if err != nil {
		return 0, err
	}
	if status != "pending" || locked {
		return 0, errConflict
	}
	return freight, nil
}
func (u session) settle(ctx context.Context, order, wallet int, p Price, freight int64) error {
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if _, err = u.tx.ExecContext(ctx, `INSERT INTO order_promotion(order_id,result) VALUES($1,$2)`, order, string(raw)); err != nil {
		return dbError(err)
	}
	for _, c := range p.Coupons {
		if _, err = u.tx.ExecContext(ctx, `INSERT INTO order_coupon(order_id,coupon_id,coupon_code,discount_cents) VALUES($1,$2,$3,$4)`, order, c.ID, c.Code, c.Savings); err != nil {
			return dbError(err)
		}
		if _, err = u.tx.ExecContext(ctx, `UPDATE coupon SET used_count=used_count+1 WHERE id=$1 AND restaurant_id=$2`, c.ID, u.org); err != nil {
			return dbError(err)
		}
	}
	if p.Coins > 0 {
		var after int64
		if err = u.tx.QueryRowContext(ctx, `UPDATE loyalty_wallet SET coins_amount=coins_amount-$2 WHERE id=$1 AND coins_amount >= $2 RETURNING coins_amount`, wallet, p.Coins).Scan(&after); err != nil {
			return dbError(err)
		}
		if _, err = u.tx.ExecContext(ctx, `INSERT INTO wallet_transaction(loyalty_wallet_id,points_delta,balance_after,reason,order_id) VALUES($1,$2,$3,'checkout_redemption',$4)`, wallet, -p.Coins, after, order); err != nil {
			return dbError(err)
		}
	}
	for _, l := range p.Lines {
		opts, err := json.Marshal(l.Item.Options)
		if err != nil {
			return err
		}
		notes := l.Item.Notes
		if len(l.Item.ExcludedIngredients) > 0 {
			if notes != "" {
				notes += "\n"
			}
			notes += "Without: "
			for i, v := range l.Item.ExcludedIngredients {
				if i > 0 {
					notes += ", "
				}
				notes += v
			}
		}
		var parent int
		if err = u.tx.QueryRowContext(ctx, `INSERT INTO order_item(order_id,item_id,item_name,unit_price_cents,line_total_cents,quantity,notes,selected_options) VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,''),$8) RETURNING id`, order, l.Item.ItemID, l.Item.ItemName, l.Subtotal/int64(l.Item.Quantity), l.Subtotal-l.Discount, l.Item.Quantity, notes, string(opts)).Scan(&parent); err != nil {
			return dbError(err)
		}
		for _, part := range l.Item.Components {
			if _, err := u.tx.ExecContext(ctx, `INSERT INTO order_item(order_id,item_id,item_name,unit_price_cents,line_total_cents,quantity,notes,parent_order_item_id) VALUES($1,$2,$3,0,0,$4,NULLIF($5,''),$6)`, order, part.ItemID, part.ItemName, part.Quantity, part.Notes, parent); err != nil {
				return dbError(err)
			}
		}
	}
	_, err = u.tx.ExecContext(ctx, `UPDATE "order" SET subtotal_cents=$2,discount_cents=$3,total_cents=$4 WHERE id=$1`, order, p.Subtotal, p.Discount+p.Coins, p.Payable+freight)
	return dbError(err)
}
