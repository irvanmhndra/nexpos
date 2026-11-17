//! The platform panel end to end, against a real database: who may see which
//! company, what a support session can do, and that every write is audited.
//!
//! Needs `DATABASE_URL` (any database on the server; it is only used to create
//! and drop a throwaway database with the Go migrations applied).

use axum::{
    Router,
    body::Body,
    http::{Method, Request, StatusCode, header},
};
use http_body_util::BodyExt;
use nexpos_http::{AppState, platform::ensure_admin};
use serde_json::{Value, json};
use sqlx::{
    PgPool,
    postgres::{PgConnectOptions, PgPoolOptions},
};
use tower::ServiceExt;

const ADMIN_EMAIL: &str = "admin@platform.test";
const ADMIN_PASSWORD: &str = "admin-password-for-tests";
const SUPPORT_PASSWORD: &str = "support-password-for-tests";

async fn migrate(db: &PgPool) {
    let dir = concat!(
        env!("CARGO_MANIFEST_DIR"),
        "/../../../backend-go/migrations"
    );
    let mut files: Vec<_> = std::fs::read_dir(dir)
        .expect("migrations directory")
        .map(|entry| entry.unwrap().path())
        .filter(|path| path.to_string_lossy().ends_with(".up.sql"))
        .collect();
    files.sort();
    for file in files {
        let sql = std::fs::read_to_string(&file).unwrap();
        sqlx::raw_sql(&sql)
            .execute(db)
            .await
            .unwrap_or_else(|e| panic!("{}: {e}", file.display()));
    }
}

async fn call(
    app: &Router,
    method: Method,
    path: &str,
    token: &str,
    body: Option<Value>,
) -> (StatusCode, Value) {
    let mut req = Request::builder().method(method).uri(path);
    if !token.is_empty() {
        req = req.header(header::AUTHORIZATION, format!("Bearer {token}"));
    }
    let req = match body {
        Some(body) => req
            .header(header::CONTENT_TYPE, "application/json")
            .body(Body::from(body.to_string())),
        None => req.body(Body::empty()),
    }
    .unwrap();
    let res = app.clone().oneshot(req).await.unwrap();
    let status = res.status();
    let bytes = res.into_body().collect().await.unwrap().to_bytes();
    let json = serde_json::from_slice(&bytes).unwrap_or(Value::Null);
    (status, json)
}

async fn get(app: &Router, path: &str, token: &str) -> (StatusCode, Value) {
    call(app, Method::GET, path, token, None).await
}

async fn post(app: &Router, path: &str, token: &str, body: Value) -> (StatusCode, Value) {
    call(app, Method::POST, path, token, Some(body)).await
}

/// Registers a tenant and returns its access token and company id.
async fn register(app: &Router, name: &str, email: &str) -> (String, i64) {
    let (status, body) = post(
        app,
        "/api/v1/auth/register",
        "",
        json!({ "name": name, "email": email, "password": "tenant-password" }),
    )
    .await;
    assert_eq!(status, StatusCode::CREATED, "{body}");
    (
        body["data"]["access_token"].as_str().unwrap().to_owned(),
        body["data"]["user"]["company"]["id"].as_i64().unwrap(),
    )
}

async fn platform_login(app: &Router, email: &str, password: &str) -> (StatusCode, Value) {
    post(
        app,
        "/api/platform/v1/auth/login",
        "",
        json!({ "email": email, "password": password }),
    )
    .await
}

fn token_of(body: &Value) -> String {
    body["data"]["access_token"].as_str().unwrap().to_owned()
}

fn ids(body: &Value) -> Vec<i64> {
    body["data"]
        .as_array()
        .unwrap()
        .iter()
        .map(|row| row["id"].as_i64().unwrap())
        .collect()
}

fn actions(body: &Value) -> Vec<String> {
    body["data"]
        .as_array()
        .unwrap()
        .iter()
        .map(|row| row["action"].as_str().unwrap().to_owned())
        .collect()
}

