//! Request extractors that fail exactly like the Go API: `JsonBody` follows
//! Echo's binder plus `httputil.BindError`, `validate` follows go-playground's
//! messages from `pkg/validator`, and `AuthUser` follows `middleware.Auth`.

use std::net::SocketAddr;

use axum::{
    body::Bytes,
    extract::{ConnectInfo, FromRequest, FromRequestParts, Request},
    http::{StatusCode, header, request::Parts},
};
use chrono::Utc;
use nexpos_db::auth as db;
use nexpos_domain::auth::hash_token;
use serde::de::DeserializeOwned;

use crate::{AppError, AppState, envelope::FieldError};

/// A JSON request body. Missing or `null` fields fall back to `Default`, so
/// validation reports them the way Go does for zero values.
pub struct JsonBody<T>(pub T);

impl<S, T> FromRequest<S> for JsonBody<T>
where
    S: Send + Sync,
    T: DeserializeOwned + Default,
{
    type Rejection = AppError;

    async fn from_request(req: Request, state: &S) -> Result<Self, AppError> {
        let is_json = req
            .headers()
            .get(header::CONTENT_TYPE)
            .and_then(|v| v.to_str().ok())
            .map(|v| v.split(';').next().unwrap_or("").trim() == "application/json")
            .unwrap_or(false);
        let bytes = Bytes::from_request(req, state)
            .await
            .map_err(|_| AppError::bad_request("invalid request body"))?;

        // Echo skips binding an empty body; validation then reports the fields.
        if bytes.is_empty() {
            return Ok(Self(T::default()));
        }
        // Echo answers 415 for other media types, which BindError turns into this.
        if !is_json {
            return Err(AppError::bad_request("invalid request body"));
        }

        let de = &mut serde_json::Deserializer::from_slice(&bytes);
        serde_path_to_error::deserialize(de)
            .map(Self)
            .map_err(|err| {
                let field = err.path().to_string();
                let inner = err.into_inner();
                match inner.classify() {
                    serde_json::error::Category::Data => AppError::validation(
                        "Validation failed",
                        vec![FieldError::new(
                            field,
                            format!("must be {}", friendly_type(&inner.to_string())),
                        )],
                    ),
                    serde_json::error::Category::Syntax => {
                        AppError::bad_request("malformed JSON body")
                    }
                    serde_json::error::Category::Eof if bytes.trim_ascii().is_empty() => {
                        AppError::bad_request("request body is empty")
                    }
                    _ => AppError::bad_request("invalid request body"),
                }
            })
    }
}

/// "invalid type: string \"x\", expected i64 at line 1" → "a number", like
/// Go's `friendlyType`.
fn friendly_type(message: &str) -> String {
    let expected = message
        .split("expected ")
        .nth(1)
        .unwrap_or("")
        .split(" at line")
        .next()
        .unwrap_or("");
    if [
        "i8", "i16", "i32", "i64", "u8", "u16", "u32", "u64", "f32", "f64", "integer", "number",
        "float",
    ]
    .iter()
    .any(|t| expected.contains(t))
    {
        "a number".into()
    } else if expected.contains("string") {
        "a string".into()
    } else if expected.contains("bool") {
        "true or false".into()
    } else {
        expected.into()
    }
}

/// One go-playground tag. A field reports only its first failing rule.
#[derive(Clone, Copy)]
pub enum Rule {
    Required,
    Email,
    /// Minimum length in characters (runes), like go-playground's `min` on strings.
    Min(usize),
    Max(usize),
}

/// Collects field errors in declaration order, like `ValidateAndFormat`.
#[derive(Default)]
pub struct Validator {
    errors: Vec<FieldError>,
}

impl Validator {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn field(&mut self, name: &str, value: &str, rules: &[Rule]) -> &mut Self {
        let len = value.chars().count();
        let failed = rules.iter().find(|rule| match rule {
            Rule::Required => value.is_empty(),
            Rule::Email => !is_email(value),
            Rule::Min(n) => len < *n,
            Rule::Max(n) => len > *n,
        });
        if let Some(rule) = failed {
            let message = match rule {
                Rule::Required => format!("{name} is required"),
                Rule::Email => "Invalid email format".to_owned(),
                Rule::Min(n) => format!("{name} must be at least {n} characters"),
                Rule::Max(n) => format!("{name} must be at most {n} characters"),
            };
            self.errors.push(FieldError::new(name, message));
        }
        self
    }

    /// A `*string` field with `omitempty`: skipped only when absent or `null`;
    /// an empty string is still checked, as in go-playground.
    pub fn optional(&mut self, name: &str, value: Option<&str>, rules: &[Rule]) -> &mut Self {
        match value {
            Some(v) => self.field(name, v, rules),
            None => self,
        }
    }

