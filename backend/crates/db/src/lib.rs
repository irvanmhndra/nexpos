//! SQL for the nexpos Rust API. The schema is owned by the Go API's migrations
//! (`backend-go/migrations`); these queries read and write it as-is.

pub mod auth;
pub mod branches;
pub mod categories;
pub mod company_settings;
pub mod customers;
pub mod expenses;
pub mod products;
pub mod promotions;
pub mod reports;
pub mod suppliers;
pub mod users;