async fn scenario(db: PgPool) {
    let app = nexpos_http::app(AppState::new(db.clone()));
    let p = "/api/platform/v1";

    // The first admin is created once.
    assert!(
        ensure_admin(&db, ADMIN_EMAIL, ADMIN_PASSWORD)
            .await
            .unwrap()
    );
    assert!(
        !ensure_admin(&db, ADMIN_EMAIL, ADMIN_PASSWORD)
            .await
            .unwrap()
    );

    let (tenant_a, company_a) = register(&app, "Toko A", "a@tenant.test").await;
    let (tenant_b, company_b) = register(&app, "Toko B", "b@tenant.test").await;

    // ---- Signing in, and the two kinds of token staying apart ----
    let (status, _) = platform_login(&app, ADMIN_EMAIL, "wrong").await;
    assert_eq!(status, StatusCode::UNAUTHORIZED);
    let (status, body) = platform_login(&app, ADMIN_EMAIL, ADMIN_PASSWORD).await;
    assert_eq!(status, StatusCode::OK, "{body}");
    let admin = token_of(&body);
    assert_eq!(body["data"]["user"]["role"], "admin");
    assert!(body["data"]["user"].get("password_hash").is_none());

    assert_eq!(
        get(&app, &format!("{p}/companies"), &tenant_a).await.0,
        StatusCode::UNAUTHORIZED
    );
    assert_eq!(
        get(&app, &format!("{p}/companies"), "").await.0,
        StatusCode::UNAUTHORIZED
    );
    assert_eq!(
        get(&app, "/api/v1/branches", &admin).await.0,
        StatusCode::UNAUTHORIZED
    );

    // ---- The helicopter view ----
    let (status, body) = get(&app, &format!("{p}/companies"), &admin).await;
    assert_eq!(status, StatusCode::OK, "{body}");
    assert_eq!(ids(&body), vec![company_b, company_a]);
    assert_eq!(body["meta"]["pagination"]["total_records"], 2);
    let row = &body["data"][0];
    for key in [
        "code",
        "name",
        "is_active",
        "created_at",
        "users",
        "branches",
        "orders_30d",
        "revenue_30d",
        "last_order_at",
    ] {
        assert!(row.get(key).is_some(), "missing {key}: {row}");
    }
    assert_eq!(row["users"], 1);
    assert_eq!(row["revenue_30d"], 0);
    let (_, body) = get(&app, &format!("{p}/companies?search=toko%20a"), &admin).await;
    assert_eq!(ids(&body), vec![company_a]);
    let (_, body) = get(&app, &format!("{p}/companies?status=suspended"), &admin).await;
    assert_eq!(ids(&body), Vec::<i64>::new());
    assert_eq!(body["data"], json!([]));

    // ---- A support account limited to company A ----
    let new_staff = json!({
        "email": "support@platform.test", "name": "Support Satu", "password": SUPPORT_PASSWORD,
        "role": "support", "company_ids": [company_a], "reason": "new hire, ticket #1",
    });
    let mut no_reason = new_staff.clone();
    no_reason["reason"] = json!("");
    let (status, body) = post(&app, &format!("{p}/staff"), &admin, no_reason).await;
    assert_eq!(status, StatusCode::UNPROCESSABLE_ENTITY, "{body}");
    assert_eq!(body["errors"][0]["field"], "reason");
    let mut bad_company = new_staff.clone();
    bad_company["company_ids"] = json!([999_999]);
    assert_eq!(
        post(&app, &format!("{p}/staff"), &admin, bad_company)
            .await
            .0,
        StatusCode::BAD_REQUEST
    );

    let (status, body) = post(&app, &format!("{p}/staff"), &admin, new_staff.clone()).await;
    assert_eq!(status, StatusCode::CREATED, "{body}");
    let support_id = body["data"]["id"].as_i64().unwrap();
    assert_eq!(body["data"]["company_ids"], json!([company_a]));
    assert_eq!(body["data"]["all_companies"], false);
    assert_eq!(
        post(&app, &format!("{p}/staff"), &admin, new_staff).await.0,
        StatusCode::CONFLICT
    );

    let (status, body) = platform_login(&app, "Support@Platform.test", SUPPORT_PASSWORD).await;
    assert_eq!(status, StatusCode::OK, "{body}");
    let support = token_of(&body);

    // ---- Scoping: company B doesn't exist for this account ----
    let (_, body) = get(&app, &format!("{p}/companies"), &support).await;
    assert_eq!(ids(&body), vec![company_a]);
    assert_eq!(
        get(&app, &format!("{p}/companies/{company_a}"), &support)
            .await
            .0,
        StatusCode::OK
    );
    assert_eq!(
        get(&app, &format!("{p}/companies/{company_b}"), &support)
            .await
            .0,
        StatusCode::NOT_FOUND
    );
    let reason = json!({ "reason": "customer asked, ticket #2" });
    let (status, _) = post(
        &app,
        &format!("{p}/companies/{company_b}/support-session"),
        &support,
        reason.clone(),
    )
    .await;
    assert_eq!(status, StatusCode::NOT_FOUND);

    // ---- Roles: support looks, it doesn't suspend or manage staff ----
    let (status, _) = post(
        &app,
        &format!("{p}/companies/{company_a}/suspend"),
        &support,
        reason.clone(),
    )
    .await;
    assert_eq!(status, StatusCode::FORBIDDEN);
    assert_eq!(
        get(&app, &format!("{p}/staff"), &support).await.0,
        StatusCode::FORBIDDEN
    );

    // ---- A support session sees what the customer sees, and changes nothing ----
    let (status, _) = post(
        &app,
        &format!("{p}/companies/{company_a}/support-session"),
        &support,
        json!({}),
    )
    .await;
    assert_eq!(status, StatusCode::UNPROCESSABLE_ENTITY);
    let (status, body) = post(
        &app,
        &format!("{p}/companies/{company_a}/support-session"),
        &support,
        reason.clone(),
    )
    .await;
    assert_eq!(status, StatusCode::CREATED, "{body}");
    let session = token_of(&body);
    assert_eq!(body["data"]["user"]["email"], "a@tenant.test");
    assert_eq!(body["data"]["expires_in"], 1800);

    let (status, body) = get(&app, "/api/v1/branches", &session).await;
    assert_eq!(status, StatusCode::OK, "{body}");
    assert_eq!(body["data"].as_array().unwrap().len(), 1);
    let (status, body) = post(
        &app,
        "/api/v1/branches",
        &session,
        json!({ "code": "X", "name": "Baru" }),
    )
    .await;
    assert_eq!(status, StatusCode::FORBIDDEN, "{body}");
    assert_eq!(body["error_code"], "READ_ONLY_SESSION");
    let (status, _) = call(&app, Method::DELETE, "/api/v1/customers/1", &session, None).await;
    assert_eq!(status, StatusCode::FORBIDDEN);
    // A preview only computes, so it is allowed through (and fails on its own terms).
    let (status, _) = post(&app, "/api/v1/orders/preview", &session, json!({})).await;
    assert_eq!(status, StatusCode::UNPROCESSABLE_ENTITY);
    // The customer's own session is still read-write.
    let (status, _) = post(
        &app,
        "/api/v1/branches",
        &tenant_a,
        json!({ "code": "X", "name": "Baru" }),
    )
    .await;
    assert_eq!(status, StatusCode::CREATED);

    // ---- Suspending a company ----
    let (status, _) = post(
        &app,
        &format!("{p}/companies/{company_a}/suspend"),
        &admin,
        json!({ "reason": "no" }),
    )
    .await;
    assert_eq!(status, StatusCode::UNPROCESSABLE_ENTITY);
    let (status, body) = post(
        &app,
        &format!("{p}/companies/{company_a}/suspend"),
        &admin,
        json!({ "reason": "unpaid since October" }),
    )
    .await;
    assert_eq!(status, StatusCode::OK, "{body}");
    assert_eq!(body["data"]["is_active"], false);
    let (status, _) = post(
        &app,
        &format!("{p}/companies/{company_a}/suspend"),
        &admin,
        reason.clone(),
    )
    .await;
    assert_eq!(status, StatusCode::BAD_REQUEST);

    for token in [&tenant_a, &session] {
        let (status, body) = get(&app, "/api/v1/branches", token).await;
        assert_eq!(status, StatusCode::FORBIDDEN, "{body}");
        assert_eq!(body["error_code"], "COMPANY_SUSPENDED");
    }
    let (status, body) = post(
        &app,
        "/api/v1/auth/login",
        "",
        json!({ "email": "a@tenant.test", "password": "tenant-password" }),
    )
    .await;
    assert_eq!(status, StatusCode::FORBIDDEN, "{body}");
    assert_eq!(body["error_code"], "COMPANY_SUSPENDED");
    assert_eq!(
        get(&app, "/api/v1/branches", &tenant_b).await.0,
        StatusCode::OK
    );
    let (status, _) = post(
        &app,
        &format!("{p}/companies/{company_a}/support-session"),
        &support,
        reason.clone(),
    )
    .await;
    assert_eq!(status, StatusCode::BAD_REQUEST);

    let (status, body) = post(
        &app,
        &format!("{p}/companies/{company_a}/activate"),
        &admin,
        json!({ "reason": "paid, ticket #3" }),
    )
    .await;
    assert_eq!(status, StatusCode::OK, "{body}");
    assert_eq!(body["data"]["is_active"], true);
    assert_eq!(
        get(&app, "/api/v1/branches", &tenant_a).await.0,
        StatusCode::OK
    );

    // ---- Every write left a line, with who, why, and what changed ----
    let (status, body) = get(&app, &format!("{p}/audit-logs"), &admin).await;
    assert_eq!(status, StatusCode::OK, "{body}");
    assert_eq!(
        actions(&body),
        [
            "company.activate",
            "company.suspend",
            "company.support_session",
            "staff.create"
        ]
    );
    let suspend = &body["data"][1];
    assert_eq!(suspend["actor_email"], ADMIN_EMAIL);
    assert_eq!(suspend["reason"], "unpaid since October");
    assert_eq!(suspend["company_id"], company_a);
    assert_eq!(suspend["before"], json!({ "is_active": true }));
    assert_eq!(suspend["after"], json!({ "is_active": false }));
    assert!(body["data"][3]["after"].get("password_hash").is_none());
    let (_, body) = get(
        &app,
        &format!("{p}/audit-logs?action=company.suspend"),
        &admin,
    )
    .await;
    assert_eq!(actions(&body), ["company.suspend"]);
    // Limited staff see their companies' lines only, not staff changes.
    let (_, body) = get(&app, &format!("{p}/audit-logs"), &support).await;
    assert_eq!(
        actions(&body),
        [
            "company.activate",
            "company.suspend",
            "company.support_session"
        ]
    );

    // ---- Staff changes ----
    let own = format!(
        "{p}/staff/{}",
        body_id(&get(&app, &format!("{p}/auth/me"), &admin).await.1)
    );
    let (status, _) = call(
        &app,
        Method::PUT,
        &own,
        &admin,
        Some(json!({ "role": "ops", "reason": "testing demotion" })),
    )
    .await;
    assert_eq!(status, StatusCode::BAD_REQUEST);
    let (status, _) = call(
        &app,
        Method::PUT,
        &own,
        &admin,
        Some(json!({ "is_active": false, "reason": "testing lockout" })),
    )
    .await;
    assert_eq!(status, StatusCode::BAD_REQUEST);

    let support_path = format!("{p}/staff/{support_id}");
    let (status, body) = call(
        &app,
        Method::PUT,
        &support_path,
        &admin,
        Some(json!({ "role": "ops", "all_companies": true, "reason": "promoted to ops" })),
    )
    .await;
    assert_eq!(status, StatusCode::OK, "{body}");
    assert_eq!(body["data"]["role"], "ops");
    assert_eq!(
        body["data"]["company_ids"],
        json!([company_a]),
        "unsent fields stay"
    );
    let (_, body) = get(&app, &format!("{p}/companies"), &support).await;
    assert_eq!(
        ids(&body),
        vec![company_b, company_a],
        "the new scope applies at once"
    );

    // ---- Lockout after repeated wrong passwords ----
    for _ in 0..5 {
        assert_eq!(
            platform_login(&app, "support@platform.test", "wrong")
                .await
                .0,
            StatusCode::UNAUTHORIZED
        );
    }
    assert_eq!(
        platform_login(&app, "support@platform.test", SUPPORT_PASSWORD)
            .await
            .0,
        StatusCode::UNAUTHORIZED
    );

    // ---- Switching an account off signs it out ----
    let (status, _) = call(
        &app,
        Method::PUT,
        &support_path,
        &admin,
        Some(json!({ "is_active": false, "reason": "left the company" })),
    )
    .await;
    assert_eq!(status, StatusCode::OK);
    assert_eq!(
        get(&app, &format!("{p}/companies"), &support).await.0,
        StatusCode::UNAUTHORIZED
    );

    // ---- Signing out ----
    assert_eq!(
        post(&app, &format!("{p}/auth/logout"), &admin, json!({}))
            .await
            .0,
        StatusCode::OK
    );
    assert_eq!(
        get(&app, &format!("{p}/auth/me"), &admin).await.0,
        StatusCode::UNAUTHORIZED
    );
}

