package sales

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"regexp"
	"time"
)

type CompletionCache interface {
	Completed(context.Context, string) bool
	MarkCompleted(context.Context, string)
}

var loyaltyCPF = regexp.MustCompile(`^[0-9]{11}$`)

func ValidLoyaltyCPF(v string) bool { return loyaltyCPF.MatchString(v) }

type balance struct {
	Coins int64 `json:"coins_amount"`
}

func (s *Service) loyaltyBalance(ctx context.Context, slug string, actor int, cpf string) (any, error) {
	if actor == 0 && !ValidLoyaltyCPF(cpf) {
		return nil, errInvalid
	}
	return s.store.transact(ctx, slug, false, func(u unit) (any, error) {
		if actor == 0 {
			ok, err := u.published(ctx, 0)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, errMissing
			}
		}
		return u.loyaltyBalance(ctx, actor, cpf)
	})
}

// AwardPaidOrder credits merchandise spend once while the payment workflow
// holds the order lock. Return the Redis key to publish after SQL commits.
func AwardPaidOrder(ctx context.Context, tx *sql.Tx, order int, cache CompletionCache) (string, error) {
	var org, actor int
	var cpf, status string
	var total, freight, rate, paid int64
	var created time.Time
	err := tx.QueryRowContext(ctx, `SELECT r.id,COALESCE(o.account_id,0),COALESCE(o.customer_cpf,''),o.status,o.total_cents,o.freight_cents,
	(r.loyalty_earn_rate*100)::bigint,o.created_at,
	(SELECT COALESCE(sum(amount_cents-refunded_cents),0) FROM transaction WHERE order_id=o.id AND status IN ('paid','partially_refunded'))
	FROM "order" o JOIN physical_restaurant b ON b.id=o.physical_restaurant_id JOIN restaurant r ON r.id=b.restaurant_id WHERE o.id=$1`, order).Scan(&org, &actor, &cpf, &status, &total, &freight, &rate, &created, &paid)
	if err != nil {
		return "", err
	}
	if (actor == 0 && cpf == "") || status == "cancelled" || total <= freight || paid < total || rate == 0 {
		return "", nil
	}
	// Rate is stored in hundredths of one coin per R$1.00. Round down once.
	hi, lo := bits.Mul64(uint64(total-freight), uint64(rate))
	if hi >= 10000 {
		return "", errInvalid
	}
	points, _ := bits.Div64(hi, lo, 10000)
	if points > math.MaxInt64 {
		return "", errInvalid
	}
	if points == 0 {
		return "", nil
	}
	key := fmt.Sprintf("pratello:loyalty:earn:%d:%d", order, created.UnixMicro())
	if cache != nil && cache.Completed(ctx, key) {
		return "", nil
	}
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM wallet_transaction WHERE order_id=$1 AND reason='paid_order_earn')`, order).Scan(&exists); err != nil || exists {
		return key, err
	}
	if actor == 0 {
		// Match the CPF lock used by auth so registration and wallet creation
		// cannot race or create two different owners for the same CPF.
		if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('account-cpf:' || $1, 0))`, cpf); err != nil {
			return "", err
		}
		err = tx.QueryRowContext(ctx, `SELECT id FROM account WHERE cpf=$1 FOR UPDATE`, cpf).Scan(&actor)
		if errors.Is(err, sql.ErrNoRows) {
			err = tx.QueryRowContext(ctx, `INSERT INTO account(cpf) VALUES($1) RETURNING id`, cpf).Scan(&actor)
		}
		if err != nil {
			return "", err
		}
	}
	var wallet int
	var balance int64
	err = tx.QueryRowContext(ctx, `INSERT INTO loyalty_wallet(account_id,restaurant_id,coins_amount) VALUES($1,$2,$3)
	ON CONFLICT(account_id,restaurant_id) DO UPDATE SET coins_amount=loyalty_wallet.coins_amount+EXCLUDED.coins_amount RETURNING id,coins_amount`, actor, org, int64(points)).Scan(&wallet, &balance)
	if err != nil {
		return "", err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO wallet_transaction(loyalty_wallet_id,points_delta,balance_after,reason,order_id) VALUES($1,$2,$3,'paid_order_earn',$4)`, wallet, int64(points), balance, order)
	return key, err
}
