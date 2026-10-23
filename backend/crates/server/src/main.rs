//! nexpos Rust API server: config, database pool, tracing, graceful shutdown.

use std::{env, net::SocketAddr, str::FromStr, time::Duration};

use anyhow::Context;
use nexpos_http::{AppState, AuthConfig};
use sqlx::postgres::{PgConnectOptions, PgPoolOptions, PgSslMode};
use tokio::net::TcpListener;
use tracing_subscriber::EnvFilter;

/// Reads the same `POSTGRES_*` variables as the Go API, so both can share one
/// `.env` on the server. `SERVER_PORT` defaults to 8090 (Go uses 8080).
struct Config {
    port: u16,
    production: bool,
    db: PgConnectOptions,
    auth: AuthConfig,
    /// The Go API that serves every route not ported yet (`GO_UPSTREAM_URL`).
    upstream: Option<String>,
    /// The database's TimeZone (`DB_TIMEZONE`, IANA name, default UTC). Go shows
    /// timestamps in it, so Rust does too.
    timezone: chrono_tz::Tz,
}

impl Config {
    fn from_env() -> anyhow::Result<Self> {
        let var = |key: &str, default: &str| {
            env::var(key)
                .ok()
                .filter(|v| !v.is_empty())
                .unwrap_or(default.into())
        };

        let port = var("SERVER_PORT", "8090")
            .parse()
            .context("SERVER_PORT must be a port number")?;
        let db = PgConnectOptions::new()
            .host(&var("POSTGRES_HOST", "localhost"))
            .port(
                var("POSTGRES_PORT", "5432")
                    .parse()
                    .context("POSTGRES_PORT must be a port number")?,
            )
            .username(&var("POSTGRES_USER", "pos_user"))
            .password(&var("POSTGRES_PASSWORD", ""))
            .database(&var("POSTGRES_DB", "pos_db"))
            .ssl_mode(
                PgSslMode::from_str(&var("POSTGRES_SSLMODE", "disable"))
                    .context("invalid POSTGRES_SSLMODE")?,
            );

        let hours = |key: &str, default: &str| -> anyhow::Result<Duration> {
            let n: u64 = var(key, default)
                .parse()
                .with_context(|| format!("{key} must be a number"))?;
            Ok(Duration::from_secs(n * 3600))
        };
        let auth = AuthConfig {
            access_ttl: hours("JWT_ACCESS_EXPIRES_HOURS", "2")?,
            refresh_ttl: hours("JWT_REFRESH_EXPIRES_DAYS", "7")? * 24,
        };

        Ok(Self {
            port,
            production: var("SERVER_ENV", "development") == "production",
            db,
            auth,
            upstream: env::var("GO_UPSTREAM_URL").ok().filter(|v| !v.is_empty()),
            timezone: var("DB_TIMEZONE", "UTC")
                .parse()
                .map_err(|e| anyhow::anyhow!("invalid DB_TIMEZONE: {e}"))?,
        })
    }
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let config = Config::from_env()?;
    nexpos_http::time::set_zone(config.timezone);

    let filter = EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| EnvFilter::new("info,tower_http=info"));
    if config.production {
        tracing_subscriber::fmt()
            .json()
            .with_env_filter(filter)
            .init();
    } else {
        tracing_subscriber::fmt().with_env_filter(filter).init();
    }

    // Lazy: the server starts even while Postgres is down, and /health reports
    // it, instead of crash-looping like the Go API does on a failed connect.
    let db = PgPoolOptions::new()
        .max_connections(25)
        .acquire_timeout(Duration::from_secs(5))
        .max_lifetime(Duration::from_secs(5 * 60))
        .connect_lazy_with(config.db);

    let listener = TcpListener::bind(("0.0.0.0", config.port)).await?;
    tracing::info!(port = config.port, "nexpos-server listening");

    let mut state = AppState::new(db.clone()).with_auth(config.auth);
    if let Some(upstream) = &config.upstream {
        tracing::info!(%upstream, "forwarding routes not ported yet to the Go API");
        state = state.with_upstream(upstream);
    }

    // Connect info gives handlers the peer address, stored on sessions like Go does.
    let app = nexpos_http::app(state).into_make_service_with_connect_info::<SocketAddr>();
    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await?;

    db.close().await;
    tracing::info!("server exited gracefully");
    Ok(())
}

async fn shutdown_signal() {
    let ctrl_c = async {
        tokio::signal::ctrl_c().await.ok();
    };
    #[cfg(unix)]
    let terminate = async {
        if let Ok(mut s) = tokio::signal::unix::signal(tokio::signal::unix::SignalKind::terminate())
        {
            s.recv().await;
        }
    };
    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        () = ctrl_c => {},
        () = terminate => {},
    }
    tracing::info!("shutting down");
}