fn body_id(body: &Value) -> i64 {
    body["data"]["id"].as_i64().unwrap()
}

#[tokio::test]
async fn platform_panel() {
    let Ok(url) = std::env::var("DATABASE_URL") else {
        eprintln!("skipped: set DATABASE_URL to run the platform panel tests");
        return;
    };
    let name = format!("nexpos_platform_it_{}", std::process::id());
    let server = PgPool::connect(&url).await.expect("connect DATABASE_URL");
    sqlx::query(&format!("CREATE DATABASE {name}"))
        .execute(&server)
        .await
        .expect("create the throwaway database");

    let options: PgConnectOptions = url.parse().unwrap();
    let db = PgPoolOptions::new()
        .connect_with(options.database(&name))
        .await
        .unwrap();
    // In a task, so a failed assertion still reaches the cleanup below.
    let outcome = tokio::spawn({
        let db = db.clone();
        async move {
            migrate(&db).await;
            scenario(db).await;
        }
    })
    .await;

    db.close().await;
    sqlx::query(&format!("DROP DATABASE IF EXISTS {name} WITH (FORCE)"))
        .execute(&server)
        .await
        .expect("drop the throwaway database");
    if let Err(err) = outcome {
        std::panic::resume_unwind(err.into_panic());
    }
}
