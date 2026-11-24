package order

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
	"pratello/server/internal/menu"
	"pratello/server/internal/sales"
	"pratello/server/internal/user"
)

type postgres struct{ db *sql.DB }
type session struct {
	tx *sql.Tx
	scope
}

func NewService(db *sql.DB) *Service { return &Service{postgres{db}, time.Now} }
func equalToken(hash []byte, key string) bool {
	return len(key) == 43 && subtle.ConstantTimeCompare(hash, digest(key)) == 1
}
func dbErr(err error) error {
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
func (p postgres) transact(ctx context.Context, x command, write bool, fn func(unit, scope) (any, error)) (any, error) {
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
	a := scope{}
	branch := x.branch
	if x.op == "scan" {
		err = tx.QueryRowContext(ctx, `SELECT branch_id FROM order_qr WHERE token_hash=$1`, digest(x.code)).Scan(&branch)
	}
	if x.id > 0 {
		err = tx.QueryRowContext(ctx, `SELECT physical_restaurant_id FROM "order" WHERE id=$1`, x.id).Scan(&branch)
	}
	if err != nil {
		return nil, dbErr(err)
	}
	q := `SELECT r.id,r.name_slug,b.id,b.cnpj_verified AND b.is_active AND b.archived_at IS NULL AND r.archived_at IS NULL,b.delivery_fee_cents FROM physical_restaurant b JOIN restaurant r ON r.id=b.restaurant_id WHERE b.id=$1 AND ($2='' OR r.name_slug=$2)`
	if write {
		q += ` FOR UPDATE OF r`
	}
	if err = tx.QueryRowContext(ctx, q, branch, x.slug).Scan(&a.org, &a.slug, &a.branch, &a.published, &a.fee); err != nil {
		return nil, dbErr(err)
	}
	if x.id > 0 {
		q = `SELECT COALESCE(account_id,0),COALESCE(qr_id,0),COALESCE(menu_session_id,0),status,order_type,submitted_at IS NOT NULL,freight_cents,guest_token_hash FROM "order" WHERE id=$1 AND physical_restaurant_id=$2`
		if write {
			q += ` FOR UPDATE`
		}
		if err = tx.QueryRowContext(ctx, q, x.id, a.branch).Scan(&a.owner, &a.qr, &a.session, &a.status, &a.kind, &a.submitted, &a.freight, &a.hash); err != nil {
			return nil, dbErr(err)
		}
	}
	a.role, err = user.ResolveScope(ctx, tx, x.actor, a.org, a.branch)
	if err != nil {
		return nil, err
	}
	v, err := fn(session{tx, a}, a)
	if err != nil {
		return nil, dbErr(err)
	}
	return v, dbErr(tx.Commit())
}
func (u session) lookupQR(ctx context.Context, hash []byte) (qrRecord, error) {
	var v qrRecord
	err := u.tx.QueryRowContext(ctx, `SELECT id,COALESCE(table_id,0) FROM order_qr WHERE branch_id=$1 AND token_hash=$2`, u.branch, hash).Scan(&v.ID, &v.Table)
	return v, dbErr(err)
}
func (u session) table(ctx context.Context, id, number int) (int, int, error) {
	err := u.tx.QueryRowContext(ctx, `SELECT id,number FROM dining_table WHERE physical_restaurant_id=$1 AND is_active AND (($2>0 AND id=$2) OR ($3>0 AND number=$3))`, u.branch, id, number).Scan(&id, &number)
	return id, number, dbErr(err)
}
func (u session) attachCPF(ctx context.Context, id int, cpf string) error {
	var updated int
	err := u.tx.QueryRowContext(ctx, `UPDATE "order" SET customer_cpf=$2 WHERE id=$1 AND account_id IS NULL AND (customer_cpf IS NULL OR customer_cpf=$2) RETURNING id`, id, cpf).Scan(&updated)
	if errors.Is(err, sql.ErrNoRows) {
		return errConflict
	}
	return dbErr(err)
}

func (u session) open(ctx context.Context, x command, kind string, table, number int, hash []byte, qr int64) (int, error) {
	fee := int64(0)
	if kind == "remote" {
		fee = u.fee
	}
	var id int
	err := u.tx.QueryRowContext(ctx, `INSERT INTO "order"(physical_restaurant_id,account_id,order_type,freight_cents,total_cents,submitted_at,guest_token_hash,qr_id,customer_cpf) VALUES($1,NULLIF($2,0),$3,$4,$4,NULL,$5,NULLIF($6,0),NULLIF($7,'')) RETURNING id`, u.branch, x.actor, kind, fee, hash, qr, x.create.CPF).Scan(&id)
	if err != nil {
		return 0, dbErr(err)
	}
	if kind == "local" {
		_, err = u.tx.ExecContext(ctx, `INSERT INTO local_order(id,table_id,table_number) VALUES($1,NULLIF($2,0),NULLIF($3,0))`, id, table, number)
	} else {
		a := x.create.Address
		_, err = u.tx.ExecContext(ctx, `INSERT INTO remote_order(id,contact_phone,delivery_cep,delivery_street,delivery_number,delivery_apt,delivery_complement,delivery_city,delivery_state) VALUES($1,$2,$3,$4,$5,NULLIF($6,''),NULLIF($7,''),$8,$9)`, id, x.create.Phone, a.CEP, a.Street, a.Number, a.Apt, a.Complement, a.City, a.State)
	}
	return id, dbErr(err)
}

const summary = `o.id,o.physical_restaurant_id AS branch_id,o.account_id AS user_id,o.order_type,o.status,o.subtotal_cents,o.discount_cents,o.freight_cents,o.total_cents,o.created_at,o.updated_at,o.submitted_at,o.qr_id,
	o.display_code`
const itemFields = `i.id,i.order_id,i.item_id,i.item_name,i.unit_price_cents,i.line_total_cents,i.quantity,i.notes,i.selected_options,i.parent_order_item_id,i.voided_at,i.voided_reason`
const draftFields = `d.snapshot - 'options' || jsonb_build_object('id',d.id,'order_id',d.order_id,'selected_options',d.snapshot->'options','voided_at',d.voided_at,'voided_reason',d.voided_reason)`

func (u session) raw(ctx context.Context, q string, args ...any) (json.RawMessage, error) {
	var r json.RawMessage
	err := u.tx.QueryRowContext(ctx, q, args...).Scan(&r)
	return r, dbErr(err)
}
func (u session) view(ctx context.Context, id int) (json.RawMessage, error) {
	return u.raw(ctx, `SELECT to_jsonb(s) FROM (SELECT `+summary+`,
	(SELECT jsonb_build_object('table_id',table_id,'table_number',table_number) FROM local_order WHERE id=o.id) AS local_info,
	(SELECT jsonb_build_object('phone',contact_phone,'delivery_address',jsonb_strip_nulls(jsonb_build_object('cep',delivery_cep,'street',delivery_street,'number',delivery_number,'apt',delivery_apt,'complement',delivery_complement,'city',delivery_city,'state',delivery_state))) FROM remote_order WHERE id=o.id) AS remote_info,
	CASE WHEN o.submitted_at IS NULL THEN COALESCE((SELECT jsonb_agg(`+draftFields+` ORDER BY d.id) FROM order_draft_item d WHERE d.order_id=o.id),'[]')
	ELSE COALESCE((SELECT jsonb_agg(v ORDER BY v.id) FROM (SELECT `+itemFields+` FROM order_item i WHERE i.order_id=o.id) v),'[]') END AS items,
	COALESCE((SELECT jsonb_agg(v ORDER BY v.id) FROM (SELECT t.id,t.order_id,t.amount_cents,t.transaction_type AS payment_method,t.status,t.refunded_cents,t.created_at,t.updated_at FROM transaction t WHERE t.order_id=o.id) v),'[]') AS transactions
	FROM "order" o WHERE o.id=$1) s`, id)
}
func (u session) list(ctx context.Context, x command) (json.RawMessage, error) {
	return u.raw(ctx, `WITH found AS (SELECT `+summary+`, (SELECT jsonb_build_object('table_id',table_id,'table_number',table_number) FROM local_order WHERE id=o.id) AS local_info FROM "order" o WHERE o.physical_restaurant_id=$1 AND submitted_at IS NOT NULL AND ($2='' OR status::text=$2)), pg AS (SELECT * FROM found ORDER BY id DESC LIMIT $3 OFFSET $4) SELECT jsonb_build_object('data',COALESCE((SELECT jsonb_agg(pg ORDER BY id DESC) FROM pg),'[]'),'page',$5::int,'page_size',$3::int,'total',(SELECT count(*) FROM found))`, u.branch, x.status.Status, x.size, int64(x.page-1)*int64(x.size), x.page)
}
func (u session) lines(ctx context.Context, id int) ([]draftItem, error) {
	rows, err := u.tx.QueryContext(ctx, `SELECT id,item_id,selection,snapshot FROM order_draft_item WHERE order_id=$1 AND voided_at IS NULL ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []draftItem
	for rows.Next() {
		var d draftItem
		var s, v []byte
		if err := rows.Scan(&d.ID, &d.ItemID, &s, &v); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(s, &d.Selection); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(v, &d.Snapshot); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
func (u session) price(ctx context.Context, sel menu.Selection, at time.Time) (menu.LineSnapshot, error) {
	xs, err := menu.CheckoutLines(ctx, u.tx, u.org, u.slug, u.branch, []menu.Selection{sel}, at)
	if err != nil {
		return menu.LineSnapshot{}, err
	}
	return xs[0].LineSnapshot, nil
}
func (u session) add(ctx context.Context, id int, sel menu.Selection, line menu.LineSnapshot) (json.RawMessage, error) {
	s, err := json.Marshal(sel)
	if err != nil {
		return nil, err
	}
	v, err := json.Marshal(line)
	if err != nil {
		return nil, err
	}
	r, err := u.raw(ctx, `WITH d AS (INSERT INTO order_draft_item(order_id,item_id,selection,snapshot) VALUES($1,$2,$3,$4) RETURNING *) SELECT `+draftFields+` FROM d`, id, sel.ItemID, string(s), string(v))
	if err != nil {
		return nil, err
	}
	return r, u.updateDraftTotals(ctx, id)
}
func (u session) void(ctx context.Context, id, item int, reason string) error {
	var changed int
	err := u.tx.QueryRowContext(ctx, `UPDATE order_draft_item SET voided_at=now(),voided_reason=NULLIF($3,'') WHERE order_id=$1 AND id=$2 AND voided_at IS NULL RETURNING id`, id, item, reason).Scan(&changed)
	if err != nil {
		return dbErr(err)
	}
	return u.updateDraftTotals(ctx, id)
}
func (u session) updateDraftTotals(ctx context.Context, id int) error {
	_, err := u.tx.ExecContext(ctx, `UPDATE "order" o SET subtotal_cents=v.total,total_cents=v.total+o.freight_cents FROM (SELECT COALESCE(sum((snapshot->>'line_total_cents')::bigint),0) AS total FROM order_draft_item WHERE order_id=$1 AND voided_at IS NULL) v WHERE o.id=$1`, id)
	return dbErr(err)
}
func (u session) finalize(ctx context.Context, id, actor int, c sales.Cart, at time.Time, freight int64) error {
	if _, err := sales.FinalizeCheckout(ctx, u.tx, u.org, u.slug, u.branch, id, actor, c, at, freight); err != nil {
		return sales.MapCheckoutError(err, errInvalid, errMissing, errForbidden, errConflict)
	}
	var updated int
	err := u.tx.QueryRowContext(ctx, `UPDATE "order" o SET submitted_at=$2,
		display_code=CASE WHEN o.order_type='local' THEN
			(SELECT lpad(n::text,3,'0') FROM generate_series(0,999) n
			 WHERE NOT EXISTS (SELECT 1 FROM "order" existing
				WHERE existing.physical_restaurant_id=o.physical_restaurant_id
				AND existing.order_type='local' AND existing.status IN ('pending','preparing','ready')
				AND existing.display_code=lpad(n::text,3,'0'))
			 ORDER BY random() LIMIT 1)
			ELSE NULL END
		WHERE o.id=$1 AND (o.order_type<>'local' OR EXISTS
			(SELECT 1 FROM generate_series(0,999) n WHERE NOT EXISTS
				(SELECT 1 FROM "order" existing WHERE existing.physical_restaurant_id=o.physical_restaurant_id
				AND existing.order_type='local' AND existing.status IN ('pending','preparing','ready')
				AND existing.display_code=lpad(n::text,3,'0'))))
		RETURNING o.id`, id, at).Scan(&updated)
	if errors.Is(err, sql.ErrNoRows) {
		return errConflict
	}
	return dbErr(err)
}
func (u session) status(ctx context.Context, id int, status string) error {
	_, err := u.tx.ExecContext(ctx, `UPDATE "order" SET status=$2 WHERE id=$1`, id, status)
	return dbErr(err)
}
func (u session) newQR(ctx context.Context, table int, hash []byte) (int64, error) {
	var id int64
	err := u.tx.QueryRowContext(ctx, `INSERT INTO order_qr(branch_id,table_id,token_hash) VALUES($1,NULLIF($2,0),$3) RETURNING id`, u.branch, table, hash).Scan(&id)
	return id, dbErr(err)
}
func (u session) setFee(ctx context.Context, fee int64) error {
	_, err := u.tx.ExecContext(ctx, `UPDATE physical_restaurant SET delivery_fee_cents=$2 WHERE id=$1`, u.branch, fee)
	return dbErr(err)
}
func (u session) trackOpen(ctx context.Context, id, actor, table int, qr int64, channel, device string) error {
	_, err := u.tx.ExecContext(ctx, `WITH s AS (INSERT INTO menu_session(restaurant_id,physical_restaurant_id,dining_table_id,account_id,channel,device_type) VALUES($1,$2,NULLIF($3,0),NULLIF($4,0),$5,$6) RETURNING id),
	updated AS (UPDATE "order" SET menu_session_id=s.id FROM s WHERE "order".id=$7 RETURNING s.id AS session_id)
	INSERT INTO ux_event(session_id,restaurant_id,physical_restaurant_id,event_type,metadata) SELECT session_id,$1,$2,'menu_open',jsonb_strip_nulls(jsonb_build_object('order_id',$7::int,'qr_id',NULLIF($8::bigint,0))) FROM updated`, u.org, u.branch, table, actor, channel, device, id, qr)
	return dbErr(err)
}
func (u session) track(ctx context.Context, id int, event string, item, quantity int) error {
	_, err := u.tx.ExecContext(ctx, `WITH target AS (SELECT id,menu_session_id,total_cents,qr_id FROM "order" WHERE id=$1 AND menu_session_id IS NOT NULL),
	touched AS (UPDATE menu_session s SET last_activity_at=now(),converted_order_id=CASE WHEN $2='order_submitted' THEN target.id ELSE s.converted_order_id END FROM target WHERE s.id=target.menu_session_id RETURNING s.id)
	INSERT INTO ux_event(session_id,restaurant_id,physical_restaurant_id,event_type,item_id,quantity,cart_total_cents,metadata)
	SELECT t.menu_session_id,$3,$4,$2::text::ux_event_type_enum,CASE WHEN $2='cart_item_remove' THEN (SELECT item_id FROM order_draft_item WHERE id=$5 AND order_id=t.id) ELSE NULLIF($5,0) END,
	CASE WHEN $2='cart_item_remove' THEN (SELECT (selection->>'quantity')::int FROM order_draft_item WHERE id=$5 AND order_id=t.id) ELSE NULLIF($6,0) END,
	t.total_cents,jsonb_strip_nulls(jsonb_build_object('order_id',t.id,'qr_id',t.qr_id)) FROM target t JOIN touched ON touched.id=t.menu_session_id`, id, event, u.org, u.branch, item, quantity)
	return dbErr(err)
}
