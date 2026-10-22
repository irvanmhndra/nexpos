//! `AppError`: the Rust side of Go's `pkg/apperror`. Status codes, `error_code`
//! values, and messages match, so clients see the same errors from either backend.

use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
};

use crate::envelope::{FieldError, failure};

#[derive(Debug, thiserror::Error)]
pub enum AppError {
    /// 400 `BAD_REQUEST`
    #[error("{0}")]
    BadRequest(String),
    /// 422 `VALIDATION_ERROR` with per-field messages
    #[error("{message}")]
    Validation {
        message: String,
        errors: Vec<FieldError>,
    },
    /// 401 `UNAUTHORIZED`
    #[error("{0}")]
    Unauthorized(String),
    /// 403 `FORBIDDEN`
    #[error("{0}")]
    Forbidden(String),
    /// 404 `NOT_FOUND`; holds the resource name ("customer" → "customer not found")
    #[error("{0} not found")]
    NotFound(String),
    /// 409 `CONFLICT`
    #[error("{0}")]
    Conflict(String),
    /// A domain error with its own code, e.g. `EMAIL_ALREADY_EXISTS` or `TOKEN_EXPIRED`.
    #[error("{message}")]
    Coded {
        status: StatusCode,
        code: &'static str,
        message: String,
    },
    /// 500 `INTERNAL_ERROR`. The cause is logged, never sent to the client.
    #[error(transparent)]
    Internal(#[from] anyhow::Error),
}

pub type AppResult<T> = Result<T, AppError>;

impl AppError {
    pub fn bad_request(message: impl Into<String>) -> Self {
        Self::BadRequest(message.into())
    }

    /// `apperror.Unauthorized`: an empty message becomes "unauthorized".
    pub fn unauthorized(message: impl Into<String>) -> Self {
        let message = message.into();
        Self::Unauthorized(if message.is_empty() {
            "unauthorized".into()
        } else {
            message
        })
    }

    pub fn not_found(resource: impl Into<String>) -> Self {
        Self::NotFound(resource.into())
    }

    pub fn validation(message: impl Into<String>, errors: Vec<FieldError>) -> Self {
        Self::Validation {
            message: message.into(),
            errors,
        }
    }

    pub fn email_already_exists() -> Self {
        Self::Coded {
            status: StatusCode::CONFLICT,
            code: "EMAIL_ALREADY_EXISTS",
            message: "email already exists".into(),
        }
    }

    pub fn invalid_credentials() -> Self {
        Self::Coded {
            status: StatusCode::UNAUTHORIZED,
            code: "INVALID_CREDENTIALS",
            message: "invalid email or password".into(),
        }
    }

    pub fn user_inactive() -> Self {
        Self::Coded {
            status: StatusCode::FORBIDDEN,
            code: "USER_INACTIVE",
            message: "user account is inactive".into(),
        }
    }

    pub fn status(&self) -> StatusCode {
        match self {
            Self::BadRequest(_) => StatusCode::BAD_REQUEST,
            Self::Validation { .. } => StatusCode::UNPROCESSABLE_ENTITY,
            Self::Unauthorized(_) => StatusCode::UNAUTHORIZED,
            Self::Forbidden(_) => StatusCode::FORBIDDEN,
            Self::NotFound(_) => StatusCode::NOT_FOUND,
            Self::Conflict(_) => StatusCode::CONFLICT,
            Self::Coded { status, .. } => *status,
            Self::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
        }
    }

    pub fn code(&self) -> &'static str {
        match self {
            Self::BadRequest(_) => "BAD_REQUEST",
            Self::Validation { .. } => "VALIDATION_ERROR",
            Self::Unauthorized(_) => "UNAUTHORIZED",
            Self::Forbidden(_) => "FORBIDDEN",
            Self::NotFound(_) => "NOT_FOUND",
            Self::Conflict(_) => "CONFLICT",
            Self::Coded { code, .. } => code,
            Self::Internal(_) => "INTERNAL_ERROR",
        }
    }
}

impl From<sqlx::Error> for AppError {
    fn from(err: sqlx::Error) -> Self {
        Self::Internal(err.into())
    }
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let status = self.status();
        let code = self.code();
        match self {
            // Go's InternalError echoes the cause to the client; here it only
            // goes to the log.
            Self::Internal(err) => {
                tracing::error!(error = ?err, "internal error");
                failure(status, "internal server error", code, None)
            }
            Self::Validation { message, errors } => failure(status, message, code, Some(errors)),
            other => failure(status, other.to_string(), code, None),
        }
    }
}
