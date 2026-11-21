package review

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type postgres struct{ db *sql.DB }
type session struct{ tx *sql.Tx }

func NewService(db *sql.DB) *Service { return &Service{postgres{db}} }

func dbError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return errMissing
	}
	var pg *pq.Error
	if errors.As(err, &pg) {
		switch pg.Code {
		case "23505":
			return errConflict
		case "23514", "22021", "22P02":
			return errInvalid
		}
	}
	return err
}

func (p postgres) transact(ctx context.Context, write bool, fn func(unit) (any, error)) (any, error) {
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
	v, err := fn(session{tx})
	if err != nil {
		return nil, dbError(err)
	}
	return v, dbError(tx.Commit())
}

func (u session) publicTarget(ctx context.Context, r request) (int, error) {
	const from = ` FROM restaurant r WHERE r.name_slug=$1 AND r.archived_at IS NULL AND EXISTS(SELECT 1 FROM physical_restaurant b WHERE b.restaurant_id=r.id AND b.cnpj_verified AND b.is_active AND b.archived_at IS NULL)`
	q := `SELECT r.id` + from
	args := []any{r.slug}
	if r.item {
		q = `SELECT id FROM menu_item WHERE id=$2 AND archived_at IS NULL AND restaurant_id=(SELECT r.id` + from + `)`
		args = append(args, r.target)
	}
	var id int
	err := u.tx.QueryRowContext(ctx, q, args...).Scan(&id)
	return id, dbError(err)
}

func (u session) totals(ctx context.Context, item bool, id int) (totals, error) {
	var v totals
	if item {
		var sold int64
		err := u.tx.QueryRowContext(ctx, `SELECT COALESCE(t.review_count,0),COALESCE(t.rating_sum,0),
		(SELECT COALESCE(sum(i.quantity),0) FROM order_item i JOIN "order" o ON o.id=i.order_id
		 WHERE i.item_id=$1 AND i.voided_at IS NULL AND i.parent_order_item_id IS NULL AND o.submitted_at IS NOT NULL AND o.status<>'cancelled')
		FROM (SELECT $1::int AS id) k LEFT JOIN item_review_totals t ON t.target_id=k.id`, id).Scan(&v.Count, &v.Sum, &sold)
		v.Sold = &sold
		return v, err
	}
	err := u.tx.QueryRowContext(ctx, `SELECT COALESCE(t.review_count,0),COALESCE(t.rating_sum,0) FROM (SELECT $1::int AS id) k LEFT JOIN restaurant_review_totals t ON t.target_id=k.id`, id).Scan(&v.Count, &v.Sum)
	return v, err
}

const reviewFields = `id,rating,comment,created_at`

func scan(row interface{ Scan(...any) error }) (Review, error) {
	var v Review
	err := row.Scan(&v.ID, &v.Rating, &v.Comment, &v.Created)
	return v, dbError(err)
}
func (u session) list(ctx context.Context, item bool, id, limit int, offset int64) ([]Review, error) {
	q := `SELECT r.id,r.rating,r.comment,r.created_at,COALESCE(a.name,''),COALESCE(a.avatar_url,'')
		FROM restaurant_review r LEFT JOIN account a ON a.id=r.account_id WHERE r.restaurant_id=$1`
	if item {
		q = `SELECT r.id,r.rating,r.comment,r.created_at,COALESCE(a.name,''),COALESCE(a.avatar_url,'')
			FROM item_review r LEFT JOIN account a ON a.id=r.account_id WHERE r.item_id=$1`
	}
	rows, err := u.tx.QueryContext(ctx, q+` ORDER BY r.created_at DESC,r.id DESC LIMIT $2 OFFSET $3`, id, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Review{}
	for rows.Next() {
		var v Review
		if err := rows.Scan(&v.ID, &v.Rating, &v.Comment, &v.Created, &v.AuthorName, &v.AuthorAvatarURL); err != nil {
			return nil, dbError(err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (u session) purchase(ctx context.Context, r request) (purchase, error) {
	var p purchase
	const columns = `SELECT COALESCE(o.account_id,0),o.guest_token_hash,o.submitted_at IS NOT NULL AND o.status='delivered',`
	q := columns + `b.restaurant_id,true FROM "order" o JOIN physical_restaurant b ON b.id=o.physical_restaurant_id WHERE o.id=$1 FOR SHARE OF o`
	args := []any{r.order}
	if r.item {
		q = columns + `i.item_id,i.voided_at IS NULL AND i.parent_order_item_id IS NULL AND m.allows_review FROM "order" o JOIN order_item i ON i.order_id=o.id JOIN menu_item m ON m.id=i.item_id WHERE o.id=$1 AND i.id=$2 FOR SHARE OF o,i,m`
		args = append(args, r.line)
	}
	err := u.tx.QueryRowContext(ctx, q, args...).Scan(&p.owner, &p.guestHash, &p.delivered, &p.target, &p.eligible)
	return p, dbError(err)
}

func (u session) save(ctx context.Context, r request, target int) (Review, error) {
	if r.item {
		return scan(u.tx.QueryRowContext(ctx, `INSERT INTO item_review(item_id,order_item_id,account_id,rating,comment) VALUES($1,$2,NULLIF($3,0),$4,$5) RETURNING `+reviewFields, target, r.line, r.actor, r.input.Rating, r.input.Comment))
	}
	q := `INSERT INTO restaurant_review(restaurant_id,order_id,account_id,rating,comment) VALUES($1,$2,NULLIF($3,0),$4,$5)`
	if r.actor > 0 {
		q += ` ON CONFLICT(restaurant_id,account_id) DO UPDATE SET rating=EXCLUDED.rating,comment=COALESCE(EXCLUDED.comment,restaurant_review.comment)`
	}
	return scan(u.tx.QueryRowContext(ctx, q+` RETURNING `+reviewFields, target, r.order, r.actor, r.input.Rating, r.input.Comment))
}
