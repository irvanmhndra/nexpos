//! Wire compatibility with the Go API: the exact JSON bodies, statuses, and
//! headers it produces. Expected strings were captured from the Go server.

use axum::{
    Router,
    body::Body,
    http::{Request, StatusCode, header},
    response::IntoResponse,
};
use http_body_util::BodyExt;
use nexpos_http::{
    AppError, AppState,
    envelope::{FieldError, Pagination, success, success_with_pagination},
};
use serde_json::json;
use sqlx::postgres::{PgConnectOptions, PgPoolOptions};
use tower::ServiceExt;

/// A pool that never connects: port 1 refuses, so the DB looks down.
fn unreachable_db() -> AppState {
    let opts = PgConnectOptions::new()
        .host("127.0.0.1")
        .port(1)
        .username("nobody")
        .database("none");
    let db = PgPoolOptions::new()
        .acquire_timeout(std::time::Duration::from_millis(500))
        .connect_lazy_with(opts);
    AppState::new(db)
}

/// The real database when DATABASE_URL is set, so /health can report healthy.
async fn database() -> Option<AppState> {
    let url = std::env::var("DATABASE_URL").ok()?;
    let db = PgPoolOptions::new()
        .connect(&url)
        .await
        .expect("connect DATABASE_URL");
    Some(AppState::new(db))
}

async fn send(app: Router, req: Request<Body>) -> (StatusCode, axum::http::HeaderMap, String) {
    let res = app.oneshot(req).await.unwrap();
    let status = res.status();
    let headers = res.headers().clone();
    let body = res.into_body().collect().await.unwrap().to_bytes();
    (status, headers, String::from_utf8(body.to_vec()).unwrap())
}

async fn body_of(res: impl IntoResponse) -> (StatusCode, String) {
    let res = res.into_response();
    let status = res.status();
    let body = res.into_body().collect().await.unwrap().to_bytes();
    (status, String::from_utf8(body.to_vec()).unwrap())
}

fn get(path: &str) -> Request<Body> {
    Request::get(path).body(Body::empty()).unwrap()
}

#[tokio::test]
async fn success_envelope_matches_go() {
    let (status, body) = body_of(success(
        StatusCode::CREATED,
        "Customer created successfully",
        json!({"id": 1}),
    ))
    .await;
    assert_eq!(status, StatusCode::CREATED);
    assert_eq!(
        body,
        r#"{"success":true,"message":"Customer created successfully","data":{"id":1}}"#
    );
}

