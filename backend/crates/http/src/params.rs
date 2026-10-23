//! Path and query parameters parsed the way the Go handlers do it
//! (`strconv.ParseInt`, `strconv.Atoi`, `strconv.ParseBool`, Echo's `QueryParam`).

use axum::extract::{FromRequestParts, Path, Query};
use axum::http::request::Parts;

use crate::AppError;

/// A numeric `:id` path segment; anything else is 400 "invalid <resource> ID".
pub fn parse_id(raw: &str, resource: &str) -> Result<i64, AppError> {
    raw.parse()
        .map_err(|_| AppError::bad_request(format!("invalid {resource} ID")))
}

/// The raw `{id}` segment, parsed per resource by `parse_id`.
pub struct RawId(pub String);

impl<S: Send + Sync> FromRequestParts<S> for RawId {
    type Rejection = AppError;

    async fn from_request_parts(parts: &mut Parts, state: &S) -> Result<Self, AppError> {
        Path::<String>::from_request_parts(parts, state)
            .await
            .map(|Path(id)| Self(id))
            .map_err(|_| AppError::bad_request("invalid ID"))
    }
}

/// Query parameters; `get` returns the first value like Echo's `QueryParam`.
pub struct QueryParams(Vec<(String, String)>);

impl QueryParams {
    pub fn get(&self, key: &str) -> &str {
        self.0
            .iter()
            .find(|(k, _)| k == key)
            .map_or("", |(_, v)| v.as_str())
    }

    /// `strconv.Atoi`; unparseable values are ignored.
    pub fn int(&self, key: &str) -> Option<i64> {
        self.get(key).parse().ok()
    }

    /// `strconv.ParseBool`; unparseable values are ignored.
    pub fn bool(&self, key: &str) -> Option<bool> {
        match self.get(key) {
            "1" | "t" | "T" | "TRUE" | "true" | "True" => Some(true),
            "0" | "f" | "F" | "FALSE" | "false" | "False" => Some(false),
            _ => None,
        }
    }

    /// Page and page size with the Go services' defaults: page ≥ 1, size 1–100 (20 if unset).
    pub fn page(&self) -> (i64, i64) {
        let page = self.int("page").filter(|p| *p >= 1).unwrap_or(1);
        let per_page = match self.int("per_page") {
            Some(n) if n > 100 => 100,
            Some(n) if n >= 1 => n,
            _ => 20,
        };
        (page, per_page)
    }
}

impl<S: Send + Sync> FromRequestParts<S> for QueryParams {
    type Rejection = std::convert::Infallible;

    async fn from_request_parts(parts: &mut Parts, state: &S) -> Result<Self, Self::Rejection> {
        let pairs = Query::<Vec<(String, String)>>::from_request_parts(parts, state)
            .await
            .map(|Query(p)| p)
            .unwrap_or_default();
        Ok(Self(pairs))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn q(pairs: &[(&str, &str)]) -> QueryParams {
        QueryParams(
            pairs
                .iter()
                .map(|(k, v)| (k.to_string(), v.to_string()))
                .collect(),
        )
    }

    #[test]
    fn paging_defaults_match_go() {
        assert_eq!(q(&[]).page(), (1, 20));
        assert_eq!(q(&[("page", "0"), ("per_page", "0")]).page(), (1, 20));
        assert_eq!(q(&[("page", "3"), ("per_page", "500")]).page(), (3, 100));
        assert_eq!(q(&[("page", "abc"), ("per_page", "-5")]).page(), (1, 20));
        assert_eq!(q(&[("page", "+2")]).page(), (2, 20));
    }

    #[test]
    fn bools_and_first_value_win() {
        assert_eq!(q(&[("a", "T")]).bool("a"), Some(true));
        assert_eq!(q(&[("a", "False")]).bool("a"), Some(false));
        assert_eq!(q(&[("a", "yes")]).bool("a"), None);
        assert_eq!(q(&[("a", "1"), ("a", "0")]).bool("a"), Some(true));
    }

    #[test]
    fn ids() {
        assert_eq!(parse_id("42", "branch").unwrap(), 42);
        assert!(
            matches!(parse_id("abc", "branch"), Err(AppError::BadRequest(m)) if m == "invalid branch ID")
        );
    }
}
