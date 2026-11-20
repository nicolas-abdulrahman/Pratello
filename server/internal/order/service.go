package order

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	"regexp"
	"strings"
	"time"

	"pratello/server/internal/menu"
	"pratello/server/internal/sales"
	"pratello/server/internal/user"
)

var (
	errInvalid     = errors.New("invalid order input")
	errMissing     = errors.New("order, branch, table or QR not found")
	errForbidden   = errors.New("order access forbidden")
	errConflict    = errors.New("order cannot be changed in its current state")
	errUnavailable = errors.New("order database unavailable")
)

type address struct {
	CEP        string `json:"cep"`
	Street     string `json:"street"`
	Number     string `json:"number"`
	Apt        string `json:"apt"`
	Complement string `json:"complement"`
	City       string `json:"city"`
	State      string `json:"state"`
}

type createInput struct {
	TableID     int      `json:"table_id"`
	TableNumber int      `json:"table_number"`
	Phone       string   `json:"phone"`
	Address     *address `json:"delivery_address"`
	submitInput
	Items []menu.Selection `json:"items"`
}
type submitInput struct {
	CPF   string   `json:"customer_cpf"`
	Codes []string `json:"coupon_codes"`
	Saved []int    `json:"saved_coupon_ids"`
	Coins int64    `json:"coins"`
}
type statusInput struct {
	Status string `json:"status"`
}
type qrInput struct {
	TableID int `json:"table_id"`
}
type feeInput struct {
	Fee *int64 `json:"delivery_fee_cents"`
}

type command struct {
	op, slug, guest, code, reason       string
	actor, branch, id, item, page, size int
	create                              createInput
	submit                              submitInput
	selection                           menu.Selection
	status                              statusInput
	qr                                  qrInput
	fee                                 feeInput
	device                              string
}
type scope struct {
	org, branch, owner   int
	qr, session          int64
	slug, status, kind   string
	published, submitted bool
	fee, freight         int64
	hash                 []byte
	role                 user.Scope
}
type draftItem struct {
	ID        int
	ItemID    int
	Selection menu.Selection
	Snapshot  menu.LineSnapshot
	Voided    bool
}
type qrLink struct {
	ID    int64  `json:"id"`
	Table int    `json:"table_id,omitempty"`
	Token string `json:"-"`
}
type qrRecord struct {
	ID    int64
	Table int
}
type opened struct {
	Order json.RawMessage
	Guest string
}

type unit interface {
	attachCPF(context.Context, int, string) error
	lookupQR(context.Context, []byte) (qrRecord, error)
	table(context.Context, int, int) (int, int, error)
	open(context.Context, command, string, int, int, []byte, int64) (int, error)
	view(context.Context, int) (json.RawMessage, error)
	list(context.Context, command) (json.RawMessage, error)
	lines(context.Context, int) ([]draftItem, error)
	price(context.Context, menu.Selection, time.Time) (menu.LineSnapshot, error)
	add(context.Context, int, menu.Selection, menu.LineSnapshot) (json.RawMessage, error)
	void(context.Context, int, int, string) error
	finalize(context.Context, int, int, sales.Cart, time.Time, int64) error
	status(context.Context, int, string) error
	newQR(context.Context, int, []byte) (int64, error)
	setFee(context.Context, int64) error
	trackOpen(context.Context, int, int, int, int64, string, string) error
	track(context.Context, int, string, int, int) error
}
type repository interface {
	transact(context.Context, command, bool, func(unit, scope) (any, error)) (any, error)
}
type Service struct {
	store repository
	now   func() time.Time
}

func validID(id int) bool { return id > 0 && id <= math.MaxInt32 }
func token() (string, []byte, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", nil, err
	}
	s := base64.RawURLEncoding.EncodeToString(b[:])
	return s, digest(s), nil
}
func digest(s string) []byte         { h := sha256.Sum256([]byte(s)); return h[:] }
func canManage(a user.Scope) bool    { return a.Administrator || a.Organization == "OWNER" || a.Manager }
func canConfigure(a user.Scope) bool { return a.Administrator || a.Organization == "OWNER" }
func transition(from, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case "pending":
		return to == "preparing" || to == "cancelled"
	case "preparing":
		return to == "ready" || to == "cancelled"
	case "ready":
		return to == "delivered" || to == "cancelled"
	}
	return false
}