#[tokio::test]
async fn empty_list_is_an_array() {
    let (_, body) = body_of(success(StatusCode::OK, "ok", Vec::<i64>::new())).await;
    assert_eq!(body, r#"{"success":true,"message":"ok","data":[]}"#);
}

#[tokio::test]
async fn paginated_envelope_matches_go() {
    let page = Pagination::new(0, 1, 20, 0);
    let (_, body) = body_of(success_with_pagination(
        StatusCode::OK,
        "Customers retrieved",
        Vec::<i64>::new(),
        page,
        None,
    ))
    .await;
    assert_eq!(
        body,
        r#"{"success":true,"message":"Customers retrieved","data":[],"meta":{"pagination":{"total_records":0,"total_pages":0,"current_page":1,"per_page":20,"count":0,"next_page":null,"prev_page":null}}}"#
    );
}

#[test]
fn pagination_pages() {
    let p = Pagination::new(45, 2, 20, 20);
    assert_eq!(
        (p.total_pages, p.next_page, p.prev_page),
        (3, Some(3), Some(1))
    );
    let last = Pagination::new(45, 3, 20, 5);
    assert_eq!((last.next_page, last.prev_page), (None, Some(2)));
}

#[tokio::test]
async fn errors_match_go_apperror() {
    let cases = [
        (
            AppError::bad_request("invalid customer ID"),
            400,
            r#"{"success":false,"message":"invalid customer ID","error_code":"BAD_REQUEST"}"#,
        ),
        (
            AppError::not_found("customer"),
            404,
            r#"{"success":false,"message":"customer not found","error_code":"NOT_FOUND"}"#,
        ),
        (
            AppError::unauthorized(""),
            401,
            r#"{"success":false,"message":"unauthorized","error_code":"UNAUTHORIZED"}"#,
        ),
        (
            AppError::email_already_exists(),
            409,
            r#"{"success":false,"message":"email already exists","error_code":"EMAIL_ALREADY_EXISTS"}"#,
        ),
        (
            AppError::invalid_credentials(),
            401,
            r#"{"success":false,"message":"invalid email or password","error_code":"INVALID_CREDENTIALS"}"#,
        ),
        (
            AppError::user_inactive(),
            403,
            r#"{"success":false,"message":"user account is inactive","error_code":"USER_INACTIVE"}"#,
        ),
        (
            AppError::validation(
                "Validation failed",
                vec![FieldError::new("name", "name is required")],
            ),
            422,
            r#"{"success":false,"message":"Validation failed","error_code":"VALIDATION_ERROR","errors":[{"field":"name","message":"name is required"}]}"#,
        ),
    ];
    for (err, want_status, want_body) in cases {
        let (status, body) = body_of(err).await;
        assert_eq!(status.as_u16(), want_status, "{want_body}");
        assert_eq!(body, want_body);
    }
}

#[tokio::test]
async fn internal_errors_hide_the_cause() {
    let (status, body) = body_of(AppError::Internal(anyhow::anyhow!(
        "pq: password authentication failed"
    )))
    .await;
    assert_eq!(status, StatusCode::INTERNAL_SERVER_ERROR);
    assert_eq!(
        body,
        r#"{"success":false,"message":"internal server error","error_code":"INTERNAL_ERROR"}"#
    );
}

#[tokio::test]
async fn unknown_route_is_echo_404() {
    let (status, _, body) = send(nexpos_http::app(unreachable_db()), get("/api/v1/nope")).await;
    assert_eq!(status, StatusCode::NOT_FOUND);
    assert_eq!(
        body,
        r#"{"success":false,"message":"Not Found","error_code":"Not Found"}"#
    );
}

#[tokio::test]
async fn wrong_method_is_echo_405() {
    let req = Request::post("/health/live").body(Body::empty()).unwrap();
    let (status, _, body) = send(nexpos_http::app(unreachable_db()), req).await;
    assert_eq!(status, StatusCode::METHOD_NOT_ALLOWED);
    assert_eq!(
        body,
        r#"{"success":false,"message":"Method Not Allowed","error_code":"Method Not Allowed"}"#
    );
}

#[tokio::test]
async fn liveness_needs_no_database() {
    let (status, _, body) = send(nexpos_http::app(unreachable_db()), get("/health/live")).await;
    assert_eq!(status, StatusCode::OK);
    assert_eq!(
        body,
        r#"{"success":true,"message":"OK","data":{"status":"alive"}}"#
    );
}

#[tokio::test]
async fn health_reports_a_down_database() {
    let app = nexpos_http::app(unreachable_db());
    let (status, _, body) = send(app.clone(), get("/health")).await;
    assert_eq!(status, StatusCode::SERVICE_UNAVAILABLE);
    assert_eq!(
        body,
        r#"{"success":true,"message":"Health check completed","data":{"services":{"database":"disconnected"},"status":"unhealthy"}}"#
    );
    let (status, _, body) = send(app, get("/health/ready")).await;
    assert_eq!(status, StatusCode::SERVICE_UNAVAILABLE);
    assert_eq!(
        body,
        r#"{"success":true,"message":"Not ready","data":{"status":"not ready"}}"#
    );
}

#[tokio::test]
async fn health_reports_a_live_database() {
    let Some(state) = database().await else {
        eprintln!("skipped: DATABASE_URL not set");
        return;
    };
    let (status, _, body) = send(nexpos_http::app(state), get("/health")).await;
    assert_eq!(status, StatusCode::OK);
    assert_eq!(
        body,
        r#"{"success":true,"message":"Health check completed","data":{"services":{"database":"connected"},"status":"healthy"}}"#
    );
}

#[tokio::test]
async fn request_id_is_generated_or_propagated() {
    let app = nexpos_http::app(unreachable_db());
    let (_, headers, _) = send(app.clone(), get("/health/live")).await;
    assert!(headers.get("x-request-id").is_some_and(|v| !v.is_empty()));

    let req = Request::get("/health/live")
        .header("x-request-id", "abc-123")
        .body(Body::empty())
        .unwrap();
    let (_, headers, _) = send(app, req).await;
    assert_eq!(headers.get("x-request-id").unwrap(), "abc-123");
}

#[tokio::test]
async fn cors_preflight_allows_the_web_app() {
    let req = Request::options("/api/v1/customers")
        .header(header::ORIGIN, "https://nexpos.irvanmahendra.com")
        .header(header::ACCESS_CONTROL_REQUEST_METHOD, "PATCH")
        .header(
            header::ACCESS_CONTROL_REQUEST_HEADERS,
            "authorization,content-type",
        )
        .body(Body::empty())
        .unwrap();
    let (status, headers, _) = send(nexpos_http::app(unreachable_db()), req).await;
    assert!(status.is_success());
    assert_eq!(
        headers.get(header::ACCESS_CONTROL_ALLOW_ORIGIN).unwrap(),
        "*"
    );
    let methods = headers
        .get(header::ACCESS_CONTROL_ALLOW_METHODS)
        .unwrap()
        .to_str()
        .unwrap();
    assert!(methods.contains("PATCH"));
}
