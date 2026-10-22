//! The response envelope every endpoint returns, byte-compatible with the Go
//! API's `pkg/httputil.Response` so the web app can't tell the backends apart.

use axum::{
    Json,
    http::StatusCode,
    response::{IntoResponse, Response},
};
use serde::Serialize;

/// `{"success","message","data","error_code","errors","meta"}`. Empty fields are
/// omitted, like Go's `omitempty`, and keys keep this order.
#[derive(Debug, Serialize)]
pub struct Envelope<T: Serialize> {
    pub success: bool,
    #[serde(skip_serializing_if = "String::is_empty")]
    pub message: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub data: Option<T>,
    #[serde(skip_serializing_if = "String::is_empty")]
    pub error_code: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub errors: Option<Vec<FieldError>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub meta: Option<Meta>,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct FieldError {
    pub field: String,
    pub message: String,
}

impl FieldError {
    pub fn new(field: impl Into<String>, message: impl Into<String>) -> Self {
        Self {
            field: field.into(),
            message: message.into(),
        }
    }
}

#[derive(Debug, Serialize)]
pub struct Meta {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub pagination: Option<Pagination>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub summary: Option<serde_json::Value>,
}

/// Page info for list endpoints. `next_page`/`prev_page` are `null` at the ends.
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct Pagination {
    pub total_records: i64,
    pub total_pages: i64,
    pub current_page: i64,
    pub per_page: i64,
    pub count: i64,
    pub next_page: Option<i64>,
    pub prev_page: Option<i64>,
}

impl Pagination {
    /// Same arithmetic as the Go services: pages are 1-based and an empty
    /// result still reports page 1 of 0.
    pub fn new(total_records: i64, current_page: i64, per_page: i64, count: i64) -> Self {
        let total_pages = if per_page > 0 {
            (total_records + per_page - 1) / per_page
        } else {
            0
        };
        Self {
            total_records,
            total_pages,
            current_page,
            per_page,
            count,
            next_page: (current_page < total_pages).then_some(current_page + 1),
            prev_page: (current_page > 1).then_some(current_page - 1),
        }
    }
}

/// A status code plus an envelope; what handlers return on success.
#[derive(Debug)]
pub struct Reply<T: Serialize>(pub StatusCode, pub Envelope<T>);

impl<T: Serialize> IntoResponse for Reply<T> {
    fn into_response(self) -> Response {
        (self.0, Json(self.1)).into_response()
    }
}

/// `httputil.Success`.
pub fn success<T: Serialize>(status: StatusCode, message: impl Into<String>, data: T) -> Reply<T> {
    Reply(
        status,
        Envelope {
            success: true,
            message: message.into(),
            data: Some(data),
            error_code: String::new(),
            errors: None,
            meta: None,
        },
    )
}

/// `httputil.SuccessWithPagination`.
pub fn success_with_pagination<T: Serialize>(
    status: StatusCode,
    message: impl Into<String>,
    data: T,
    pagination: Pagination,
    summary: Option<serde_json::Value>,
) -> Reply<T> {
    Reply(
        status,
        Envelope {
            success: true,
            message: message.into(),
            data: Some(data),
            error_code: String::new(),
            errors: None,
            meta: Some(Meta {
                pagination: Some(pagination),
                summary,
            }),
        },
    )
}

/// An error envelope; used by `AppError` and the router fallbacks.
pub fn failure(
    status: StatusCode,
    message: impl Into<String>,
    error_code: impl Into<String>,
    errors: Option<Vec<FieldError>>,
) -> Response {
    Reply::<()>(
        status,
        Envelope {
            success: false,
            message: message.into(),
            data: None,
            error_code: error_code.into(),
            errors,
            meta: None,
        },
    )
    .into_response()
}
