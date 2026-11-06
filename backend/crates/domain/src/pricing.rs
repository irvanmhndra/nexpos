//! How an order is priced, computed like the Go `OrderService` with
//! `shopspring/decimal`: division rounded half away from zero at 16 places,
//! money half away from zero at 2.

use chrono::{DateTime, Utc};
use rust_decimal::{Decimal, RoundingStrategy};

const MONEY_SCALE: u32 = 2;
const DIVISION_PRECISION: u32 = 16;

/// The company settings pricing depends on.
#[derive(Debug, Clone, Copy)]
pub struct PriceRules {
    pub tax_enabled: bool,
    pub tax_rate: Decimal,
    pub tax_inclusive: bool,
    pub rounding_enabled: bool,
    pub rounding_amount: Decimal,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct LineAmounts {
    pub gross: Decimal,
    pub subtotal: Decimal,
    pub tax: Decimal,
}

fn div(a: Decimal, b: Decimal) -> Decimal {
    (a / b).round_dp_with_strategy(DIVISION_PRECISION, RoundingStrategy::MidpointAwayFromZero)
}

fn round(value: Decimal, scale: u32) -> Decimal {
    value.round_dp_with_strategy(scale, RoundingStrategy::MidpointAwayFromZero)
}

/// One order line. Tax is rounded per line, so an order's tax is exactly the sum
/// of its lines' tax.
pub fn line_amounts(
    unit_price: Decimal,
    quantity: i64,
    discount: Decimal,
    rules: &PriceRules,
) -> LineAmounts {
    let gross = unit_price * Decimal::from(quantity);
    let subtotal = gross - discount;
    let tax = if !rules.tax_enabled {
        Decimal::ZERO
    } else if rules.tax_inclusive {
        round(
            div(
                subtotal * rules.tax_rate,
                Decimal::ONE_HUNDRED + rules.tax_rate,
            ),
            MONEY_SCALE,
        )
    } else {
        round(
            div(subtotal * rules.tax_rate, Decimal::ONE_HUNDRED),
            MONEY_SCALE,
        )
    };
    LineAmounts {
        gross,
        subtotal,
        tax,
    }
}

/// What the customer pays: gross minus discounts, plus tax only when prices
/// exclude it. Cash rounding applies last.
pub fn grand_total(gross: Decimal, discount: Decimal, tax: Decimal, rules: &PriceRules) -> Decimal {
    let mut total = gross - discount;
    if !rules.tax_inclusive {
        total += tax;
    }
    if rules.rounding_enabled && rules.rounding_amount > Decimal::ZERO {
        total = round(div(total, rules.rounding_amount), 0) * rules.rounding_amount;
    }
    total
}

/// The sale price while it is lower and within its window, otherwise the price.
pub fn effective_price(
    price: Decimal,
    sale_price: Option<Decimal>,
    sale_start: Option<DateTime<Utc>>,
    sale_end: Option<DateTime<Utc>>,
    now: DateTime<Utc>,
) -> Decimal {
    match sale_price {
        Some(sale)
            if sale < price
                && sale_start.is_none_or(|start| now >= start)
                && sale_end.is_none_or(|end| now <= end) =>
        {
            sale
        }
        _ => price,
    }
}

/// A promotion's terms.
#[derive(Debug, Clone, Copy)]
pub struct PromoTerms<'a> {
    pub discount_type: Option<&'a str>,
    pub discount_value: Option<Decimal>,
    pub min_purchase: Option<Decimal>,
    pub max_discount: Option<Decimal>,
}

/// The discount a promotion gives on this subtotal; `None` when it doesn't apply.
pub fn promotion_discount(terms: &PromoTerms<'_>, subtotal: Decimal) -> Option<Decimal> {
    let kind = terms.discount_type?;
    let value = terms.discount_value?;
    if terms.min_purchase.is_some_and(|min| subtotal < min) {
        return None;
    }
    Some(match kind {
        "percentage" => {
            let discount = round(div(subtotal * value, Decimal::ONE_HUNDRED), MONEY_SCALE);
            match terms.max_discount {
                Some(max) if discount > max => max,
                _ => discount,
            }
        }
        "fixed" => value.min(subtotal),
        _ => Decimal::ZERO,
    })
}

/// An order's payment status after `total_paid`.
pub fn payment_status(total_paid: Decimal, grand_total: Decimal) -> &'static str {
    if total_paid <= Decimal::ZERO {
        "unpaid"
    } else if total_paid >= grand_total {
        "paid"
    } else {
        "partial"
    }
}

