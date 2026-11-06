#!/usr/bin/env bash
# Runs the Go integration suite against the Rust API on a throwaway database:
# Go on :18080 (serves the routes not ported yet), Rust on :18090 in front of it.
# Cleans up after itself.
#
#   backend/scripts/conformance.sh                  # tests in conformance-tests.txt
#   backend/scripts/conformance.sh '^TestBranch'    # a specific set
#
# Needs a local Postgres. Override with PGHOST, PGPORT, PGUSER, PGPASSWORD.
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
tests="${1:-$(paste -sd'|' "$root/backend/conformance-tests.txt")}"
export PGHOST="${PGHOST:-127.0.0.1}" PGPORT="${PGPORT:-5432}" PGUSER="${PGUSER:-$(whoami)}" PGPASSWORD="${PGPASSWORD:-}"
db="nexpos_conformance_$$"
work="$(mktemp -d)"
pids=()

cleanup() {
  for pid in "${pids[@]:-}"; do [ -n "$pid" ] && kill "$pid" 2>/dev/null || true; done
  wait 2>/dev/null || true
  psql -q -d postgres -c "DROP DATABASE IF EXISTS $db WITH (FORCE)" >/dev/null 2>&1 || true
  rm -rf "$work"
}
trap cleanup EXIT

wait_for() {
  for _ in $(seq 1 60); do
    curl -sf "$1" >/dev/null && return 0
    sleep 0.5
  done
  echo "server did not start: $1" >&2
  cat "$work"/*.log >&2
  return 1
}

unset GOROOT
psql -q -d postgres -c "CREATE DATABASE $db"
# sqlx always reads UTC, so tell Rust which zone Go will show timestamps in.
zone="$(psql -At -d "$db" -c 'SHOW timezone')"

(cd "$root/backend-go" && go build -o "$work/nexpos-go" ./cmd/api)
(cd "$root/backend" && cargo build -q --release --bin nexpos-server)

export POSTGRES_HOST="$PGHOST" POSTGRES_PORT="$PGPORT" POSTGRES_USER="$PGUSER" POSTGRES_PASSWORD="$PGPASSWORD" POSTGRES_DB="$db" POSTGRES_SSLMODE=disable

(cd "$root/backend-go" && SERVER_PORT=18080 SERVER_ENV=test JWT_SECRET=conformance-secret exec "$work/nexpos-go") >"$work/go.log" 2>&1 &
pids+=($!)
SERVER_PORT=18090 DB_TIMEZONE="$zone" GO_UPSTREAM_URL=http://127.0.0.1:18080 "$root/backend/target/release/nexpos-server" >"$work/rust.log" 2>&1 &
pids+=($!)
wait_for http://127.0.0.1:18080/health/live
wait_for http://127.0.0.1:18090/health/live

cd "$root/backend-go"
TARGET_BASE_URL=http://127.0.0.1:18090 \
COMPARE_BASE_URL=http://127.0.0.1:18080 \
TEST_DATABASE_URL="postgres://$PGUSER${PGPASSWORD:+:$PGPASSWORD}@$PGHOST:$PGPORT/$db?sslmode=disable" \
  go test -count=1 -p 1 -v -run "$tests" ./tests/integration/ 2>&1 |
  grep -E '^(--- |ok|FAIL|panic)|Error:|expected:|actual  :|Messages:|Error Trace:' | cut -c1-260
status="${PIPESTATUS[0]}"
if [ "$status" -ne 0 ]; then
  echo "--- rust errors ---"
  grep -a ERROR "$work/rust.log" | tail -20 | cut -c1-300 || true
fi
exit "$status"
