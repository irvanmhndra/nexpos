//! SQL for the nexpos Rust API. The schema is owned by the Go API's migrations
//! (`backend-go/migrations`); these queries read and write it as-is.

pub mod auth;
pub mod branches;
pub mod customers;
