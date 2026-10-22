//! `/health`, `/health/live`, `/health/ready`, answering exactly like the Go
//! `HealthHandler` (note: unhealthy is still `success: true`, with a 503).

use axum::{Router, extract::State, http::StatusCode, routing::get};
use serde_json::{Value, json};

use crate::{
    envelope::{Reply, success},
    state::AppState,
};

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/health", get(check))
        .route("/health/live", get(liveness))
        .route("/health/ready", get(readiness))
}

async fn db_ok(state: &AppState) -> bool {
    sqlx::query("SELECT 1").execute(&state.db).await.is_ok()
}

async fn check(State(state): State<AppState>) -> Reply<Value> {
    let (status, database, code) = if db_ok(&state).await {
        ("healthy", "connected", StatusCode::OK)
    } else {
        ("unhealthy", "disconnected", StatusCode::SERVICE_UNAVAILABLE)
    };
    success(
        code,
        "Health check completed",
        json!({ "status": status, "services": { "database": database } }),
    )
}

async fn liveness() -> Reply<Value> {
    success(StatusCode::OK, "OK", json!({ "status": "alive" }))
}

async fn readiness(State(state): State<AppState>) -> Reply<Value> {
    if db_ok(&state).await {
        success(StatusCode::OK, "Ready", json!({ "status": "ready" }))
    } else {
        success(
            StatusCode::SERVICE_UNAVAILABLE,
            "Not ready",
            json!({ "status": "not ready" }),
        )
    }
}