#[cfg(test)]
mod tests {
    use std::str::FromStr;

    use chrono::Duration;

    use super::*;

    fn d(text: &str) -> Decimal {
        Decimal::from_str(text).unwrap()
    }

    fn rules(rate: &str, inclusive: bool) -> PriceRules {
        PriceRules {
            tax_enabled: true,
            tax_rate: d(rate),
            tax_inclusive: inclusive,
            rounding_enabled: false,
            rounding_amount: Decimal::ZERO,
        }
    }

    #[test]
    fn exclusive_tax_is_added_and_inclusive_tax_is_extracted() {
        let line = line_amounts(d("100"), 3, d("10.5"), &rules("11", false));
        assert_eq!(line.gross, d("300"));
        assert_eq!(line.subtotal, d("289.5"));
        assert_eq!(line.tax, d("31.85"));
        assert_eq!(
            line_amounts(d("100"), 3, d("10.5"), &rules("11", true)).tax,
            d("28.69")
        );
    }

    #[test]
    fn no_tax_when_disabled() {
        let mut off = rules("11", false);
        off.tax_enabled = false;
        assert_eq!(line_amounts(d("100"), 1, d("0"), &off).tax, Decimal::ZERO);
    }

    #[test]
    fn half_cents_round_away_from_zero() {
        // 12.5 at 1% is 0.125.
        assert_eq!(
            line_amounts(d("12.5"), 1, d("0"), &rules("1", false)).tax,
            d("0.13")
        );
    }

    #[test]
    fn grand_total_adds_only_exclusive_tax_and_rounds_last() {
        assert_eq!(
            grand_total(d("300"), d("10.5"), d("31.85"), &rules("11", false)),
            d("321.35")
        );
        assert_eq!(
            grand_total(d("300"), d("10.5"), d("28.69"), &rules("11", true)),
            d("289.5")
        );
        let mut cash = rules("11", true);
        cash.rounding_enabled = true;
        cash.rounding_amount = d("100");
        assert_eq!(
            grand_total(d("300"), d("10.5"), d("28.69"), &cash),
            d("300")
        );
        cash.rounding_amount = Decimal::ZERO;
        assert_eq!(
            grand_total(d("300"), d("10.5"), d("28.69"), &cash),
            d("289.5")
        );
    }

    #[test]
    fn sale_price_only_when_lower_and_in_window() {
        let now = Utc::now();
        let day = Duration::days(1);
        let p = d("100");
        assert_eq!(effective_price(p, None, None, None, now), p);
        assert_eq!(effective_price(p, Some(d("90")), None, None, now), d("90"));
        assert_eq!(effective_price(p, Some(d("110")), None, None, now), p);
        assert_eq!(
            effective_price(p, Some(d("90")), Some(now + day), None, now),
            p
        );
        assert_eq!(
            effective_price(p, Some(d("90")), None, Some(now - day), now),
            p
        );
        assert_eq!(
            effective_price(p, Some(d("90")), Some(now), Some(now), now),
            d("90")
        );
    }

    #[test]
    fn promotion_terms() {
        let percent = PromoTerms {
            discount_type: Some("percentage"),
            discount_value: Some(d("10")),
            min_purchase: None,
            max_discount: Some(d("20")),
        };
        assert_eq!(promotion_discount(&percent, d("100")), Some(d("10")));
        assert_eq!(promotion_discount(&percent, d("289.5")), Some(d("20")));
        let fixed = PromoTerms {
            discount_type: Some("fixed"),
            discount_value: Some(d("50")),
            min_purchase: Some(d("10")),
            max_discount: None,
        };
        assert_eq!(promotion_discount(&fixed, d("30")), Some(d("30")));
        assert_eq!(promotion_discount(&fixed, d("9")), None);
        let bare = PromoTerms {
            discount_type: None,
            discount_value: None,
            min_purchase: None,
            max_discount: None,
        };
        assert_eq!(promotion_discount(&bare, d("9")), None);
    }

    #[test]
    fn payment_statuses() {
        assert_eq!(payment_status(d("0"), d("10")), "unpaid");
        assert_eq!(payment_status(d("5"), d("10")), "partial");
        assert_eq!(payment_status(d("10"), d("10")), "paid");
        assert_eq!(payment_status(d("11"), d("10")), "paid");
    }
}
