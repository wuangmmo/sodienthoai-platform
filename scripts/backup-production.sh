#!/usr/bin/env bash
set -euo pipefail
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"
mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_DIR"
ts="$(date -u +%Y%m%dT%H%M%SZ)"
file="$BACKUP_DIR/sodienthoai-$ts.sql.gz"
docker compose -f docker-compose.production.yml exec -T postgres pg_dump -U sodienthoai -d sodienthoai --no-owner --no-privileges | gzip -9 > "$file"
gzip -t "$file"
test -s "$file"
chmod 600 "$file"
printf 'backup_ok %s\n' "$file"
