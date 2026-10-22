package model

import "github.com/shopspring/decimal"

// Money and rates are decimal.Decimal: the database columns are NUMERIC, and
// decimal.Decimal cannot represent most rupiah-and-cents values exactly, so sums and
// tax drifted by fractions of a cent before being rounded on insert.
func init() {
	// Keep money as JSON numbers so API clients see the same shape as before;
	// decoding accepts both numbers and strings.
	decimal.MarshalJSONWithoutQuotes = true
}

// MoneyScale is the number of decimal places stored for money (NUMERIC(15,2)).
const MoneyScale = 2