var postal = regexp.MustCompile(`^[0-9]{8}$`)
var state = regexp.MustCompile(`^[A-Z]{2}$`)
var phone = regexp.MustCompile(`^\+?[0-9 ()-]{8,22}$`)

func (s *Service) execute(ctx context.Context, x command) (any, error) {
	write := x.op != "get" && x.op != "list"
	return s.store.transact(ctx, x, write, func(u unit, a scope) (any, error) {
		staff := canManage(a.role) || a.role.Employee
		at := s.now()
		switch x.op {
		case "fee":
			if !canConfigure(a.role) {
				return nil, errForbidden
			}
			if x.fee.Fee == nil || *x.fee.Fee < 0 {
				return nil, errInvalid
			}
			return x.fee, u.setFee(ctx, *x.fee.Fee)
		case "qr":
			if !canManage(a.role) {
				return nil, errForbidden
			}
			if x.qr.TableID != 0 && !validID(x.qr.TableID) {
				return nil, errInvalid
			}
			if x.qr.TableID != 0 {
				if _, _, err := u.table(ctx, x.qr.TableID, 0); err != nil {
					return nil, err
				}
			}
			key, hash, err := token()
			if err != nil {
				return nil, err
			}
			id, err := u.newQR(ctx, x.qr.TableID, hash)
			return qrLink{id, x.qr.TableID, key}, err
		case "list":
			if !staff {
				return nil, errForbidden
			}
			return u.list(ctx, x)
		case "local", "remote", "scan":
			if !a.published {
				return nil, errMissing
			}
			return s.open(ctx, u, a, x, at)
		}
		owner := x.actor > 0 && x.actor == a.owner
		guest := x.actor == 0 && a.owner == 0 && len(a.hash) > 0 && equalToken(a.hash, x.guest)
		if !staff && !owner && !guest {
			return nil, errForbidden
		}
		if x.op == "get" {
			return u.view(ctx, x.id)
		}
		if x.op == "status" {
			if !staff {
				return nil, errForbidden
			}
			if !a.submitted || !transition(a.status, x.status.Status) {
				return nil, errConflict
			}
			if x.status.Status != a.status {
				if err := u.status(ctx, x.id, x.status.Status); err != nil {
					return nil, err
				}
			}
			return u.view(ctx, x.id)
		}
		if a.submitted || a.status != "pending" {
			return nil, errConflict
		}
		if !a.published {
			return nil, errMissing
		}
		switch x.op {
		case "add":
			return add(ctx, u, x.id, x.selection, at)
		case "void":
			if len(x.reason) > 500 {
				return nil, errInvalid
			}
			if err := u.void(ctx, x.id, x.item, x.reason); err != nil {
				return nil, err
			}
			return nil, u.track(ctx, x.id, "cart_item_remove", x.item, 0)
		case "submit":
			if x.submit.CPF != "" {
				if a.owner != 0 || !sales.ValidLoyaltyCPF(x.submit.CPF) {
					return nil, errInvalid
				}
				if err := u.attachCPF(ctx, x.id, x.submit.CPF); err != nil {
					return nil, err
				}
			}
			if err := submit(ctx, u, x.id, a.owner, x.submit, at, a.freight); err != nil {
				return nil, err
			}
			return u.view(ctx, x.id)
		}
		return nil, errInvalid
	})
}

