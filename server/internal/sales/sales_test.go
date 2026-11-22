package sales

import (
	"context"
	"testing"
	"time"

	"pratello/server/internal/menu"
)

func TestFIN012CouponDiscountLimits(t *testing.T) {
	for _, coupon := range []Coupon{
		{Code: "LIMIT", Scope: "order", Type: "percent", Value: 101},
		{Code: "LIMIT", Scope: "order", Type: "percent", Value: 0},
		{Code: "LIMIT", Scope: "order", Type: "fixed_cents", Value: -1},
	} {
		if _, err := prepare(context.Background(), nil, coupon, input{}); err == nil {
			t.Fatalf("accepted invalid discount: %+v", coupon)
		}
	}
	assertCouponEligibility(t, Coupon{ID: 1, Scope: "order", Active: true, Type: "fixed_cents", Value: 2000}, time.Time{}, 1000, "", 1000)
}

func TestFIN033BestDiscountPerLine(t *testing.T) {
	burger := 1
	lines := []menu.CheckoutLine{
		{LineSnapshot: menu.LineSnapshot{ItemID: burger, Quantity: 1, LineTotalCents: 2700}, SaleSavings: 300},
		{LineSnapshot: menu.LineSnapshot{ItemID: 2, Quantity: 1, LineTotalCents: 1000}},
	}
	p, err := calculate(lines, []Coupon{
		{ID: 1, Scope: "item", Target: &burger, Active: true, Type: "percent", Value: 20},
		{ID: 2, Scope: "order", Active: true, Type: "percent", Value: 5},
		{ID: 3, Scope: "order", Active: true, Type: "fixed_cents", Value: 100},
	}, time.Now())
	if err != nil || p.Subtotal != 4000 || p.Payable != 3350 || len(p.Coupons) != 2 || p.Lines[0].Discount != 600 || p.Lines[1].Discount != 50 {
		t.Fatalf("wrong winners: %+v %v", p, err)
	}
}

func TestFIN034ComboCouponsAndExtras(t *testing.T) {
	combo, standalone := 1, 2
	lines := []menu.CheckoutLine{
		{LineSnapshot: menu.LineSnapshot{ItemID: combo, Quantity: 2, LineTotalCents: 4400}, BasePrice: 2000, SaleSavings: 300, Combo: true},
		{LineSnapshot: menu.LineSnapshot{ItemID: standalone, Quantity: 1, LineTotalCents: 1000}},
	}
	p, err := calculate(lines, []Coupon{
		{ID: 1, Scope: "combo", Target: &combo, Active: true, Type: "percent", Value: 20},
		{ID: 2, Scope: "combo", Target: &standalone, Active: true, Type: "percent", Value: 100},
		{ID: 3, Scope: "order", Active: true, Type: "percent", Value: 10},
		{ID: 4, Scope: "order", Active: true, Type: "fixed_cents", Value: 300},
	}, time.Now())
	if err != nil || p.Subtotal != 6000 || p.Payable != 4900 || len(p.Coupons) != 2 || p.Lines[0].CouponID != 1 || p.Lines[1].CouponID != 3 {
		t.Fatalf("wrong combo or order discount: %+v %v", p, err)
	}
	full, err := calculate(lines, []Coupon{{ID: 1, Scope: "order", Active: true, Type: "fixed_cents", Value: 7000}}, time.Now())
	if err != nil || full.Payable != 0 || full.Discount != 6000 {
		t.Fatalf("fixed discount exceeded subtotal: %+v %v", full, err)
	}
}

func TestFIN025CouponExpiresAtDeadline(t *testing.T) {
	at := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	coupon := Coupon{ID: 1, Scope: "order", Active: true, Type: "fixed_cents", Value: 200, Expires: &at}
	assertCouponEligibility(t, coupon, at.Add(-time.Second), 1000, "", 200)
	assertCouponEligibility(t, coupon, at, 1000, "EXPIRED", 0)
}

func TestFIN026InactiveCouponIsRejected(t *testing.T) {
	coupon := Coupon{ID: 1, Scope: "order", Active: false, Type: "fixed_cents", Value: 200}
	assertCouponEligibility(t, coupon, time.Time{}, 1000, "INACTIVE", 0)
}

func TestFIN027CouponRequiresMinimumSubtotal(t *testing.T) {
	coupon := Coupon{ID: 1, Scope: "order", Active: true, Type: "fixed_cents", Value: 200, Minimum: 1000}
	assertCouponEligibility(t, coupon, time.Time{}, 999, "MIN_ORDER_NOT_MET", 0)
	assertCouponEligibility(t, coupon, time.Time{}, 1000, "", 200)
}

func assertCouponEligibility(t *testing.T, coupon Coupon, at time.Time, subtotal int64, reason string, saving int64) {
	t.Helper()
	if got := coupon.reason(subtotal, at); got != reason {
		t.Fatalf("got reason %q want %q", got, reason)
	}
	p, err := calculate([]menu.CheckoutLine{{LineSnapshot: menu.LineSnapshot{ItemID: 1, Quantity: 1, LineTotalCents: subtotal}}}, []Coupon{coupon}, at)
	if err != nil || p.Discount != saving {
		t.Fatalf("got discount %+v, error %v; want discount %d", p, err, saving)
	}
}
