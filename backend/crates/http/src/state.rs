use std::time::Duration;

use sqlx::PgPool;

/// Shared by every handler. Cheap to clone: the pool and client are reference-counted.
#[derive(Clone)]
pub struct AppState {
    pub db: PgPool,
    pub auth: AuthConfig,
    /// The Go API. Routes Rust doesn't serve yet are forwarded there (strangler).
    pub upstream: Option<Upstream>,
}

impl AppState {
    pub fn new(db: PgPool) -> Self {
        Self {
            db,
            auth: AuthConfig::default(),
            upstream: None,
        }
    }

    pub fn with_auth(mut self, auth: AuthConfig) -> Self {
        self.auth = auth;
        self
    }

    pub fn with_upstream(mut self, base_url: &str) -> Self {
        self.upstream = Some(Upstream {
            base_url: base_url.trim_end_matches('/').to_owned(),
            client: reqwest::Client::new(),
        });
        self
    }
}

/// Session lifetimes. Same env and defaults as the Go API
/// (`JWT_ACCESS_EXPIRES_HOURS`=2, `JWT_REFRESH_EXPIRES_DAYS`=7).
#[derive(Clone, Copy, Debug)]
pub struct AuthConfig {
    pub access_ttl: Duration,
    pub refresh_ttl: Duration,
}

impl Default for AuthConfig {
    fn default() -> Self {
        Self {
            access_ttl: Duration::from_secs(2 * 3600),
            refresh_ttl: Duration::from_secs(7 * 24 * 3600),
        }
    }
}

#[derive(Clone)]
pub struct Upstream {
    pub base_url: String,
    pub client: reqwest::Client,
}
