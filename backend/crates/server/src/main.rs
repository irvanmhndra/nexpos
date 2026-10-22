//! nexpos Rust API server: config, database pool, tracing, graceful shutdown.

use std::{env, str::FromStr, time::Duration};

use anyhow::Context;
use nexpos_http::AppState;
use sqlx::postgres::{PgConnectOptions, PgPoolOptions, PgSslMode};
use tokio::net::TcpListener;
use tracing_subscriber::EnvFilter;

/// Reads the same `POSTGRES_*` variables as the Go API, so both can share one
/// `.env` on the server. `SERVER_PORT` defaults to 8090 (Go uses 8080).
struct Config {
    port: u16,
    production: bool,
    db: PgConnectOptions,
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

        Ok(Self {
            port,
            production: var("SERVER_ENV", "development") == "production",
            db,
        })
    }
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let config = Config::from_env()?;

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

    axum::serve(listener, nexpos_http::app(AppState { db: db.clone() }))
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
