package sales

import (
	"context"
	"errors"
	"math"
	"math/bits"
	"net/url"
	"regexp"
	"strings"
	"time"

	"pratello/server/internal/menu"
)

var (
	errInvalid     = errors.New("invalid promotion input")
	errMissing     = errors.New("promotion or checkout not found")
	errForbidden   = errors.New("promotion access forbidden")
	errConflict    = errors.New("promotion or checkout conflict")
	errUnavailable = errors.New("sales database unavailable")
)

// MapCheckoutError lets callers translate expected checkout failures to their
// own domain errors without hiding unexpected persistence failures.
func MapCheckoutError(err, invalid, missing, forbidden, conflict error) error {
	switch {
	case errors.Is(err, errInvalid), errors.Is(err, menu.ErrInvalid):
		return invalid
	case errors.Is(err, errMissing), errors.Is(err, menu.ErrMissing):
		return missing
	case errors.Is(err, errForbidden):
		return forbidden
	case errors.Is(err, errConflict), errors.Is(err, menu.ErrNotOrderable):
		return conflict
	}
	return err
}

type Coupon struct {
	ID          int        `json:"id"`
	Slug        string     `json:"restaurant_slug"`
	Code        string     `json:"code"`
	Name        string     `json:"name,omitempty"`
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	Type        string     `json:"discount_type"`
	Value       int64      `json:"discount_value"`
	Minimum     int64      `json:"min_order_cents"`
	Expires     *time.Time `json:"expires_at"`
	Max         *int       `json:"max_uses"`
	Used        int        `json:"used_count"`
	Active      bool       `json:"is_active"`
	Scope       string     `json:"target_scope"`
	Target      *int       `json:"target_item_id"`
	Banner      bool       `json:"is_banner"`
	Image       string     `json:"background_image_url,omitempty"`
	Created     time.Time  `json:"created_at,omitempty"`
}

type optional[T any] struct {
	set   bool
	value *T
}
type input struct {
	Code        *string             `json:"code"`
	Name        *string             `json:"name"`
	Title       *string             `json:"title"`
	Description *string             `json:"description"`
	Type        *string             `json:"discount_type"`
	Value       *int64              `json:"discount_value"`
	Minimum     *int64              `json:"min_order_cents"`
	Expires     optional[time.Time] `json:"expires_at"`
	Max         optional[int]       `json:"max_uses"`
	Active      *bool               `json:"is_active"`
	Scope       *string             `json:"target_scope"`
	Target      optional[int]       `json:"target_item_id"`
	Banner      *bool               `json:"is_banner"`
	Image       *string             `json:"background_image_url"`
}

type Cart struct {
	Selections []menu.Selection `json:"items"`
	Codes      []string         `json:"coupon_codes"`
	Saved      []int            `json:"saved_coupon_ids"`
	Coins      int64            `json:"coins"`
}

type Line struct {
	Item     menu.LineSnapshot `json:"item"`
	Subtotal int64             `json:"subtotal_cents"`
	Discount int64             `json:"discount_cents"`
	CouponID int               `json:"coupon_id,omitempty"`
}

type Applied struct {
	ID      int    `json:"coupon_id"`
	Code    string `json:"code"`
	Savings int64  `json:"discount_cents"`
}

type Price struct {
	Lines    []Line    `json:"lines"`
	Coupons  []Applied `json:"coupons"`
	Subtotal int64     `json:"subtotal_cents"`
	Discount int64     `json:"discount_cents"`
	Coins    int64     `json:"coins_redeemed"`
	Payable  int64     `json:"payable_cents"`
}

type validation struct {
	Valid   bool   `json:"valid"`
	Code    string `json:"code"`
	Reason  string `json:"reason,omitempty"`
	Savings *int64 `json:"discount_cents,omitempty"`
}

type page struct {
	Data  []Coupon `json:"data"`
	Page  int      `json:"page"`
	Size  int      `json:"page_size"`
	Total int      `json:"total"`
}

type request struct {
	op, slug, code                       string
	actor, branch, id, page, size, order int
	subtotal                             int64
	edit                                 input
	cart                                 Cart
}

type unit interface {
	loyaltyBalance(context.Context, int, string) (balance, error)
	owner(context.Context, int) (bool, error)
	published(context.Context, int) (bool, error)
	coupons(context.Context) ([]Coupon, error)
	target(context.Context, int) (bool, error)
	save(context.Context, Coupon) (Coupon, error)
	claim(context.Context, int, int) error
	claims(context.Context, int) (map[int]bool, error)
	lines(context.Context, int, []menu.Selection, time.Time) ([]menu.CheckoutLine, error)
	wallet(context.Context, int, bool) (int, int64, error)
	draft(context.Context, int, int, int) (int64, error)
	settle(context.Context, int, int, Price, int64) error
}

