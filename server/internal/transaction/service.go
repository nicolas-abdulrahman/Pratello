package transaction

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

var (
	errInvalid     = errors.New("invalid payment request")
	errMissing     = errors.New("payment resource not found")
	errForbidden   = errors.New("payment access forbidden")
	errConflict    = errors.New("payment conflicts with order or transaction state")
	errUnavailable = errors.New("payment database unavailable")
)

type Payment struct {
	Reference   string
	Method      string
	AmountCents int64
	Refund      bool
}

// Provider must treat repeat calls with the same Reference as idempotent.
type Provider interface {
	MakePayment(context.Context, Payment) error
}
type MockProvider struct{}

func (MockProvider) MakePayment(context.Context, Payment) error { return nil }

type Transaction struct {
	ID               int       `json:"id"`
	Order            int       `json:"order_id"`
	Amount           int64     `json:"amount_cents"`
	Method           string    `json:"payment_method"`
	Status           string    `json:"status"`
	Refunded         int64     `json:"refunded_cents"`
	Created          time.Time `json:"created_at"`
	Updated          time.Time `json:"updated_at"`
	key, fingerprint string
}
type Pix struct {
	Transaction int        `json:"transaction_id"`
	Code        string     `json:"copy_paste_code"`
	QR          string     `json:"qr_code_base64"`
	Expires     time.Time  `json:"expires_at"`
	Paid        *time.Time `json:"paid_at"`
}
type startInput struct {
	Method string `json:"payment_method"`
	Amount int64  `json:"amount_cents"`
}
type refundInput struct {
	Amount *int64 `json:"amount_cents"`
	Reason string `json:"reason"`
}
type request struct {
	op, key, guest   string
	actor, order, id int
	start            startInput
	refund           refundInput
}
type orderScope struct {
	id, customer   int
	staff, manager bool
	total          int64
	status         string
	submitted      bool
	guestHash      []byte
}
type refund struct {
	amount int64
	reason string
}

type unit interface {
	transactions(context.Context) ([]Transaction, error)
	pix(context.Context, int) (Pix, error)
	reserved(context.Context, time.Time) (int64, error)
	start(context.Context, request, string, string) (Transaction, error)
	savePix(context.Context, Pix) (Pix, error)
	confirm(context.Context, int, time.Time) (Transaction, error)
	refundByKey(context.Context, int, string) (refund, error)
	addRefund(context.Context, Transaction, request, int64) (Transaction, error)
}
type repository interface {
	transact(context.Context, request, bool, func(unit, orderScope) (any, error)) (any, error)
}
type Service struct {
	store    repository
	provider Provider
	now      func() time.Time
}

