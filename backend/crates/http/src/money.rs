//! Decimals on the wire, like `shopspring/decimal` with
//! `MarshalJSONWithoutQuotes`: written as JSON numbers without trailing zeros,
//! read from a JSON number or a quoted string.

use std::str::FromStr;

use rust_decimal::Decimal;
use serde::{Deserialize, Deserializer, Serialize, Serializer, de::Error};
use serde_json::value::RawValue;

/// Marks a decimal that failed to parse; `JsonBody` maps it to Go's bad-body error.
pub const INVALID_DECIMAL: &str = "invalid decimal value";

/// `11.50` → `11.5`, `100.00` → `100`, exactly what Go's `Decimal.String` prints.
pub fn render(value: &Decimal) -> String {
    value.normalize().to_string()
}

fn parse(text: &str) -> Option<Decimal> {
    let text = text.trim();
    let text = text
        .strip_prefix('"')
        .and_then(|t| t.strip_suffix('"'))
        .unwrap_or(text);
    Decimal::from_str(text)
        .or_else(|_| Decimal::from_scientific(text))
        .ok()
}

/// `#[serde(serialize_with = "crate::money::serialize")]`
pub fn serialize<S: Serializer>(value: &Decimal, s: S) -> Result<S::Ok, S::Error> {
    RawValue::from_string(render(value))
        .map_err(serde::ser::Error::custom)?
        .serialize(s)
}

/// For optional decimals; pair with `skip_serializing_if = "Option::is_none"`.
pub fn serialize_some<S: Serializer>(value: &Option<Decimal>, s: S) -> Result<S::Ok, S::Error> {
    match value {
        Some(v) => serialize(v, s),
        None => s.serialize_none(),
    }
}

/// `#[serde(serialize_with = "crate::money::serialize_float")]`: a float the way
/// Go's encoder writes it, so `50.0` is `50` (serde_json would write `50.0`).
pub fn serialize_float<S: Serializer>(value: &f64, s: S) -> Result<S::Ok, S::Error> {
    RawValue::from_string(value.to_string())
        .map_err(serde::ser::Error::custom)?
        .serialize(s)
}

/// `#[serde(deserialize_with = "crate::money::deserialize_opt")]` for
/// `*decimal.Decimal` fields: `null` or missing is `None`.
pub fn deserialize_opt<'de, D: Deserializer<'de>>(d: D) -> Result<Option<Decimal>, D::Error> {
    let raw: Option<Box<RawValue>> = Option::deserialize(d)?;
    raw.map(|r| parse(r.get()).ok_or_else(|| D::Error::custom(INVALID_DECIMAL)))
        .transpose()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn renders_without_trailing_zeros() {
        for (input, want) in [
            ("11.50", "11.5"),
            ("100.00", "100"),
            ("0.00", "0"),
            ("11.555", "11.555"),
            ("-3.10", "-3.1"),
        ] {
            assert_eq!(render(&Decimal::from_str(input).unwrap()), want);
        }
    }

    #[test]
    fn parses_numbers_and_quoted_strings() {
        assert_eq!(parse("11.5"), Decimal::from_str("11.5").ok());
        assert_eq!(parse("\"100.50\""), Decimal::from_str("100.50").ok());
        assert_eq!(parse("1e2"), Decimal::from_str("100").ok());
        assert_eq!(parse("\"abc\""), None);
        assert_eq!(parse("true"), None);
    }
}