type repository interface {
	transact(context.Context, string, bool, func(unit) (any, error)) (any, error)
}
type Service struct {
	store repository
	now   func() time.Time
}

var codePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{0,49}$`)

func normalize(s string) string { return strings.ToUpper(strings.TrimSpace(s)) }
func positiveID(n int) bool     { return n > 0 && n <= math.MaxInt32 }
func assign[T any](dest *T, p *T) {
	if p != nil {
		*dest = *p
	}
}

func (s *Service) execute(ctx context.Context, r request) (any, error) {
	write := r.op == "create" || r.op == "update" || r.op == "deactivate" || r.op == "claim" && r.actor > 0 || r.op == "finalize"
	return s.store.transact(ctx, r.slug, write, func(u unit) (any, error) {
		switch r.op {
		case "create", "update", "deactivate", "list":
			ok, err := u.owner(ctx, r.actor)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, errForbidden
			}
		case "banners", "claim", "quote", "finalize":
			ok, err := u.published(ctx, r.branch)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, errMissing
			}
		}
		at := s.now()
		all, err := u.coupons(ctx)
		if err != nil {
			return nil, err
		}
		byCode := map[string]Coupon{}
		byID := map[int]Coupon{}
		for _, c := range all {
			byCode[normalize(c.Code)] = c
			byID[c.ID] = c
		}
		switch r.op {
		case "list":
			out := page{Data: []Coupon{}, Page: r.page, Size: r.size, Total: len(all)}
			start := min(int64(len(all)), int64(r.page-1)*int64(r.size))
			end := min(int64(len(all)), start+int64(r.size))
			out.Data = append(out.Data, all[start:end]...)
			return out, nil
		case "create", "update", "deactivate":
			old := byCode[normalize(r.code)]
			if r.op == "create" {
				old = Coupon{Slug: r.slug, Active: true, Scope: "order"}
			} else if old.ID == 0 {
				return nil, errMissing
			}
			if r.op == "deactivate" {
				old.Active = false
				_, err := u.save(ctx, old)
				return nil, err
			}
			c, err := prepare(ctx, u, old, r.edit)
			if err != nil {
				return nil, err
			}
			return u.save(ctx, c)
		case "validate":
			c := byCode[normalize(r.code)]
			if c.ID == 0 {
				return nil, errMissing
			}
			reason := c.reason(r.subtotal, at)
			if reason == "" {
				ok, err := targetLive(ctx, u, c)
				if err != nil {
					return nil, err
				}
				if !ok {
					reason = "INACTIVE"
				}
			}
			v := validation{Valid: reason == "", Code: c.Code, Reason: reason}
			if v.Valid && c.Scope == "order" {
				n := discount(c, r.subtotal)
				v.Savings = &n
			}
			return v, nil
		case "banners", "mine":
			claims := map[int]bool{}
			if r.op == "mine" {
				claims, err = u.claims(ctx, r.actor)
				if err != nil {
					return nil, err
				}
			}
			out := []Coupon{}
			for _, c := range all {
				if !(r.op == "banners" && c.Banner || claims[c.ID]) || c.reason(math.MaxInt64, at) != "" {
					continue
				}
				ok, err := targetLive(ctx, u, c)
				if err != nil {
					return nil, err
				}
				if ok {
					out = append(out, c)
				}
			}
			return out, nil
		case "claim":
			c := byID[r.id]
			if c.ID == 0 || !c.Banner || c.reason(math.MaxInt64, at) != "" {
				return nil, errMissing
			}
			ok, err := targetLive(ctx, u, c)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, errMissing
			}
			if r.actor > 0 {
				if err := u.claim(ctx, r.actor, c.ID); err != nil {
					return nil, err
				}
			}
			return c, nil
		case "quote", "finalize":
			var freight int64
			if r.op == "finalize" {
				freight, err = u.draft(ctx, r.order, r.branch, r.actor)
				if err != nil {
					return nil, err
				}
			}
			return checkout(ctx, u, r, byCode, byID, at, freight)
		}
		return nil, errInvalid
	})
}

func prepare(ctx context.Context, u unit, c Coupon, p input) (Coupon, error) {
	if p.Code != nil {
		if c.ID != 0 {
			return c, errInvalid
		}
		c.Code = normalize(*p.Code)
	}
	assign(&c.Name, p.Name)
	assign(&c.Title, p.Title)
	assign(&c.Description, p.Description)
	assign(&c.Type, p.Type)
	assign(&c.Value, p.Value)
	assign(&c.Minimum, p.Minimum)
	assign(&c.Active, p.Active)
	assign(&c.Scope, p.Scope)
	assign(&c.Banner, p.Banner)
	assign(&c.Image, p.Image)
	if p.Expires.set {
		c.Expires = p.Expires.value
	}
	if p.Max.set {
		c.Max = p.Max.value
	}
	if p.Target.set {
		c.Target = p.Target.value
	}
	if !codePattern.MatchString(c.Code) || c.Minimum < 0 || c.Value < 1 || c.Type != "fixed_cents" && (c.Type != "percent" || c.Value > 100) || c.Max != nil && !positiveID(*c.Max) {
		return c, errInvalid
	}
	if c.Scope == "order" {
		if c.Target != nil {
			return c, errInvalid
		}
	} else if (c.Scope != "item" && c.Scope != "combo") || c.Target == nil || !positiveID(*c.Target) {
		return c, errInvalid
	}
	if c.Image != "" {
		v, e := url.Parse(c.Image)
		if e != nil || v.User != nil || !(v.Hostname() != "" && (v.Scheme == "https" || v.Scheme == "http") || v.Host == "" && v.Scheme == "" && strings.HasPrefix(c.Image, "/media/")) || strings.ContainsAny(c.Image, "\\\r\n\t ") {
			return c, errInvalid
		}
	}
	if c.Banner && (strings.TrimSpace(c.Title) == "" || c.Image == "") {
		return c, errInvalid
	}
	ok, err := targetLive(ctx, u, c)
	if err != nil {
		return c, err
	}
	if !ok {
		return c, errInvalid
	}
	return c, nil
}

func targetLive(ctx context.Context, u unit, c Coupon) (bool, error) {
	if c.Target == nil {
		return true, nil
	}
	combo, err := u.target(ctx, *c.Target)
	if errors.Is(err, errMissing) {
		return false, nil
	}
	return combo == (c.Scope == "combo"), err
}

func (c Coupon) reason(subtotal int64, at time.Time) string {
	switch {
	case !c.Active:
		return "INACTIVE"
	case c.Expires != nil && !at.Before(*c.Expires):
		return "EXPIRED"
	case c.Max != nil && c.Used >= *c.Max:
		return "MAX_USES_REACHED"
	case subtotal < c.Minimum:
		return "MIN_ORDER_NOT_MET"
	}
	return ""
}

func checkout(ctx context.Context, u unit, r request, byCode map[string]Coupon, byID map[int]Coupon, at time.Time, freight int64) (Price, error) {
	if !positiveID(r.branch) || len(r.cart.Selections) == 0 || len(r.cart.Selections) > 100 || len(r.cart.Codes)+len(r.cart.Saved) > 50 || r.cart.Coins < 0 {
		return Price{}, errInvalid
	}
	if r.actor == 0 && (r.cart.Coins > 0 || len(r.cart.Saved) > 0) {
		return Price{}, errForbidden
	}
	var candidates []Coupon
	seen := map[int]bool{}
	add := func(c Coupon) error {
		if c.ID == 0 {
			return errMissing
		}
		if !seen[c.ID] {
			seen[c.ID] = true
			candidates = append(candidates, c)
		}
		return nil
	}
	for _, code := range r.cart.Codes {
		if err := add(byCode[normalize(code)]); err != nil {
			return Price{}, err
		}
	}
	if len(r.cart.Saved) > 0 {
		claimed, err := u.claims(ctx, r.actor)
		if err != nil {
			return Price{}, err
		}
		for _, id := range r.cart.Saved {
			if !claimed[id] {
				return Price{}, errMissing
			}
			if err := add(byID[id]); err != nil {
				return Price{}, err
			}
		}
	}
	lines, err := u.lines(ctx, r.branch, r.cart.Selections, at)
	if err != nil {
		return Price{}, err
	}
	p, err := calculate(lines, candidates, at)
	if err != nil {
		return p, err
	}
	if r.cart.Coins > p.Payable {
		return Price{}, errInvalid
	}
	walletID := 0
	if r.cart.Coins > 0 {
		var bal int64
		walletID, bal, err = u.wallet(ctx, r.actor, r.op == "finalize")
		if err != nil {
			return Price{}, err
		}
		if bal < r.cart.Coins {
			return Price{}, errConflict
		}
		p.Coins = r.cart.Coins
		p.Payable -= p.Coins
	}
	if r.op == "finalize" {
		if freight < 0 || freight > math.MaxInt64-p.Payable {
			return Price{}, errInvalid
		}
		if err := u.settle(ctx, r.order, walletID, p, freight); err != nil {
			return Price{}, err
		}
	}
	return p, nil
}

// ApplyToOrder finalizes pricing once for a draft checkout owned by actor. The
// order workflow supplies the customer's selections before payment is started.
func (s *Service) ApplyToOrder(ctx context.Context, slug string, branch, order, actor int, cart Cart) (Price, error) {
	if !positiveID(order) || !positiveID(branch) {
		return Price{}, errInvalid
	}
	v, err := s.execute(ctx, request{op: "finalize", slug: slug, branch: branch, order: order, actor: actor, cart: cart})
	if err != nil {
		return Price{}, err
	}
	return v.(Price), nil
}

func mulDiv(a, b, d int64) int64 {
	hi, lo := bits.Mul64(uint64(a), uint64(b))
	q, _ := bits.Div64(hi, lo, uint64(d))
	return int64(q)
}
func discount(c Coupon, base int64) int64 {
	if c.Type == "percent" {
		return mulDiv(base, c.Value, 100)
	}
	return min(c.Value, base)
}

// Allocate each coupon's fixed or percentage discount across its eligible lines
// in cents. A line can receive only one source of savings.
func calculate(lines []menu.CheckoutLine, coupons []Coupon, at time.Time) (Price, error) {
	p := Price{Lines: make([]Line, len(lines)), Coupons: []Applied{}}
	scopes := map[int]int64{}
	for i, l := range lines {
		q := int64(l.Quantity)
		if q <= 0 || l.SaleSavings < 0 || l.SaleSavings > math.MaxInt64/q || l.LineTotalCents < 0 {
			return Price{}, errInvalid
		}
		saved := l.SaleSavings * q
		if l.LineTotalCents > math.MaxInt64-saved || p.Subtotal > math.MaxInt64-l.LineTotalCents-saved {
			return Price{}, errInvalid
		}
		gross := l.LineTotalCents + saved
		p.Subtotal += gross
		scopes[l.ItemID] += gross
		snapshot := l.LineSnapshot
		snapshot.UnitPriceCents = gross / q
		snapshot.LineTotalCents = gross
		p.Lines[i] = Line{Item: snapshot, Subtotal: gross, Discount: saved}
	}
	var order Coupon
	var orderSavings int64
	for _, c := range coupons {
		if c.Scope != "order" || c.reason(p.Subtotal, at) != "" {
			continue
		}
		n := discount(c, p.Subtotal)
		if n > orderSavings || n > 0 && n == orderSavings && c.ID < order.ID {
			order, orderSavings = c, n
		}
	}
	apply := func(c Coupon, total int64, match func(menu.CheckoutLine) bool) {
		if total == 0 {
			return
		}
		d := discount(c, total)
		var cumulative, previous int64
		for i, l := range lines {
			if !match(l) {
				continue
			}
			v := &p.Lines[i]
			cumulative += v.Subtotal
			amount := mulDiv(d, cumulative, total)
			part := amount - previous
			previous = amount
			if part > v.Discount || part > 0 && part == v.Discount && v.CouponID != 0 && c.ID < v.CouponID {
				v.Discount, v.CouponID = part, c.ID
			}
		}
	}
	if order.ID != 0 {
		apply(order, p.Subtotal, func(menu.CheckoutLine) bool { return true })
	}
	for _, c := range coupons {
		if c.Target == nil || c.reason(p.Subtotal, at) != "" {
			continue
		}
		id := *c.Target
		if len(lines) == 0 {
			continue
		}
		var total int64
		for _, l := range lines {
			if l.ItemID == id && l.Combo == (c.Scope == "combo") {
				total = scopes[id]
				break
			}
		}
		apply(c, total, func(l menu.CheckoutLine) bool { return l.ItemID == id && l.Combo == (c.Scope == "combo") })
	}
	applied := map[int]int64{}
	for i := range p.Lines {
		l := &p.Lines[i]
		p.Discount += l.Discount
		if l.CouponID != 0 {
			applied[l.CouponID] += l.Discount
		}
	}
	for _, c := range coupons {
		if n := applied[c.ID]; n > 0 {
			p.Coupons = append(p.Coupons, Applied{c.ID, c.Code, n})
			delete(applied, c.ID)
		}
	}
	p.Payable = p.Subtotal - p.Discount
	return p, nil
}
