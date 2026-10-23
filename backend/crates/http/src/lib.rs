//! HTTP layer of the nexpos Rust API: routes, the response envelope, errors,
//! and middleware. Wire-compatible with the Go API in `backend-go/`.

mod auth;
mod branches;
mod customers;
pub mod envelope;
pub mod error;
pub mod extract;
mod health;
pub mod params;
mod proxy;
pub mod state;
pub mod time;

use std::time::Duration;

use axum::{
    Router,
    extract::DefaultBodyLimit,
    http::{HeaderName, Method, StatusCode, header},
    response::Response,
};
use tower_http::{
    catch_panic::CatchPanicLayer,
    cors::{Any, CorsLayer},
    request_id::{MakeRequestUuid, PropagateRequestIdLayer, SetRequestIdLayer},
    timeout::TimeoutLayer,
    trace::TraceLayer,
};

pub use error::{AppError, AppResult};
pub use state::{AppState, AuthConfig};

const REQUEST_ID: HeaderName = HeaderName::from_static("x-request-id");

/// Matches nginx's `client_max_body_size 15m` in front of the API.
const BODY_LIMIT: usize = 15 * 1024 * 1024;

/// The whole API. Paths Rust doesn't serve go to the Go API when an upstream is
/// configured, and otherwise get Echo's JSON 404; wrong methods get Echo's 405.
pub fn app(state: AppState) -> Router {
    Router::new()
        .merge(health::router())
        .merge(auth::router())
        .merge(branches::router())
        .merge(customers::router())
        .fallback(proxy::fallback)
        .method_not_allowed_fallback(method_not_allowed)
        .with_state(state)
        .layer(DefaultBodyLimit::max(BODY_LIMIT))
        .layer(TimeoutLayer::with_status_code(
            StatusCode::REQUEST_TIMEOUT,
            Duration::from_secs(30),
        ))
        .layer(CatchPanicLayer::custom(|_| panic_response()))
        .layer(cors())
        .layer(TraceLayer::new_for_http())
        .layer(PropagateRequestIdLayer::new(REQUEST_ID))
        .layer(SetRequestIdLayer::new(REQUEST_ID, MakeRequestUuid))
}

/// Same policy as the Go API: any origin, these methods and headers.
fn cors() -> CorsLayer {
    CorsLayer::new()
        .allow_origin(Any)
        .allow_methods([
            Method::GET,
            Method::POST,
            Method::PUT,
            Method::PATCH,
            Method::DELETE,
            Method::OPTIONS,
        ])
        .allow_headers([
            header::ACCEPT,
            header::AUTHORIZATION,
            header::CONTENT_TYPE,
            REQUEST_ID,
        ])
}

async fn method_not_allowed() -> Response {
    envelope::failure(
        StatusCode::METHOD_NOT_ALLOWED,
        "Method Not Allowed",
        "Method Not Allowed",
        None,
    )
}

fn panic_response() -> Response {
    tracing::error!("handler panicked");
    envelope::failure(
        StatusCode::INTERNAL_SERVER_ERROR,
        "internal server error",
        "INTERNAL_ERROR",
        None,
    )
}
