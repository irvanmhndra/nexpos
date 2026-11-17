#!/usr/bin/env bash
# Daily database backup for the VPS: pg_dump from the running container, gzip,
# keep locally for a week, and copy to R2 when rclone is configured.
#
# Install once on the server (as the deploy user):
#   scp ops/backup-db.sh <server>:~/nexpos/backup.sh && ssh <server> chmod +x ~/nexpos/backup.sh
#   crontab -e   →   15 19 * * * ~/nexpos/backup.sh >> ~/nexpos/backup.log 2>&1   # 02:15 WIB
#
# Off-site copy (recommended): `rclone config` a remote named "r2" for the R2
# account, then set BACKUP_REMOTE=r2:<bucket>/nexpos in ~/nexpos/.env.
#
# Restore into a scratch database to test a backup (do this once a month):
#   gunzip -c <file> | docker exec -i nexpos_db psql -U "$POSTGRES_USER" -d <scratch_db>
set -euo pipefail

dir="${BACKUP_DIR:-$HOME/nexpos/backups}"
container="${BACKUP_CONTAINER:-nexpos_db}"
keep_days="${BACKUP_KEEP_DAYS:-7}"

# Only the names it needs from the app's env file; nothing is printed.
set -a
# shellcheck disable=SC1091
. "$HOME/nexpos/.env"
set +a

mkdir -p "$dir"
file="$dir/${POSTGRES_DB}_$(date -u +%Y%m%dT%H%M%SZ).sql.gz"
tmp="$file.partial"

docker exec "$container" pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --no-owner --clean --if-exists | gzip > "$tmp"

# An empty or truncated dump must not replace a good one silently.
if ! gunzip -t "$tmp" || [ "$(gunzip -c "$tmp" | tail -5 | grep -c 'PostgreSQL database dump complete')" -ne 1 ]; then
  rm -f "$tmp"
  echo "$(date -u +%FT%TZ) backup FAILED: dump is incomplete" >&2
  exit 1
fi
mv "$tmp" "$file"
echo "$(date -u +%FT%TZ) backup ok: $file ($(du -h "$file" | cut -f1))"

if [ -n "${BACKUP_REMOTE:-}" ]; then
  rclone copy "$file" "$BACKUP_REMOTE" && echo "$(date -u +%FT%TZ) copied to $BACKUP_REMOTE"
fi

find "$dir" -name '*.sql.gz' -mtime +"$keep_days" -delete
