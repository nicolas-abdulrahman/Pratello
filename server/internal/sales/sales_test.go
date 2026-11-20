package sales

import (
	"testing"
	"time"

	"pratello/server/internal/menu"
)

func TestBestDiscountPerLine(t *testing.T) {
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

func TestFIN025Through027CouponEligibility(t *testing.T) {
	at := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	c := Coupon{ID: 1, Scope: "order", Active: true, Type: "fixed_cents", Value: 200, Minimum: 1000, Expires: &at}
	for _, tc := range []struct {
		name     string
		date     time.Time
		subtotal int64
		inactive bool
		reason   string
		saving   int64
	}{
		{"below minimum", at.Add(-time.Second), 999, false, "MIN_ORDER_NOT_MET", 0},
		{"at minimum", at.Add(-time.Second), 1000, false, "", 200},
		{"at expiry", at, 1000, false, "EXPIRED", 0},
		{"deactivated", at.Add(-time.Second), 1000, true, "INACTIVE", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			coupon := c
			coupon.Active = !tc.inactive
			if got := coupon.reason(tc.subtotal, tc.date); got != tc.reason {
				t.Fatalf("got reason %q want %q", got, tc.reason)
			}
			p, err := calculate([]menu.CheckoutLine{{LineSnapshot: menu.LineSnapshot{ItemID: 1, Quantity: 1, LineTotalCents: tc.subtotal}}}, []Coupon{coupon}, tc.date)
			if err != nil || p.Discount != tc.saving {
				t.Fatalf("discount %+v %v", p, err)
			}
		})
	}
}
