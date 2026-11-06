//! Derived report metrics, computed like the Go `ReportService` with
//! `shopspring/decimal` (division rounded half away from zero at 16 places).

use rust_decimal::{Decimal, RoundingStrategy};

const MONEY_SCALE: u32 = 2;
const DIVISION_PRECISION: u32 = 16;

/// Revenue per order, rounded to money precision; zero without orders.
pub fn avg_order_value(revenue: Decimal, orders: i64) -> Decimal {
    if orders <= 0 {
        return Decimal::ZERO;
    }
    (revenue / Decimal::from(orders))
        .round_dp_with_strategy(MONEY_SCALE, RoundingStrategy::MidpointAwayFromZero)
}

/// Gross profit as a percentage of revenue, for display; zero without revenue.
pub fn gross_margin(revenue: Decimal, gross_profit: Decimal) -> f64 {
    if revenue <= Decimal::ZERO {
        return 0.0;
    }
    let ratio = (gross_profit / revenue)
        .round_dp_with_strategy(DIVISION_PRECISION, RoundingStrategy::MidpointAwayFromZero);
    // Through text so the float is the nearest one, like Go's InexactFloat64.
    (ratio * Decimal::ONE_HUNDRED)
        .normalize()
        .to_string()
        .parse()
        .unwrap_or(0.0)
}

#[cfg(test)]
mod tests {
    use std::str::FromStr;

    use super::*;

    fn d(s: &str) -> Decimal {
        Decimal::from_str(s).unwrap()
    }

    #[test]
    fn average_rounds_half_away_from_zero() {
        assert_eq!(avg_order_value(d("33333"), 2), d("16666.5"));
        assert_eq!(avg_order_value(d("233.33"), 3), d("77.78"));
        assert_eq!(avg_order_value(d("0.25"), 10), d("0.03"));
        assert_eq!(avg_order_value(d("100"), 0), Decimal::ZERO);
    }

    #[test]
    fn margin_matches_go() {
        assert_eq!(gross_margin(d("33333"), d("20333")), 60.99960999609996);
        assert_eq!(gross_margin(d("100"), d("50")), 50.0);
        assert_eq!(gross_margin(d("0"), d("0")), 0.0);
        assert_eq!(gross_margin(d("100"), d("-25")), -25.0);
    }
}
