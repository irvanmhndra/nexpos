//! Session tokens, password hashes, and company codes, interchangeable with the
//! Go API (`pkg/sessiontoken`, `golang.org/x/crypto/bcrypt`, `generateCompanyCode`),
//! so sessions and accounts work on either backend.

use rand::RngCore;
use sha2::{Digest, Sha256};

/// Go's `bcrypt.DefaultCost`.
const BCRYPT_COST: u32 = 10;

/// A random 256-bit token, hex encoded. Handed to the client once, never stored.
pub fn new_token() -> String {
    let mut bytes = [0u8; 32];
    rand::rng().fill_bytes(&mut bytes);
    hex::encode(bytes)
}

/// Hex SHA-256 of a token: the only form the database stores. Must match Go's
/// `sessiontoken.Hash` and migration 000036.
pub fn hash_token(token: &str) -> String {
    hex::encode(Sha256::digest(token.as_bytes()))
}

/// A bcrypt hash in Go's `$2a$` format, so the Go API can verify it too.
pub fn hash_password(password: &str) -> Result<String, bcrypt::BcryptError> {
    Ok(bcrypt::hash_with_result(password, BCRYPT_COST)?.format_for_version(bcrypt::Version::TwoA))
}

/// Checks a password against a bcrypt hash made by either backend. A malformed
/// hash counts as a mismatch, like Go's `CompareHashAndPassword`.
pub fn verify_password(password: &str, hash: &str) -> bool {
    bcrypt::verify(password, hash).unwrap_or(false)
}

/// The code of a newly registered company: the name upper-cased without spaces,
/// cut to 5 characters, plus `-` and 4 random hex digits (e.g. `BUDIS-3F9A`).
pub fn company_code(name: &str) -> String {
    let mut suffix = [0u8; 2];
    rand::rng().fill_bytes(&mut suffix);
    company_code_with_suffix(name, &hex::encode_upper(suffix))
}

fn company_code_with_suffix(name: &str, suffix: &str) -> String {
    let prefix: String = name
        .to_uppercase()
        .replace(' ', "")
        .chars()
        .take(5)
        .collect();
    format!("{prefix}-{suffix}")
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn token_is_64_hex_chars_and_unique() {
        let (a, b) = (new_token(), new_token());
        assert_eq!(a.len(), 64);
        assert!(a.chars().all(|c| c.is_ascii_hexdigit()));
        assert_ne!(a, b);
    }

    #[test]
    fn token_hash_matches_go_and_postgres() {
        // sessiontoken.Hash("abc") in Go, and encode(sha256('abc'), 'hex') in Postgres.
        assert_eq!(
            hash_token("abc"),
            "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
        );
    }

    #[test]
    fn password_hash_is_go_compatible() {
        let hash = hash_password("password123").unwrap();
        assert!(hash.starts_with("$2a$10$"), "{hash}");
        assert!(verify_password("password123", &hash));
        assert!(!verify_password("wrong", &hash));
    }

    #[test]
    fn verifies_a_hash_made_by_go() {
        // bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
        let go_hash = "$2a$10$.VCFxL/BbudIXRfUwIl7ju3Tl7Kv6xl66pfUovu0JmSSwuIT/rZxK";
        assert!(verify_password("password123", go_hash));
        assert!(!verify_password("password124", go_hash));
        assert!(!verify_password("password123", "not-a-hash"));
    }

    #[test]
    fn company_code_follows_go() {
        assert_eq!(
            company_code_with_suffix("Budi Santoso", "3F9A"),
            "BUDIS-3F9A"
        );
        assert_eq!(company_code_with_suffix("Al", "0001"), "AL-0001");
        let code = company_code("Test User");
        assert!(code.starts_with("TESTU-") && code.len() == 10, "{code}");
    }
}
