#!/usr/bin/env bash
set -euo pipefail
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
file="${1:?usage: scripts/restore-production.sh backup.sql.gz}"
test -f "$file"
gzip -t "$file"
echo "Refusing automatic restore into production."
echo "Use this only in a disposable restore-test database/container after reviewing the backup."
exit 2
