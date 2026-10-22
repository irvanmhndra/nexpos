use sqlx::PgPool;

/// Shared by every handler. Cheap to clone: the pool is reference-counted.
#[derive(Clone)]
pub struct AppState {
    pub db: PgPool,
}