func (s *Service) open(ctx context.Context, u unit, a scope, x command, at time.Time) (opened, error) {
	if x.create.CPF != "" && (x.actor > 0 || !sales.ValidLoyaltyCPF(x.create.CPF)) {
		return opened{}, errInvalid
	}
	kind, channel := x.op, "counter"
	tableID, number := 0, 0
	var qr int64
	if x.op == "scan" {
		record, err := u.lookupQR(ctx, digest(x.code))
		if err != nil {
			return opened{}, err
		}
		qr = record.ID
		kind, channel = "local", "qr_table"
		x.create.TableID = record.Table
	}
	if kind == "local" {
		p := x.create
		if p.Address != nil || p.Phone != "" || p.TableID != 0 && p.TableNumber != 0 || p.TableID < 0 || p.TableNumber < 0 || p.TableID > math.MaxInt32 || p.TableNumber > math.MaxInt32 {
			return opened{}, errInvalid
		}
		if p.TableID != 0 || p.TableNumber != 0 {
			var err error
			tableID, number, err = u.table(ctx, p.TableID, p.TableNumber)
			if err != nil {
				return opened{}, err
			}
		}
	} else {
		channel = "remote_link"
		p := x.create
		if p.TableID != 0 || p.TableNumber != 0 || !validDelivery(p.Phone, p.Address) {
			return opened{}, errInvalid
		}
	}
	if len(x.create.Items) > 100 || len(x.create.Items) == 0 && (len(x.create.Codes) > 0 || len(x.create.Saved) > 0 || x.create.Coins != 0) {
		return opened{}, errInvalid
	}
	var hash []byte
	var guest string
	if x.actor == 0 {
		var err error
		guest, hash, err = token()
		if err != nil {
			return opened{}, err
		}
	}
	id, err := u.open(ctx, x, kind, tableID, number, hash, qr)
	if err != nil {
		return opened{}, err
	}
	if err := u.trackOpen(ctx, id, x.actor, tableID, qr, channel, x.device); err != nil {
		return opened{}, err
	}
	for _, line := range x.create.Items {
		if _, err := add(ctx, u, id, line, at); err != nil {
			return opened{}, err
		}
	}
	if len(x.create.Items) > 0 {
		fee := int64(0)
		if kind == "remote" {
			fee = a.fee
		}
		if err := submit(ctx, u, id, x.actor, x.create.submitInput, at, fee); err != nil {
			return opened{}, err
		}
	}
	v, err := u.view(ctx, id)
	return opened{v, guest}, err
}

func validDelivery(p string, a *address) bool {
	if a == nil || !phone.MatchString(p) || !postal.MatchString(a.CEP) || !state.MatchString(a.State) {
		return false
	}
	digits := 0
	for _, ch := range p {
		if ch >= '0' && ch <= '9' {
			digits++
		}
	}
	if digits < 8 || digits > 15 {
		return false
	}
	return strings.TrimSpace(a.Street) != "" && strings.TrimSpace(a.Number) != "" && strings.TrimSpace(a.City) != ""
}

func add(ctx context.Context, u unit, id int, sel menu.Selection, at time.Time) (json.RawMessage, error) {
	if !validID(sel.ItemID) || !validID(sel.Quantity) || len(sel.Notes) > 2000 {
		return nil, errInvalid
	}
	active, err := u.lines(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(active) >= 100 {
		return nil, errConflict
	}
	line, err := u.price(ctx, sel, at)
	if err != nil {
		return nil, err
	}
	var total int64
	for _, v := range active {
		if total > math.MaxInt64-v.Snapshot.LineTotalCents {
			return nil, errInvalid
		}
		total += v.Snapshot.LineTotalCents
	}
	if total > math.MaxInt64-line.LineTotalCents {
		return nil, errInvalid
	}
	v, err := u.add(ctx, id, sel, line)
	if err != nil {
		return nil, err
	}
	return v, u.track(ctx, id, "cart_item_add", sel.ItemID, sel.Quantity)
}
func submit(ctx context.Context, u unit, id, actor int, p submitInput, at time.Time, freight int64) error {
	lines, err := u.lines(ctx, id)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return errConflict
	}
	c := sales.Cart{Codes: p.Codes, Saved: p.Saved, Coins: p.Coins}
	for _, v := range lines {
		c.Selections = append(c.Selections, v.Selection)
	}
	if err := u.finalize(ctx, id, actor, c, at, freight); err != nil {
		return err
	}
	return u.track(ctx, id, "order_submitted", 0, 0)
}