    pub fn finish(&mut self) -> Result<(), AppError> {
        if self.errors.is_empty() {
            Ok(())
        } else {
            Err(AppError::validation(
                "Validation failed",
                std::mem::take(&mut self.errors),
            ))
        }
    }
}

/// `local@domain` with no whitespace; close enough to go-playground's `email`
/// for every address the web app sends.
fn is_email(value: &str) -> bool {
    match value.rsplit_once('@') {
        Some((local, domain)) => {
            !local.is_empty()
                && !domain.is_empty()
                && !value.chars().any(char::is_whitespace)
                && !domain.contains('@')
                && !domain.starts_with('.')
                && !domain.ends_with('.')
        }
        None => false,
    }
}

/// Where a request came from, stored on the session like Go does.
pub struct ClientInfo {
    /// The connection's peer address (Echo's `RealIP` without an IP extractor).
    pub ip: String,
    pub user_agent: String,
}

impl<S: Send + Sync> FromRequestParts<S> for ClientInfo {
    type Rejection = std::convert::Infallible;

    async fn from_request_parts(parts: &mut Parts, _: &S) -> Result<Self, Self::Rejection> {
        let ip = parts
            .extensions
            .get::<ConnectInfo<SocketAddr>>()
            .map(|c| c.0.ip().to_string())
            .unwrap_or_default();
        let user_agent = parts
            .headers
            .get(header::USER_AGENT)
            .and_then(|v| v.to_str().ok())
            .unwrap_or_default()
            .to_owned();
        Ok(Self { ip, user_agent })
    }
}

/// The signed-in user, from `Authorization: Bearer <access token>`.
#[derive(Debug, Clone, Copy)]
pub struct AuthUser {
    pub user_id: i64,
    pub company_id: i64,
    pub branch_id: Option<i64>,
}

impl FromRequestParts<AppState> for AuthUser {
    type Rejection = AppError;

    async fn from_request_parts(parts: &mut Parts, state: &AppState) -> Result<Self, AppError> {
        let token = parts
            .headers
            .get(header::AUTHORIZATION)
            .and_then(|v| v.to_str().ok())
            .and_then(|v| v.strip_prefix("Bearer "))
            .ok_or_else(|| AppError::unauthorized("unauthorized"))?;

        let session = db::session_by_access_hash(&state.db, &hash_token(token))
            .await
            .ok()
            .flatten()
            .ok_or_else(|| AppError::unauthorized("invalid or expired token"))?;

        if Utc::now() > session.access_token_expires_at {
            return Err(AppError::Coded {
                status: StatusCode::UNAUTHORIZED,
                code: "TOKEN_EXPIRED",
                message: "token expired".into(),
            });
        }

        // Best-effort, like Go: a failed touch never fails the request.
        let _ = db::touch_session(&state.db, session.id).await;

        Ok(Self {
            user_id: session.user_id,
            company_id: session.company_id,
            branch_id: session.branch_id,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn email_rule() {
        for ok in ["a@b.co", "test@example.com", "a.b+c@sub.example.id"] {
            assert!(is_email(ok), "{ok}");
        }
        for bad in [
            "",
            "invalid-email",
            "not-an-email",
            "@x.com",
            "a@",
            "a b@c.com",
            "a@.com",
        ] {
            assert!(!is_email(bad), "{bad}");
        }
    }

    #[test]
    fn first_failing_rule_per_field() {
        let err = Validator::new()
            .field("name", "", &[Rule::Required, Rule::Min(3)])
            .field("email", "x", &[Rule::Required, Rule::Email])
            .field("password", "short", &[Rule::Required, Rule::Min(8)])
            .finish()
            .unwrap_err();
        let AppError::Validation { errors, .. } = err else {
            panic!("not a validation error")
        };
        assert_eq!(
            errors,
            vec![
                FieldError::new("name", "name is required"),
                FieldError::new("email", "Invalid email format"),
                FieldError::new("password", "password must be at least 8 characters"),
            ]
        );
    }

    #[test]
    fn friendly_types() {
        assert_eq!(
            friendly_type("invalid type: string \"x\", expected i64 at line 1 column 9"),
            "a number"
        );
        assert_eq!(
            friendly_type("invalid type: integer `5`, expected a string at line 1 column 10"),
            "a string"
        );
        assert_eq!(
            friendly_type("invalid type: string \"y\", expected a boolean at line 1 column 5"),
            "true or false"
        );
    }
}
