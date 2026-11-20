package transaction

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"pratello/server/internal/sales"
	"pratello/server/internal/user"
)

type postgres struct {
	db    *sql.DB
	cache sales.CompletionCache
}
type session struct {
	tx         *sql.Tx
	order      int
	cache      sales.CompletionCache
	completion *string
}

func NewService(db *sql.DB, provider Provider) *Service {
	return NewServiceWithLoyaltyCache(db, provider, nil)
}

func NewServiceWithLoyaltyCache(db *sql.DB, provider Provider, cache sales.CompletionCache) *Service {
	return &Service{postgres{db, cache}, provider, time.Now}
}

func dbError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return errMissing
	}
	var pg *pq.Error
	if errors.As(err, &pg) {
		switch pg.Code {
		case "23505":
			return errConflict
		case "23514", "22001", "22003", "22021", "22P02":
			return errInvalid
		}
	}
	return err
}
func (p postgres) transact(ctx context.Context, r request, write bool, fn func(unit, orderScope) (any, error)) (any, error) {
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
	q := `SELECT o.id,COALESCE(o.account_id,0),o.total_cents,o.status,b.restaurant_id,b.id,o.submitted_at IS NOT NULL,o.guest_token_hash FROM "order" o JOIN physical_restaurant b ON b.id=o.physical_restaurant_id WHERE `
	id := r.order
	if r.id != 0 {
		id = r.id
		q += `o.id=(SELECT order_id FROM transaction WHERE id=$1)`
	} else {
		q += `o.id=$1`
	}
	if write {
		q += ` FOR UPDATE OF o`
	}
	var o orderScope
	var org, branch int
	if err = tx.QueryRowContext(ctx, q, id).Scan(&o.id, &o.customer, &o.total, &o.status, &org, &branch, &o.submitted, &o.guestHash); err != nil {
		return nil, dbError(err)
	}
	scope, err := user.ResolveScope(ctx, tx, r.actor, org, branch)
	if err != nil {
		return nil, err
	}
	o.manager = scope.Administrator || scope.Organization == "OWNER" || scope.Manager
	o.staff = o.manager || scope.Employee
	var completion string
	v, err := fn(session{tx, o.id, p.cache, &completion}, o)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, dbError(err)
	}
	if completion != "" && p.cache != nil {
		p.cache.MarkCompleted(ctx, completion)
	}
	return v, nil
}

func (u session) award(ctx context.Context) error {
	key, err := sales.AwardPaidOrder(ctx, u.tx, u.order, u.cache)
	if err == nil {
		*u.completion = key
	}
	return dbError(err)
}

const fields = `id,order_id,amount_cents,transaction_type,status,refunded_cents,created_at,updated_at,COALESCE(idempotency_key,''),COALESCE(request_fingerprint,'')`

func scan(row interface{ Scan(...any) error }) (Transaction, error) {
	var t Transaction
	err := row.Scan(&t.ID, &t.Order, &t.Amount, &t.Method, &t.Status, &t.Refunded, &t.Created, &t.Updated, &t.key, &t.fingerprint)
	return t, dbError(err)
}
func (u session) transactions(ctx context.Context) ([]Transaction, error) {
	rows, err := u.tx.QueryContext(ctx, `SELECT `+fields+` FROM transaction WHERE order_id=$1 ORDER BY id`, u.order)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Transaction{}
	for rows.Next() {
		t, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (u session) reserved(ctx context.Context, at time.Time) (int64, error) {
	var n int64
	err := u.tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(CASE WHEN t.status='pending' THEN t.amount_cents ELSE t.amount_cents-t.refunded_cents END),0) FROM transaction t LEFT JOIN pix_charge p ON p.transaction_id=t.id WHERE t.order_id=$1 AND (t.status IN ('paid','partially_refunded') OR (t.status='pending' AND p.expires_at>$2))`, u.order, at).Scan(&n)
	return n, dbError(err)
}
func (u session) start(ctx context.Context, r request, fingerprint, status string) (Transaction, error) {
	return scan(u.tx.QueryRowContext(ctx, `INSERT INTO transaction(order_id,transaction_type,amount_cents,status,created_by_account_id,idempotency_key,request_fingerprint) VALUES($1,$2,$3,$4,NULLIF($5,0),$6,$7) RETURNING `+fields, u.order, r.start.Method, r.start.Amount, status, r.actor, r.key, fingerprint))
}

const pixFields = `transaction_id,copy_paste_code,qr_code_base64,expires_at,paid_at`

func scanPix(row interface{ Scan(...any) error }) (Pix, error) {
	var p Pix
	err := row.Scan(&p.Transaction, &p.Code, &p.QR, &p.Expires, &p.Paid)
	return p, dbError(err)
}
func (u session) pix(ctx context.Context, id int) (Pix, error) {
	return scanPix(u.tx.QueryRowContext(ctx, `SELECT `+pixFields+` FROM pix_charge WHERE transaction_id=$1`, id))
}
func (u session) savePix(ctx context.Context, p Pix) (Pix, error) {
	return scanPix(u.tx.QueryRowContext(ctx, `INSERT INTO pix_charge(transaction_id,copy_paste_code,qr_code_base64,expires_at) VALUES($1,$2,$3,$4) RETURNING `+pixFields, p.Transaction, p.Code, p.QR, p.Expires))
}
func (u session) confirm(ctx context.Context, id int, at time.Time) (Transaction, error) {
	if _, err := u.tx.ExecContext(ctx, `UPDATE pix_charge SET paid_at=$2 WHERE transaction_id=$1`, id, at); err != nil {
		return Transaction{}, dbError(err)
	}
	return scan(u.tx.QueryRowContext(ctx, `UPDATE transaction SET status='paid' WHERE id=$1 AND order_id=$2 RETURNING `+fields, id, u.order))
}
func (u session) refundByKey(ctx context.Context, id int, key string) (refund, error) {
	var f refund
	err := u.tx.QueryRowContext(ctx, `SELECT amount_cents,COALESCE(reason,'') FROM refund WHERE transaction_id=$1 AND idempotency_key=$2`, id, key).Scan(&f.amount, &f.reason)
	return f, dbError(err)
}
func (u session) addRefund(ctx context.Context, t Transaction, r request, amount int64) (Transaction, error) {
	if _, err := u.tx.ExecContext(ctx, `INSERT INTO refund(transaction_id,amount_cents,reason,created_by_account_id,idempotency_key) VALUES($1,$2,NULLIF($3,''),$4,$5)`, t.ID, amount, r.refund.Reason, r.actor, r.key); err != nil {
		return Transaction{}, dbError(err)
	}
	return scan(u.tx.QueryRowContext(ctx, `UPDATE transaction SET refunded_cents=refunded_cents+$2,status=CASE WHEN refunded_cents+$2=amount_cents THEN 'refunded'::transaction_status_enum ELSE 'partially_refunded'::transaction_status_enum END WHERE id=$1 RETURNING `+fields, t.ID, amount))
}