func (s *Service) execute(ctx context.Context, r request) (any, error) {
	write := r.op == "start" || r.op == "pixCreate" || r.op == "confirm" || r.op == "refund"
	if write && r.op != "confirm" && !validKey(r.key) {
		return nil, errInvalid
	}
	if r.op == "start" && (!validMethod(r.start.Method) || r.start.Amount <= 0) {
		return nil, errInvalid
	}
	if r.op == "refund" && (r.refund.Amount != nil && *r.refund.Amount <= 0 || len(r.refund.Reason) > 1000) {
		return nil, errInvalid
	}
	return s.store.transact(ctx, r, write, func(u unit, o orderScope) (any, error) {
		digest := sha256.Sum256([]byte(r.guest))
		guest := r.actor == 0 && o.customer == 0 && len(r.guest) == 43 && subtle.ConstantTimeCompare(o.guestHash, digest[:]) == 1
		if !guest && (!o.staff && (r.actor <= 0 || o.customer != r.actor)) {
			return nil, errForbidden
		}
		if !o.submitted {
			return nil, errConflict
		}
		if (r.op == "refund" && !o.manager) || (r.op == "start" && r.start.Method == "cash" && !o.staff) {
			return nil, errForbidden
		}
		rows, err := u.transactions(ctx)
		if err != nil {
			return nil, err
		}
		if r.op == "list" {
			return rows, nil
		}
		at := s.now()
		if r.op == "start" {
			fingerprint := fmt.Sprintf("%s:%d", r.start.Method, r.start.Amount)
			for _, t := range rows {
				if t.key == r.key {
					if t.fingerprint != fingerprint {
						return nil, errConflict
					}
					return t, nil
				}
			}
			if o.status == "cancelled" {
				return nil, errConflict
			}
			n, err := u.reserved(ctx, at)
			if err != nil {
				return nil, err
			}
			if o.total < n || r.start.Amount > o.total-n {
				return nil, errConflict
			}
			status := "paid"
			if r.start.Method == "pix" {
				status = "pending"
			}
			t, err := u.start(ctx, r, fingerprint, status)
			if err != nil {
				return nil, err
			}
			if r.start.Method == "pix" {
				_, err = issuePIX(ctx, u, t, at)
			} else {
				err = s.provider.MakePayment(ctx, Payment{fmt.Sprintf("charge:%d:%s", o.id, r.key), t.Method, t.Amount, false})
			}
			return t, err
		}
		var t Transaction
		for _, candidate := range rows {
			if candidate.ID == r.id {
				t = candidate
				break
			}
		}
		if t.ID == 0 {
			return nil, errMissing
		}
		switch r.op {
		case "get":
			return t, nil
		case "pixGet", "pixCreate", "confirm":
			if t.Method != "pix" {
				return nil, errConflict
			}
			p, err := u.pix(ctx, t.ID)
			if errors.Is(err, errMissing) && r.op == "pixCreate" {
				if t.Status != "pending" || o.status == "cancelled" {
					return nil, errConflict
				}
				n, err := u.reserved(ctx, at)
				if err != nil {
					return nil, err
				}
				if o.total < n || t.Amount > o.total-n {
					return nil, errConflict
				}
				return issuePIX(ctx, u, t, at)
			}
			if err != nil {
				return nil, err
			}
			if r.op != "confirm" {
				return p, nil
			}
			if p.Paid != nil {
				return t, nil
			}
			if t.Status != "pending" || o.status == "cancelled" || !at.Before(p.Expires) {
				return nil, errConflict
			}
			if err := s.provider.MakePayment(ctx, Payment{fmt.Sprintf("pix:%d", t.ID), t.Method, t.Amount, false}); err != nil {
				return nil, err
			}
			return u.confirm(ctx, t.ID, at)
		case "refund":
			old, err := u.refundByKey(ctx, t.ID, r.key)
			if err == nil {
				if old.reason != r.refund.Reason || r.refund.Amount != nil && *r.refund.Amount != old.amount {
					return nil, errConflict
				}
				return t, nil
			}
			if !errors.Is(err, errMissing) {
				return nil, err
			}
			if t.Status != "paid" && t.Status != "partially_refunded" {
				return nil, errConflict
			}
			amount := t.Amount - t.Refunded
			if r.refund.Amount != nil {
				amount = *r.refund.Amount
			}
			if amount <= 0 || amount > t.Amount-t.Refunded {
				return nil, errConflict
			}
			if err := s.provider.MakePayment(ctx, Payment{fmt.Sprintf("refund:%d:%s", t.ID, r.key), t.Method, amount, true}); err != nil {
				return nil, err
			}
			return u.addRefund(ctx, t, r, amount)
		}
		return nil, errInvalid
	})
}

func validKey(v string) bool {
	return len(v) > 0 && len(v) <= 128 && strings.IndexFunc(v, func(r rune) bool { return r < 33 || r > 126 }) == -1
}
func validMethod(v string) bool {
	switch v {
	case "credit_card", "debit_card", "pix", "cash":
		return true
	}
	return false
}

func issuePIX(ctx context.Context, u unit, t Transaction, at time.Time) (Pix, error) {
	// This namespace is intentionally not a real bank's PIX payload.
	code := fmt.Sprintf("PRATELLO-MOCK-PIX:%d:%d", t.ID, t.Amount)
	png, err := qrcode.Encode(code, qrcode.Medium, 256)
	if err != nil {
		return Pix{}, err
	}
	return u.savePix(ctx, Pix{Transaction: t.ID, Code: code, QR: base64.StdEncoding.EncodeToString(png), Expires: at.Add(15 * time.Minute)})
}
