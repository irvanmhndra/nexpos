//! Timestamps rendered like Go's `time.Time` JSON (RFC 3339, trailing zeros of
//! the fraction dropped). lib/pq hands Go its times in the database's TimeZone,
//! while sqlx always reads UTC, so the zone is configured once at startup
//! (`DB_TIMEZONE`, default UTC like the production database).

use std::sync::OnceLock;

use chrono::{DateTime, Offset, Timelike, Utc};
use chrono_tz::Tz;
use serde::Serializer;

static ZONE: OnceLock<Tz> = OnceLock::new();

/// Sets the zone timestamps are shown in. Only the first call has an effect.
pub fn set_zone(zone: Tz) {
    let _ = ZONE.set(zone);
}

fn zone() -> Tz {
    *ZONE.get().unwrap_or(&Tz::UTC)
}

/// `2026-10-23T19:47:14.198694+07:00`, or `…Z` at offset zero.
pub fn go_rfc3339(at: &DateTime<Utc>, zone: Tz) -> String {
    let local = at.with_timezone(&zone);
    let mut out = local.format("%Y-%m-%dT%H:%M:%S").to_string();
    let nanos = local.nanosecond() % 1_000_000_000;
    if nanos > 0 {
        let frac = format!("{nanos:09}");
        out.push('.');
        out.push_str(frac.trim_end_matches('0'));
    }
    let offset = local.offset().fix().local_minus_utc();
    if offset == 0 {
        out.push('Z');
    } else {
        let sign = if offset < 0 { '-' } else { '+' };
        let abs = offset.abs();
        out.push_str(&format!("{sign}{:02}:{:02}", abs / 3600, abs % 3600 / 60));
    }
    out
}

/// `#[serde(serialize_with = "crate::time::serialize")]`
pub fn serialize<S: Serializer>(at: &DateTime<Utc>, s: S) -> Result<S::Ok, S::Error> {
    s.serialize_str(&go_rfc3339(at, zone()))
}

#[cfg(test)]
mod tests {
    use chrono::TimeZone;

    use super::*;

    #[test]
    fn matches_go_marshal_json() {
        let at = Utc.with_ymd_and_hms(2026, 10, 23, 12, 47, 14).unwrap()
            + chrono::Duration::microseconds(198_690);
        assert_eq!(go_rfc3339(&at, Tz::UTC), "2026-10-23T12:47:14.19869Z");
        assert_eq!(
            go_rfc3339(&at, chrono_tz::Asia::Jakarta),
            "2026-10-23T19:47:14.19869+07:00"
        );
        let whole = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap();
        assert_eq!(go_rfc3339(&whole, Tz::UTC), "2026-01-01T00:00:00Z");
        assert_eq!(
            go_rfc3339(&whole, chrono_tz::America::Sao_Paulo),
            "2025-12-31T21:00:00-03:00"
        );
    }
}
