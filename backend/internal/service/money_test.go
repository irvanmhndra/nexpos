package service

import (
	"testing"

	"github.com/shopspring/decimal"
)

// money builds a decimal from a literal; exact for the round values tests use.
func money(v float64) decimal.Decimal { return decimal.NewFromFloat(v) }

func moneyPtr(v float64) *decimal.Decimal {
	d := money(v)
	return &d
}

// assertMoney compares by value: decimal.Decimal values with different
// exponents (1000 vs 1000.00) are equal amounts but not deeply equal.
func assertMoney(t *testing.T, want float64, got decimal.Decimal, msgAndArgs ...any) {
	t.Helper()
	if !got.Equal(money(want)) {
		t.Errorf("money: want %v, got %s %v", want, got, msgAndArgs)
	}
}
